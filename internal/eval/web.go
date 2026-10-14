package eval

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

func requestExtras(req *rt.Dict, r *http.Request, body []byte) {
	cookies := rt.NewDict()
	for _, c := range r.Cookies() {
		if _, has := cookies.Get(rt.Str(c.Name)); !has {
			cookies.Set(rt.Str(c.Name), rt.Str(c.Value))
		}
	}
	req.Set(rt.Str("cookie"), cookies)
	form := rt.NewDict()
	files := []rt.Value{}
	ctype, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch {
	case ctype == "application/x-www-form-urlencoded":
		if q, err := url.ParseQuery(string(body)); err == nil {
			setSorted(form, q)
		}
	case strings.HasPrefix(ctype, "multipart/") && params["boundary"] != "":
		mr := multipart.NewReader(bytes.NewReader(body), params["boundary"])
		fields := url.Values{}
		for {
			p, err := mr.NextPart()
			if err != nil {
				break
			}
			data, _ := io.ReadAll(io.LimitReader(p, 8<<20))
			if p.FileName() == "" {
				fields.Add(p.FormName(), string(data))
				continue
			}
			f := rt.NewDict()
			f.Set(rt.Str("nam"), rt.Str(p.FormName()))
			f.Set(rt.Str("esm"), rt.Str(p.FileName()))
			noe := p.Header.Get("Content-Type")
			if noe == "" {
				noe = "application/octet-stream"
			}
			f.Set(rt.Str("noe"), rt.Str(noe))
			f.Set(rt.Str("andaze"), rt.Int(int64(len(data))))
			f.Set(rt.Str("matn"), rt.Str(data))
			files = append(files, f)
		}
		setSorted(form, fields)
	}
	req.Set(rt.Str("form"), form)
	req.Set(rt.Str("parvande"), rt.NewList(files...))
	req.Set(rt.Str("params"), rt.NewDict())
}

func setSorted(d *rt.Dict, q url.Values) {
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if len(q[k]) > 0 {
			d.Set(rt.Str(k), rt.Str(q[k][0]))
		}
	}
}

func answerExtras(w http.ResponseWriter, r *http.Request, v rt.Value) {
	d, ok := v.(*rt.Dict)
	if !ok {
		return
	}
	if c, ok := d.Get(rt.Str("cookie")); ok {
		if cd, ok := c.(*rt.Dict); ok {
			for _, k := range cd.Keys() {
				name, ok := k.(rt.Str)
				if !ok {
					continue
				}
				val, _ := cd.Get(k)
				ck := &http.Cookie{Name: string(name), Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode}
				switch t := val.(type) {
				case *rt.Dict:
					ck.Value = rt.Show(dictGet(t, "meghdar", rt.Str("")))
					if n, ok := toInt(dictGet(t, "umr", rt.Nil)); ok {
						ck.MaxAge = int(n)
						if n == 0 {
							ck.MaxAge = -1
						}
					}
					if p := dictStr(t, "masir"); p != "" {
						ck.Path = p
					}
					if b, ok := dictGet(t, "httponly", rt.Bool(true)).(rt.Bool); ok {
						ck.HttpOnly = bool(b)
					}
				case rt.Hich, nil:
					ck.MaxAge = -1
				default:
					ck.Value = rt.Show(val)
				}
				http.SetCookie(w, ck)
			}
		}
	}
	if c, ok := d.Get(rt.Str("cors")); ok {
		origin := ""
		switch t := c.(type) {
		case rt.Bool:
			if t {
				origin = "*"
			}
		case rt.Str:
			origin = string(t)
		}
		if origin != "" {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if origin != "*" {
				h.Add("Vary", "Origin")
				h.Set("Access-Control-Allow-Credentials", "true")
			}
		}
	}
}

func dictGet(d *rt.Dict, key string, def rt.Value) rt.Value {
	if v, ok := d.Get(rt.Str(key)); ok {
		return v
	}
	return def
}

func notModified(r *http.Request, h http.Header, status int) bool {
	tag := h.Get("ETag")
	if status != 200 || tag == "" || (r.Method != "GET" && r.Method != "HEAD") {
		return false
	}
	for _, t := range strings.Split(r.Header.Get("If-None-Match"), ",") {
		if strings.TrimSpace(t) == tag || strings.TrimSpace(t) == "*" {
			return true
		}
	}
	return false
}

func fileTag(body []byte) string {
	sum := sha256.Sum256(body)
	return `"` + hex.EncodeToString(sum[:8]) + `"`
}

const sessionCookie = "mihan_jalase"

var sessions struct {
	mu   sync.Mutex
	m    map[string]*rt.Dict
	used map[string]time.Time
}

var pendingSession struct {
	mu sync.Mutex
	id string
}

func fnSession(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	req, ok := args[0].(*rt.Dict)
	if !ok {
		return nil, in.faultAt(node, "E4360",
			"jalase darkhast ro mikhad",
			"jalase needs the request dict, not %s", rt.TypeName(args[0]))
	}
	id := ""
	if c, ok := req.Get(rt.Str("cookie")); ok {
		if cd, ok := c.(*rt.Dict); ok {
			id = dictStr(cd, sessionCookie)
		}
	}
	sessions.mu.Lock()
	defer sessions.mu.Unlock()
	if sessions.m == nil {
		sessions.m = map[string]*rt.Dict{}
		sessions.used = map[string]time.Time{}
	}
	if d, ok := sessions.m[id]; ok && id != "" {
		sessions.used[id] = time.Now()
		return d, nil
	}

	for k, t := range sessions.used {
		if time.Since(t) > 24*time.Hour {
			delete(sessions.m, k)
			delete(sessions.used, k)
		}
	}
	var b [16]byte
	_, _ = rand.Read(b[:])
	id = hex.EncodeToString(b[:])
	d := rt.NewDict()
	sessions.m[id] = d
	sessions.used[id] = time.Now()
	if c, ok := req.Get(rt.Str("cookie")); ok {
		if cd, ok := c.(*rt.Dict); ok {
			cd.Set(rt.Str(sessionCookie), rt.Str(id))
		}
	}
	pendingSession.mu.Lock()
	pendingSession.id = id
	pendingSession.mu.Unlock()
	return d, nil
}

func takePendingSession() string {
	pendingSession.mu.Lock()
	defer pendingSession.mu.Unlock()
	id := pendingSession.id
	pendingSession.id = ""
	return id
}

func fnJSONAnswer(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	text, err := fnJSONWrite(in, args[:1], node)
	if err != nil {
		return nil, err
	}
	status := int64(200)
	if len(args) > 1 {
		n, ok := toInt(args[1])
		if !ok {
			return nil, in.faultAt(node, "E4361",
				"vaziat bayad adad bashe",
				"the status has to be a whole number, not %s", rt.TypeName(args[1]))
		}
		status = int64(n)
	}
	return pageAnswer(int(status), string(text.(rt.Str)), "application/json; charset=utf-8"), nil
}

func fnTemplate(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	src, ok := args[0].(rt.Str)
	data, ok2 := args[1].(*rt.Dict)
	if !ok || !ok2 {
		return nil, in.faultAt(node, "E4362",
			"qaleb yek matn va yek dict mikhad",
			"qaleb needs the template text and a dict of values")
	}
	s := string(src)
	var b strings.Builder
	for {
		i := strings.Index(s, "{{")
		if i < 0 {
			b.WriteString(s)
			break
		}
		b.WriteString(s[:i])
		raw := strings.HasPrefix(s[i:], "{{{")
		open, close := 2, "}}"
		if raw {
			open, close = 3, "}}}"
		}
		j := strings.Index(s[i+open:], close)
		if j < 0 {
			b.WriteString(s[i:])
			break
		}
		name := strings.TrimSpace(s[i+open : i+open+j])
		v, found := lookupPath(data, name)
		if !found {
			return nil, in.faultAt(node, "E4363",
				fmt.Sprintf("qaleb '%s' ro nadare", name),
				"the template asks for %s, which the dict does not have", name)
		}
		if raw {
			b.WriteString(rt.Show(v))
		} else {
			b.WriteString(html.EscapeString(rt.Show(v)))
		}
		s = s[i+open+j+len(close):]
	}
	return rt.Str(b.String()), nil
}

func lookupPath(d *rt.Dict, name string) (rt.Value, bool) {
	var cur rt.Value = d
	for _, part := range strings.Split(name, ".") {
		cd, ok := cur.(*rt.Dict)
		if !ok {
			return nil, false
		}
		if cur, ok = cd.Get(rt.Str(part)); !ok {
			return nil, false
		}
	}
	return cur, true
}

func fnHTMLSafe(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	return rt.Str(html.EscapeString(rt.Show(args[0]))), nil
}

func matchRoute(route, path string, params *rt.Dict) bool {
	if route == path {
		return true
	}
	if !strings.Contains(route, ":") && !strings.HasSuffix(route, "*") {
		return false
	}
	rp := strings.Split(strings.Trim(route, "/"), "/")
	pp := strings.Split(strings.Trim(path, "/"), "/")
	got := map[string]string{}
	for i, part := range rp {
		if part == "*" && i == len(rp)-1 {
			got["*"] = strings.Join(pp[min(i, len(pp)):], "/")
			goto ok
		}
		if i >= len(pp) {
			return false
		}
		if strings.HasPrefix(part, ":") {
			if pp[i] == "" {
				return false
			}
			v, err := url.PathUnescape(pp[i])
			if err != nil {
				v = pp[i]
			}
			got[part[1:]] = v
			continue
		}
		if part != pp[i] {
			return false
		}
	}
	if len(pp) != len(rp) {
		return false
	}
ok:
	keys := make([]string, 0, len(got))
	for k := range got {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		params.Set(rt.Str(k), rt.Str(got[k]))
	}
	return true
}

var events struct {
	mu   sync.Mutex
	subs map[string]map[chan string]bool
}

func serveEvents(w http.ResponseWriter, r *http.Request, channel string) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is not supported here", http.StatusInternalServerError)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	ch := make(chan string, 64)
	events.mu.Lock()
	if events.subs == nil {
		events.subs = map[string]map[chan string]bool{}
	}
	if events.subs[channel] == nil {
		events.subs[channel] = map[chan string]bool{}
	}
	events.subs[channel][ch] = true
	events.mu.Unlock()
	defer func() {
		events.mu.Lock()
		delete(events.subs[channel], ch)
		events.mu.Unlock()
	}()
	io.WriteString(w, ": salam\n\n")
	fl.Flush()
	tick := time.NewTicker(20 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-ch:
			io.WriteString(w, msg)
			fl.Flush()
		case <-tick.C:
			io.WriteString(w, ": hanuz\n\n")
			fl.Flush()
		}
	}
}

func fnSendEvent(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	ch, ok := args[0].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4364",
			"esm-e kanal bayad matn bashe",
			"the channel name has to be text, not %s", rt.TypeName(args[0]))
	}
	var data string
	if s, ok := args[1].(rt.Str); ok {
		data = string(s)
	} else {
		v, err := fnJSONWrite(in, args[1:2], node)
		if err != nil {
			return nil, err
		}
		data = string(v.(rt.Str))
	}
	var msg strings.Builder
	if len(args) > 2 {
		msg.WriteString("event: " + strings.ReplaceAll(rt.Show(args[2]), "\n", " ") + "\n")
	}
	for _, line := range strings.Split(data, "\n") {
		msg.WriteString("data: " + line + "\n")
	}
	msg.WriteString("\n")
	events.mu.Lock()
	defer events.mu.Unlock()
	n := 0
	for c := range events.subs[string(ch)] {
		select {
		case c <- msg.String():
			n++
		default:
		}
	}
	return rt.Int(int64(n)), nil
}

//go:embed mihan-android.js
var androidJS string

func fnAndroid(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	return rt.Bool(os.Getenv("MIHAN_ANDROID") != ""), nil
}

func requestLog(method, path string, status int, took time.Duration) {
	if os.Getenv("MIHAN_REQUEST_LOG") == "" {
		return
	}
	fmt.Fprintf(os.Stderr, "@@mihan-req %s %s %d %s\n", method, path, status, strconv.FormatFloat(took.Seconds()*1000, 'f', 1, 64))
}
