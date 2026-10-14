package eval

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

var requestKeys = []string{"address", "method", "path", "sorok", "matn", "sarlawah"}

type server struct {
	mu sync.Mutex

	calls sync.Mutex
	held  bool

	srv *http.Server

	addr string

	ln net.Listener

	tabe rt.Value

	interp *Interp

	out io.Writer

	stopped chan struct{}
}

var theServer struct {
	mu sync.Mutex
	s  *server
}

func (s *server) url() string {
	if s == nil || s.addr == "" {
		return ""
	}
	return "http://" + hostOf(s.addr)
}

func hostOf(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}

func fnServeStart(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {

	tabe := args[len(args)-1]
	var baftsh ast.Node
	hasAfter := len(args) >= 3
	if hasAfter {
		baftsh = node
		tabe = args[len(args)-1-1]
	}
	port := 8080
	if len(args) >= 2 {
		n, ok := toInt(args[0])
		if !ok {
			return rt.Nil, in.faultAt(node, "E4280",
				"shomare-ye dargah bayad adad-e kamel bashe",
				"the port has to be a whole number, not %s", rt.TypeName(args[0]))
		}
		port = int(n)
	}
	if tabe == nil {
		return rt.Nil, in.faultAt(node, "E4281",
			"server yek funct-e javabgoo lazem dare",
			"a server needs a funct to answer with, and none was given")
	}

	s := &server{tabe: tabe, interp: in, out: in.Out}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.answer)
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return rt.Nil, in.faultAt(node, "E4282",
			"natoonestam roye in dargah goosh bedam: "+err.Error(),
			"could not listen on that port: %v", err)
	}
	s.ln = ln
	s.addr = ln.Addr().String()
	s.srv = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	theServer.mu.Lock()
	if theServer.s != nil && theServer.s.srv != nil {
		theServer.mu.Unlock()
		ln.Close()
		return rt.Nil, in.faultAt(node, "E4283",
			"yek server alan dare kar mikone",
			"a server is already running on port %d", portOf(theServer.s.addr))
	}
	theServer.s = s
	theServer.mu.Unlock()

	go s.srv.Serve(ln)
	announce(s.url())
	if f := os.Getenv("MIHAN_WINDOW_FILE"); f != "" {

		if _, err := os.Stat(f); err != nil {
			_ = os.WriteFile(f, []byte(s.url()), 0o644)
		}
	}
	if hasAfter {

		s.enter()
		_, err := s.interp.callValue(baftsh, args[len(args)-1], nil, nil)
		s.leave()
		if err != nil {
			s.srv.Close()
			return rt.Nil, err
		}
	}

	<-s.done()
	return rt.Nil, nil
}

func portOf(addr string) int {
	_, p, err := net.SplitHostPort(addr)
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(p)
	return n
}

func (s *server) done() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped == nil {
		s.stopped = make(chan struct{})
	}
	return s.stopped
}

func (s *server) signalStop() {
	s.mu.Lock()
	ch := s.stopped
	if ch == nil {
		ch = make(chan struct{})
		s.stopped = ch
	}
	s.mu.Unlock()
	select {
	case <-ch:
	default:
		close(ch)
	}
}

func (s *server) answer(w http.ResponseWriter, r *http.Request) {
	if s.builtinRoute(w, r) {
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
	if err != nil {
		http.Error(w, "the request could not be read", http.StatusBadRequest)
		return
	}
	sorok := rt.NewDict()
	for k, vs := range r.URL.Query() {
		if len(vs) > 0 {
			sorok.Set(rt.Str(k), rt.Str(vs[0]))
		}
	}
	sarlawah := rt.NewDict()
	for k, vs := range r.Header {
		if len(vs) > 0 {
			sarlawah.Set(rt.Str(k), rt.Str(vs[0]))
		}
	}
	req := rt.NewDict()
	req.Set(rt.Str("address"), rt.Str(r.RemoteAddr))
	req.Set(rt.Str("method"), rt.Str(r.Method))
	req.Set(rt.Str("path"), rt.Str(r.URL.Path))
	req.Set("sorok", sorok)
	req.Set(rt.Str("matn"), rt.Str(string(body)))
	req.Set("sarlawah", sarlawah)
	requestExtras(req, r, body)

	started := time.Now()
	s.enter()
	takePendingSession()
	out, callErr := s.interp.callValue(nil, s.tabe, nil, []rt.Value{req})
	sid := takePendingSession()
	s.leave()
	if sid != "" {
		http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: sid, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	}
	if callErr != nil {
		requestLog(r.Method, r.URL.Path, 500, time.Since(started))

		msg := errText(callErr)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "khata: %s", msg)
		return
	}
	status, text, headers := responseParts(out)
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	answerExtras(w, r, out)
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	defer func() { requestLog(r.Method, r.URL.Path, status, time.Since(started)) }()
	if notModified(r, w.Header(), status) {
		status = http.StatusNotModified
		w.WriteHeader(status)
		return
	}
	if os.Getenv("MIHAN_ANDROID") != "" && strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") {
		text = withPhoneScript(text)
	}
	w.WriteHeader(status)
	if r.Method != "HEAD" {
		io.WriteString(w, text)
	}
}

func withPhoneScript(page string) string {
	const tag = `<script src="/_mihan/mihan.js"></script>`
	if strings.Contains(page, "/_mihan/mihan.js") {
		return page
	}
	low := strings.ToLower(page)
	if i := strings.Index(low, "<head>"); i >= 0 {
		return page[:i+6] + tag + page[i+6:]
	}
	if i := strings.Index(low, "<body"); i >= 0 {
		return page[:i] + tag + page[i:]
	}
	return tag + page
}

func (s *server) builtinRoute(w http.ResponseWriter, r *http.Request) bool {
	switch {
	case r.URL.Path == "/_mihan/mihan.js":
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		io.WriteString(w, androidJS)
		return true
	case strings.HasPrefix(r.URL.Path, "/_mihan/rooydad/"):
		serveEvents(w, r, strings.TrimPrefix(r.URL.Path, "/_mihan/rooydad/"))
		return true
	}
	return false
}

func errText(err error) string {
	if p, ok := rt.AsPanic(err); ok {
		return rt.Show(p.Value)
	}
	return err.Error()
}

func responseParts(v rt.Value) (int, string, map[string]string) {
	status := http.StatusOK
	headers := map[string]string{}
	switch t := v.(type) {
	case rt.Str:
		return status, string(t), headers
	case *rt.Dict:
		if raw, ok := t.Get("vaziat"); ok {
			if s, ok := toInt(raw); ok {
				status = int(s)
			}
		}
		body, ok := t.Get("matn")
		if !ok {
			return status, rt.Show(v), headers
		}
		if h, ok := t.Get("sarlawah"); ok {
			if hd, ok := h.(*rt.Dict); ok {
				for _, key := range hd.Keys() {
					k, ok := key.(rt.Str)
					if !ok {
						continue
					}
					if val, ok := hd.Get(key); ok {
						if s, ok := val.(rt.Str); ok {
							headers[string(k)] = string(s)
						}
					}
				}
			}
		}
		if m, ok := body.(rt.Str); ok {
			return status, string(m), headers
		}
		return status, rt.Show(body), headers
	}
	return status, rt.Show(v), headers
}

func fnServePort(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	theServer.mu.Lock()
	s := theServer.s
	theServer.mu.Unlock()
	if s == nil || s.addr == "" {
		return rt.Int(0), nil
	}
	return rt.Int(int64(portOf(s.addr))), nil
}

func fnServeURL(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	theServer.mu.Lock()
	s := theServer.s
	theServer.mu.Unlock()
	return rt.Str(s.url()), nil
}

func fnServeStop(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	theServer.mu.Lock()
	s := theServer.s
	theServer.mu.Unlock()
	if s == nil || s.srv == nil {
		return rt.Nil, nil
	}
	s.srv.Close()
	s.signalStop()
	return rt.Nil, nil
}

func fnOpenWindow(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	if len(args) < 1 {
		return rt.Nil, in.faultAt(node, "E4290",
			"baz kardan-e panjere yek neshani mikhad",
			"opening a window needs an address")
	}
	target, ok := args[0].(rt.Str)
	if !ok {
		return rt.Nil, in.faultAt(node, "E4291",
			"neshani bayad matn bashe",
			"the address has to be text, not %s", rt.TypeName(args[0]))
	}
	addr := string(target)

	if !strings.Contains(addr, "://") {
		full, err := filepath.Abs(addr)
		if err != nil {
			full = addr
		}
		slashed := filepath.ToSlash(full)
		if !strings.HasPrefix(slashed, "/") {
			slashed = "/" + slashed
		}
		addr = "file://" + strings.ReplaceAll(slashed, " ", "%20")
	}
	width, height := 1100, 750
	if len(args) >= 3 {
		if w, ok := toInt(args[1]); ok {
			width = int(w)
		}
		if h, ok := toInt(args[2]); ok {
			height = int(h)
		}
	}
	if f := os.Getenv("MIHAN_WINDOW_FILE"); f != "" {

		if err := os.WriteFile(f, []byte(addr), 0o644); err != nil {
			return rt.Nil, in.faultAt(node, "E4292",
				"natoonestam panjere ro baz konam: "+err.Error(),
				"could not open the window: %v", err)
		}
		return rt.Nil, nil
	}
	if os.Getenv("MIHAN_NO_WINDOW") != "" {

		return rt.Nil, nil
	}
	runner := windowRunner()
	cmd := exec.Command(runner, windowArgsFor(runner, addr, width, height)...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return rt.Nil, in.faultAt(node, "E4292",
			"natoonestam panjere ro baz konam: "+err.Error(),
			"could not open the window: %v", err)
	}

	started := time.Now()
	go func() {
		_ = cmd.Wait()
		if !windowOwned(runner) || time.Since(started) < 3*time.Second {
			return
		}
		if profile := windowProfile(); profile != "" {
			_ = os.RemoveAll(profile)
		}
		os.Exit(0)
	}()
	return rt.Nil, nil
}

func windowRunner() string {
	if b := os.Getenv("MIHAN_BROWSER"); b != "" {
		return b
	}
	if runtime.GOOS != "windows" {

		for _, name := range linuxBrowsers {
			if p, err := exec.LookPath(name); err == nil {
				return p
			}
		}
		return "xdg-open"
	}

	for _, p := range edgePaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	for _, p := range webView2Paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "msedge"
}

var webView2Paths = []string{
	`C:\Program Files (x86)\Microsoft\EdgeWebView\Application\msedgewebview2.exe`,
	`C:\Program Files\Microsoft\EdgeWebView\Application\msedgewebview2.exe`,
}

var edgePaths = []string{
	`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
	`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
}

func windowProfile() string {
	return filepath.Join(os.TempDir(), "mihan-window-"+strconv.Itoa(os.Getpid()))
}

var linuxBrowsers = []string{
	"google-chrome", "google-chrome-stable", "chromium", "chromium-browser",
	"microsoft-edge", "microsoft-edge-stable", "brave-browser", "firefox",
}

func isFirefox(runner string) bool {
	return strings.Contains(strings.ToLower(filepath.Base(runner)), "firefox")
}

func windowOwned(runner string) bool {
	return filepath.Base(runner) != "xdg-open"
}

func windowArgsFor(runner, addr string, width, height int) []string {
	if filepath.Base(runner) == "xdg-open" {
		return []string{addr}
	}
	if isFirefox(runner) {
		profile := windowProfile()
		_ = os.MkdirAll(profile, 0o755)
		return []string{"--new-instance", "--profile", profile,
			"--width", strconv.Itoa(width), "--height", strconv.Itoa(height), addr}
	}
	args := windowArgs(addr, width, height)
	if os.Getenv("MIHAN_WINDOW_NO_SANDBOX") != "" {
		args = append(args, "--no-sandbox")
	}
	return args
}

func windowArgs(addr string, width, height int) []string {

	profile := windowProfile()
	return []string{
		"--app=" + addr,
		"--window-size=" + strconv.Itoa(width) + "," + strconv.Itoa(height),
		"--user-data-dir=" + profile,
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-features=Translate,TranslateUI,OptimizationGuideModelDownloading",
		"--lang=en",
	}
}

func (s *server) enter() {
	s.calls.Lock()
	s.held = true
}

func (s *server) leave() {
	s.held = false
	s.calls.Unlock()
}

func whileWaiting(wait func()) {
	theServer.mu.Lock()
	s := theServer.s
	theServer.mu.Unlock()
	if s == nil || !s.held {
		wait()
		return
	}
	s.leave()
	defer s.enter()
	wait()
}

func OpenWindow(addr string, width, height int) (*exec.Cmd, bool, error) {
	runner := windowRunner()
	cmd := exec.Command(runner, windowArgsFor(runner, addr, width, height)...)
	if err := cmd.Start(); err != nil {
		return nil, false, err
	}
	return cmd, windowOwned(runner), nil
}

func announce(url string) {
	if os.Getenv("MIHAN_NO_WINDOW") != "" || os.Getenv("MIHAN_ANDROID") != "" {
		return
	}
	open := os.Getenv("MIHAN_OPEN_BROWSER") != ""

	if st, err := os.Stderr.Stat(); open || (err == nil && st.Mode()&os.ModeCharDevice != 0) {
		fmt.Fprintf(os.Stderr, "Mihan site is ready: %s  (Ctrl+C to stop)\n", url)
	}
	if open {
		if err := OpenInBrowser(url); err != nil {
			fmt.Fprintf(os.Stderr, "could not open the browser (%v); open %s yourself\n", err, url)
		}
	}
}

func OpenInBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}
