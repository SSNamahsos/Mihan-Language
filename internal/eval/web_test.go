package eval

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func ask(t *testing.T, c *http.Client, method, url, ctype, body string, hdr map[string]string) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest(method, url, strings.NewReader(body))
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	if c == nil {
		c = http.DefaultClient
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, string(b)
}

const apiProgram = `
kala = {"1": {"id": 1, "nam": "ketab"}}
funct ye_kala(darakht)
    id = darakht["params"]["id"]
    agar darakht["method"] == "GET"
        agar kala.begir(id, hich) == hich
            bargard json_pasokh({"khata": "nist"}, 404)
        bargard json_pasokh(kala[id])
    agar darakht["method"] == "PUT"
        kala[id] = json_bekhan(darakht["matn"])
        bargard json_pasokh(kala[id])
    agar darakht["method"] == "DELETE"
        kala.hazf(id)
        bargard {"vaziat": 204, "matn": ""}
    bargard {"vaziat": 405, "matn": "na"}

funct hame(darakht)
    agar darakht["method"] == "POST"
        x = json_bekhan(darakht["matn"])
        kala[neshan(x["id"])] = x
        bargard json_pasokh(x, 201)
    bargard {"vaziat": 200, "matn": json_benevis(kala), "sarlawah": {"Content-Type": "application/json", "X-Tedad": neshan(tule(kala))}, "cors": dorost}

funct jostoju(darakht)
    bargard json_pasokh({"q": darakht["sorok"].begir("q", ""), "agent": darakht["sarlawah"].begir("X-Test", "")})

funct pasokh(darakht)
    agar darakht["method"] == "OPTIONS"
        bargard {"vaziat": 204, "matn": "", "cors": "http://example.com"}
    bargard rah(darakht, [
        {"path": "/api/kala", "tabe": hame},
        {"path": "/api/kala/:id", "tabe": ye_kala},
        {"path": "/api/jostoju", "tabe": jostoju, "method": "GET"},
    ])
`

func TestWebJSONAPI(t *testing.T) {
	base := strings.TrimSuffix(startTestServer(t, apiProgram), "/")
	resp, body := ask(t, nil, "GET", base+"/api/kala/1", "", "", nil)
	if resp.StatusCode != 200 || !strings.Contains(resp.Header.Get("Content-Type"), "application/json") || !strings.Contains(body, `"ketab"`) {
		t.Fatalf("GET: %d %s %q", resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
	resp, body = ask(t, nil, "POST", base+"/api/kala", "application/json", `{"id": 2, "nam": "medad"}`, nil)
	if resp.StatusCode != 201 || !strings.Contains(body, "medad") {
		t.Fatalf("POST: %d %q", resp.StatusCode, body)
	}
	resp, body = ask(t, nil, "PUT", base+"/api/kala/2", "application/json", `{"id": 2, "nam": "khodkar"}`, nil)
	if resp.StatusCode != 200 || !strings.Contains(body, "khodkar") {
		t.Fatalf("PUT: %d %q", resp.StatusCode, body)
	}
	resp, body = ask(t, nil, "GET", base+"/api/kala", "", "", nil)
	var all map[string]any
	if err := json.Unmarshal([]byte(body), &all); err != nil || len(all) != 2 || resp.Header.Get("X-Tedad") != "2" {
		t.Fatalf("list: %q %v %q", body, err, resp.Header.Get("X-Tedad"))
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("cors: dorost did not allow every origin")
	}
	resp, _ = ask(t, nil, "DELETE", base+"/api/kala/2", "", "", nil)
	if resp.StatusCode != 204 {
		t.Fatalf("DELETE: %d", resp.StatusCode)
	}
	resp, _ = ask(t, nil, "GET", base+"/api/kala/2", "", "", nil)
	if resp.StatusCode != 404 {
		t.Fatalf("deleted item: %d", resp.StatusCode)
	}
	resp, body = ask(t, nil, "GET", base+"/api/jostoju?q=sib+zard", "", "", map[string]string{"X-Test": "ok"})
	if !strings.Contains(body, `"sib zard"`) || !strings.Contains(body, `"ok"`) {
		t.Fatalf("query and header: %q", body)
	}
	resp, _ = ask(t, nil, "POST", base+"/api/jostoju", "", "", nil)
	if resp.StatusCode != 405 {
		t.Fatalf("a route for GET only answered POST with %d", resp.StatusCode)
	}
	resp, _ = ask(t, nil, "OPTIONS", base+"/api/kala", "", "", map[string]string{"Origin": "http://example.com", "Access-Control-Request-Method": "PUT"})
	if resp.StatusCode != 204 || resp.Header.Get("Access-Control-Allow-Origin") != "http://example.com" ||
		!strings.Contains(resp.Header.Get("Access-Control-Allow-Methods"), "PUT") || resp.Header.Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("preflight: %d %v", resp.StatusCode, resp.Header)
	}
}

func TestWebCookiesSessionsUploads(t *testing.T) {
	base := strings.TrimSuffix(startTestServer(t, `
funct pasokh(darakht)
    agar darakht["path"] == "/shomar"
        j = jalase(darakht)
        j["n"] = j.begir("n", 0) + 1
        bargard neshan(j["n"])
    agar darakht["path"] == "/cookie"
        bargard {"vaziat": 200, "matn": darakht["cookie"].begir("rang", "hich"), "cookie": {"rang": "abi", "movaghat": {"meghdar": "x", "umr": 60, "httponly": ghalat}}}
    agar darakht["path"] == "/upload"
        f = darakht["parvande"][0]
        bargard $"{darakht["form"]["onvan"]} {f["nam"]} {f["esm"]} {f["noe"]} {f["andaze"]} {f["matn"]}"
    agar darakht["path"] == "/form"
        bargard darakht["form"]["a"] + "|" + darakht["form"]["b"]
    bargard "?"
`), "/")
	jar, _ := cookiejar.New(nil)
	c := &http.Client{Jar: jar}
	for want := 1; want <= 3; want++ {
		if _, body := ask(t, c, "GET", base+"/shomar", "", "", nil); body != string(rune('0'+want)) {
			t.Fatalf("session visit %d said %q", want, body)
		}
	}
	if _, body := ask(t, nil, "GET", base+"/shomar", "", "", nil); body != "1" {
		t.Fatalf("a new visitor shared the session: %q", body)
	}
	resp, body := ask(t, c, "GET", base+"/cookie", "", "", nil)
	if body != "hich" || len(resp.Cookies()) != 2 {
		t.Fatalf("cookie: %q %v", body, resp.Header["Set-Cookie"])
	}
	for _, ck := range resp.Cookies() {
		if ck.Name == "movaghat" && (ck.MaxAge != 60 || ck.HttpOnly) {
			t.Fatalf("cookie options were not kept: %+v", ck)
		}
	}
	if _, body := ask(t, c, "GET", base+"/cookie", "", "", nil); body != "abi" {
		t.Fatalf("the cookie did not come back: %q", body)
	}
	var mb bytes.Buffer
	mw := multipart.NewWriter(&mb)
	mw.WriteField("onvan", "aks")
	fw, _ := mw.CreateFormFile("file", "salam.txt")
	fw.Write([]byte("salam donya"))
	mw.Close()
	if _, body := ask(t, nil, "POST", base+"/upload", mw.FormDataContentType(), mb.String(), nil); body != "aks file salam.txt application/octet-stream 11 salam donya" {
		t.Fatalf("upload: %q", body)
	}
	if _, body := ask(t, nil, "POST", base+"/form", "application/x-www-form-urlencoded", "a=1+2&b=%D8%B3", nil); body != "1 2|س" {
		t.Fatalf("form: %q", body)
	}
}

func TestWebStaticCachingTemplatesEvents(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "style.css"), []byte("@keyframes chalesh { from { opacity: 0 } to { opacity: 1 } }\n.k { animation: chalesh 1s; transition: transform .3s }"), 0o644)
	os.WriteFile(filepath.Join(dir, "app.js"), []byte("document.title = 'js';"), 0o644)
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<script src='/_mihan/mihan.js'></script>"), 0o644)
	base := strings.TrimSuffix(startTestServer(t, `
funct pasokh(darakht)
    agar darakht["path"] == "/qaleb"
        bargard qaleb("<b>{{nam}}</b> {{{khas}}} {{a.b}}", {"nam": "<i>x</i>", "khas": "<u>y</u>", "a": {"b": 3}})
    agar darakht["path"] == "/ferest"
        bargard neshan(rooydad_ferest("chat", {"payam": darakht["sorok"]["p"]}, "jadid"))
    bargard parvande_pasokh("`+filepath.ToSlash(dir)+`", darakht["path"])
`), "/")
	for path, ctype := range map[string]string{"/style.css": "text/css", "/app.js": "text/javascript", "/": "text/html"} {
		resp, body := ask(t, nil, "GET", base+path, "", "", nil)
		if resp.StatusCode != 200 || !strings.HasPrefix(resp.Header.Get("Content-Type"), ctype) {
			t.Fatalf("%s: %d %s", path, resp.StatusCode, resp.Header.Get("Content-Type"))
		}
		tag := resp.Header.Get("ETag")
		if tag == "" || resp.Header.Get("Cache-Control") != "no-cache" || resp.Header.Get("Last-Modified") == "" {
			t.Fatalf("%s has no caching headers: %v", path, resp.Header)
		}
		if path == "/style.css" && !strings.Contains(body, "@keyframes") {
			t.Fatal("the CSS animation did not arrive whole")
		}
		resp, body = ask(t, nil, "GET", base+path, "", "", map[string]string{"If-None-Match": tag})
		if resp.StatusCode != 304 || body != "" {
			t.Fatalf("%s with its tag: %d %q", path, resp.StatusCode, body)
		}
	}
	if _, body := ask(t, nil, "GET", base+"/qaleb", "", "", nil); body != "<b>&lt;i&gt;x&lt;/i&gt;</b> <u>y</u> 3" {
		t.Fatalf("qaleb: %q", body)
	}
	resp, body := ask(t, nil, "GET", base+"/_mihan/mihan.js", "", "", nil)
	if !strings.Contains(body, "window.mihan") || !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/javascript") {
		t.Fatal("mihan.js is not served")
	}

	sse, err := http.Get(base + "/_mihan/rooydad/chat")
	if err != nil {
		t.Fatal(err)
	}
	defer sse.Body.Close()
	if sse.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("events: %s", sse.Header.Get("Content-Type"))
	}
	rd := bufio.NewReader(sse.Body)
	rd.ReadString('\n')
	rd.ReadString('\n')
	if _, n := ask(t, nil, "GET", base+"/ferest?p=salam", "", "", nil); n != "1" {
		t.Fatalf("one page was listening, the program counted %s", n)
	}
	got := make(chan string, 1)
	go func() {
		a, _ := rd.ReadString('\n')
		b, _ := rd.ReadString('\n')
		got <- a + b
	}()
	select {
	case s := <-got:
		if s != "event: jadid\ndata: {\"payam\":\"salam\"}\n" {
			t.Fatalf("event: %q", s)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("the event never arrived")
	}
}
