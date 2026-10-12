package eval

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"mihan/internal/diag"
	"mihan/internal/parser"
)

func startTestServer(t *testing.T, handler string) string {
	t.Helper()
	stopTestServer()
	src := handler + "\nserva_bgo(0, pasokh)\n"
	bag := &diag.Bag{}
	file := parser.Parse("test.mhn", []byte(src), bag)
	if bag.HasErrors() {
		t.Fatalf("the program did not read:\n%s", src)
	}
	var said bytes.Buffer
	in := New(&said, nil)
	stopped := make(chan error, 1)
	go func() { stopped <- in.RunFile(file) }()
	t.Cleanup(func() {
		stopTestServer()
		select {
		case <-stopped:
		case <-time.After(2 * time.Second):

		}
	})

	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("the program said:\n%s", said.String())
		}
	})
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if url := serverURL(); url != "" {
			if resp, err := http.Get(url); err == nil {
				resp.Body.Close()
				return url
			}
		}
		select {
		case err := <-stopped:
			t.Fatalf("the server stopped instead of serving: %v", err)
		default:
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("the server never came up")
	return ""
}

func serverURL() string {
	theServer.mu.Lock()
	s := theServer.s
	theServer.mu.Unlock()
	if s == nil {
		return ""
	}
	return s.url() + "/"
}

func stopTestServer() {
	theServer.mu.Lock()
	s := theServer.s
	theServer.s = nil
	theServer.mu.Unlock()
	if s == nil {
		return
	}
	if s.srv != nil {
		s.srv.Close()
	}
	s.signalStop()
}

func get(t *testing.T, url string) (int, string) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("asking %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func TestAServerAnswersARealRequest(t *testing.T) {
	url := startTestServer(t, `
funct pasokh(darakht)
    bargard {
        "vaziat": 200,
        "matn": $"salam, {darakht["path"]}!"
    }`)
	code, body := get(t, url+"salam")
	if code != 200 {
		t.Errorf("the server answered %d, want 200:\n%s", code, body)
	}
	if body != "salam, /salam!" {
		t.Errorf("the server said %q, want %q", body, "salam, /salam!")
	}
}

func TestAServerReadsAQueryString(t *testing.T) {
	url := startTestServer(t, `
funct pasokh(darakht)
    bargard $"nam: {darakht["sorok"]["nam"]}"`)
	_, body := get(t, url+"?nam=Mihan")
	if body != "nam: Mihan" {
		t.Errorf("the server said %q, want %q", body, "nam: Mihan")
	}
}

func TestAMissingStatusIsAnAnswerToo(t *testing.T) {
	url := startTestServer(t, `
funct pasokh(darakht)
    agar darakht["path"] == "/nast"
        bargard { "vaziat": 404, "matn": "page nist" }
    bargard { "vaziat": 200, "matn": "salam" }`)
	code, body := get(t, url+"nast")
	if code != 404 || body != "page nist" {
		t.Errorf("a missing page gave %d %q, want 404 \"page nist\"", code, body)
	}
	code, body = get(t, url+"hast")
	if code != 200 || body != "salam" {
		t.Errorf("a page that exists gave %d %q, want 200 \"salam\"", code, body)
	}
}

func TestATextAnswerIsServedAsItIs(t *testing.T) {
	url := startTestServer(t, `
funct pasokh(darakht)
    bargard "<h1>salam</h1>"`)
	code, body := get(t, url)
	if code != 200 || body != "<h1>salam</h1>" {
		t.Errorf("the server gave %d %q, want 200 \"<h1>salam</h1>\"", code, body)
	}
}

func TestAFailureInTheAnswerBecomesAPage(t *testing.T) {
	url := startTestServer(t, `
funct pasokh(darakht)
    x = 1 / 0
    bargard "happend"`)
	code, body := get(t, url)
	if code != 500 {
		t.Errorf("a failure inside the answer gave %d, want 500", code)
	}
	if !strings.Contains(body, "khata") {
		t.Errorf("the page should say there was a mistake, it said %q", body)
	}

	code, _ = get(t, url)
	if code != 500 {
		t.Errorf("the server did not survive a failure; the next request gave %d", code)
	}
}

func TestAServerCanSetAHeader(t *testing.T) {
	url := startTestServer(t, `
funct pasokh(darakht)
    bargard {
        "vaziat": 200,
        "matn": "{\"ok\": true}",
        "sarlawah": { "Content-Type": "application/json" }
    }`)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("the header is %q, want application/json", got)
	}
}

func TestAPortOfZeroStillGivesARealPort(t *testing.T) {
	url := startTestServer(t, `
funct pasokh(darakht)
    bargard "salam"`)
	if url == "http://127.0.0.1:0/" {
		t.Fatal("the machine chose port 0, which is not a port a browser can open")
	}
	_, body := get(t, url)
	if body != "salam" {
		t.Errorf("the server said %q, want salam", body)
	}
}
