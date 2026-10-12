package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"mihan/internal/eval"
)

//go:embed preview.html
var previewPage string

var previewDevices = map[string][2]int{
	"pixel":  {412, 915},
	"small":  {360, 640},
	"tablet": {800, 1280},
}

type previewEvent struct {
	N    int    `json:"n"`
	Kind string `json:"kind"`
	Text string `json:"text"`
	At   string `json:"at"`
}

type preview struct {
	path, device string
	self         string

	mu      sync.Mutex
	events  []previewEvent
	subs    map[chan previewEvent]bool
	proc    *exec.Cmd
	urlFile string
	url     string
	output  strings.Builder
	running bool
	gen     int
	lastSub time.Time
}

const previewUsage = "mihan preview app.mhn --android [--device pixel|small|tablet] [--no-open] [--port n]"

func cmdPreview(args []string) error {
	p := &preview{device: "pixel", subs: map[chan previewEvent]bool{}}
	android, open, port := false, true, 0
	for i := 0; i < len(args); i++ {
		switch a := args[i]; {
		case a == "--android":
			android = true
		case a == "--device" && i+1 < len(args):
			i++
			p.device = args[i]
		case a == "--no-open":
			open = false
		case a == "--port" && i+1 < len(args):
			i++
			fmt.Sscan(args[i], &port)
		case strings.HasPrefix(a, "-"):
			return fatal("I do not know the option %s\n  usage: %s", a, previewUsage)
		default:
			p.path = a
		}
	}
	if p.path == "" {
		return fatal("give me the program to preview\n  usage: %s", previewUsage)
	}
	if !android {
		return fatal("say which kind of preview: --android (the only one there is so far)\n  usage: %s", previewUsage)
	}
	if _, ok := previewDevices[p.device]; !ok {
		return fatal("--device must be pixel, small or tablet, not %s", p.device)
	}
	if info, err := os.Stat(p.path); err != nil {
		return fatal("%s is not here", p.path)
	} else if info.IsDir() {
		p.path = filepath.Join(p.path, "main.mhn")
	}
	if err := checkOnly(p.path); err != nil {
		return err
	}
	self, err := os.Executable()
	if err != nil {
		return fatal("cannot find my own executable: %v", err)
	}
	p.self = self
	work, err := os.MkdirTemp("", "mihan-preview-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	p.urlFile = filepath.Join(work, "mihan-window.txt")

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fatal("cannot listen: %v", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.servePage)
	mux.HandleFunc("/_preview/events", p.serveEvents)
	mux.HandleFunc("/_preview/state", p.serveState)
	mux.HandleFunc("/_preview/do", p.serveDo)
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	addr := "http://" + ln.Addr().String() + "/"

	p.start(work)
	fmt.Printf("preview of %s as an Android app (%s) at %s\n", p.path, p.device, addr)
	fmt.Println("this is a faithful preview on this computer, not an Android emulator; Ctrl+C stops it")

	done := make(chan struct{})
	var once sync.Once
	finish := func() { once.Do(func() { close(done) }) }
	if open {
		size := previewDevices[p.device]
		cmd, owned, err := eval.OpenWindow(addr, size[0]+520, min(size[1]+120, 1000))
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not open a window (%v); open %s yourself\n", err, addr)
		} else if owned {
			go func() { cmd.Wait(); finish() }()
		}
	}

	go func() {
		for range time.Tick(time.Second) {
			p.mu.Lock()
			idle := len(p.subs) == 0 && !p.lastSub.IsZero() && time.Since(p.lastSub) > 6*time.Second
			p.mu.Unlock()
			if idle {
				finish()
				return
			}
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	select {
	case <-sig:
	case <-done:
	}
	p.stop()
	srv.Close()
	return nil
}

func (p *preview) emit(kind, text string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	ev := previewEvent{N: len(p.events), Kind: kind, Text: text, At: time.Now().Format("15:04:05.000")}
	p.events = append(p.events, ev)
	if len(p.events) > 5000 {
		p.events = p.events[len(p.events)-4000:]
	}
	for c := range p.subs {
		select {
		case c <- ev:
		default:
		}
	}
}

func (p *preview) start(work string) {
	p.mu.Lock()
	p.gen++
	gen := p.gen
	p.url = ""
	p.output.Reset()
	p.mu.Unlock()
	os.Remove(p.urlFile)
	files := filepath.Join(work, "files")
	os.MkdirAll(files, 0o755)
	abs, _ := filepath.Abs(p.path)
	cmd := exec.Command(p.self, "run", abs)
	cmd.Dir = files
	cmd.Env = append(os.Environ(),
		"MIHAN_ANDROID=1", "MIHAN_WINDOW_FILE="+p.urlFile, "MIHAN_REQUEST_LOG=1",
		"MIHAN_NO_WINDOW=", "MIHAN_PREVIEW=1")
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	p.emit("life", "onCreate: starting "+filepath.Base(p.path))
	if err := cmd.Start(); err != nil {
		p.emit("error", "could not start the program: "+err.Error())
		return
	}
	p.mu.Lock()
	p.proc = cmd
	p.running = true
	p.mu.Unlock()
	var wg sync.WaitGroup
	read := func(r io.Reader, isErr bool) {
		defer wg.Done()
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 64<<10), 1<<20)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "@@mihan-req ") {
				p.emit("req", strings.TrimPrefix(line, "@@mihan-req "))
				continue
			}
			p.mu.Lock()
			if gen == p.gen {
				p.output.WriteString(line + "\n")
			}
			p.mu.Unlock()
			if isErr {
				p.emit("error", line)
			} else {
				p.emit("print", line)
			}
		}
	}
	wg.Add(2)
	go read(stdout, false)
	go read(stderr, true)
	go func() {
		for {
			time.Sleep(150 * time.Millisecond)
			p.mu.Lock()
			alive := p.running && gen == p.gen
			p.mu.Unlock()
			if !alive {
				return
			}
			if b, err := os.ReadFile(p.urlFile); err == nil {
				u := strings.TrimSpace(string(b))
				p.mu.Lock()
				changed := u != "" && u != p.url
				if changed {
					p.url = u
				}
				p.mu.Unlock()
				if changed {
					p.emit("url", u)
				}
			}
		}
	}()
	go func() {
		wg.Wait()
		err := cmd.Wait()
		code := 0
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		}
		p.mu.Lock()
		current := gen == p.gen
		if current {
			p.running = false
		}
		p.mu.Unlock()
		if current {
			p.emit("exit", fmt.Sprintf("the program ended (status %d)", code))
		}
	}()
}

func (p *preview) stop() {
	p.mu.Lock()
	cmd := p.proc
	p.running = false
	p.gen++
	p.mu.Unlock()
	if cmd != nil && cmd.Process != nil {
		cmd.Process.Kill()
	}
}

func (p *preview) servePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	size := previewDevices[p.device]
	page := strings.NewReplacer("{{device}}", p.device, "{{w}}", fmt.Sprint(size[0]), "{{h}}", fmt.Sprint(size[1]),
		"{{name}}", strings.TrimSuffix(filepath.Base(p.path), filepath.Ext(p.path))).Replace(previewPage)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, page)
}

func (p *preview) serveState(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	st := map[string]any{"url": p.url, "output": p.output.String(), "running": p.running, "device": p.device}
	p.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(st)
}

func (p *preview) serveDo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST", http.StatusMethodNotAllowed)
		return
	}
	what := r.URL.Query().Get("what")
	text := r.URL.Query().Get("text")
	switch what {
	case "finish":
		p.emit("life", "onDestroy: the app was closed")
		p.stop()
	case "restart":
		p.stop()
		p.start(filepath.Dir(p.urlFile))
	case "pause", "resume", "rotate", "back", "home":
		p.emit("life", text)
	case "toast", "vibrate", "share", "console", "jserror", "page":
		kind := "bridge"
		if what == "jserror" {
			kind = "error"
		}
		p.emit(kind, what+": "+text)
	default:
		http.Error(w, "unknown", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (p *preview) serveEvents(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	ch := make(chan previewEvent, 256)
	p.mu.Lock()
	past := append([]previewEvent(nil), p.events...)
	p.subs[ch] = true
	p.lastSub = time.Now()
	p.mu.Unlock()
	defer func() {
		p.mu.Lock()
		delete(p.subs, ch)
		p.lastSub = time.Now()
		p.mu.Unlock()
	}()
	send := func(ev previewEvent) {
		b, _ := json.Marshal(ev)
		fmt.Fprintf(w, "data: %s\n\n", b)
	}
	for _, ev := range past {
		send(ev)
	}
	fl.Flush()
	tick := time.NewTicker(15 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case ev := <-ch:
			send(ev)
			fl.Flush()
		case <-tick.C:
			io.WriteString(w, ": .\n\n")
			fl.Flush()
		}
	}
}
