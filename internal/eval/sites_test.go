package eval

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mihan/internal/diag"
	"mihan/internal/parser"
)

func startServingFile(t *testing.T, rel string) string {
	t.Helper()
	stopTestServer()
	path := filepath.Join(repoRoot(t), rel)
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	bag := &diag.Bag{}
	file := parser.Parse(path, src, bag)
	if bag.HasErrors() {
		t.Fatalf("%s did not read: %s", rel, bag.Sorted()[0].EN)
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
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if u := serverURL(); u != "" {
			return strings.TrimRight(u, "/")
		}
		select {
		case err := <-stopped:
			t.Fatalf("%s stopped instead of serving: %v\n%s", rel, err, said.String())
		default:
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("%s never came up; it said:\n%s", rel, said.String())
	return ""
}

func TestTheCalculatorPageAndItsAPI(t *testing.T) {
	base := startServingFile(t, filepath.Join("examples", "site", "hesab", "hesab.mhn"))
	code, body := get(t, base+"/")
	if code != 200 || !strings.Contains(body, "hesab-safhe") {
		t.Fatalf("the page was %d %q", code, body)
	}
	_, js := get(t, base+"/app.js")
	if !strings.Contains(js, "/api/hesab") {
		t.Fatalf("the page does not ask the program:\n%s", js)
	}
	resp, err := http.Post(base+"/api/hesab", "application/json", strings.NewReader(`{"a":6,"b":7}`))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || string(got) != `{"jam":13,"zarb":42,"kam":-1}` {
		t.Fatalf("the arithmetic was %d %q", resp.StatusCode, got)
	}
}

func TestTheNotesPageKeepsWhatWasPosted(t *testing.T) {
	note := filepath.Join(t.TempDir(), "yaddasht.json")
	t.Setenv("MIHAN_YADDASHT", note)
	base := startServingFile(t, filepath.Join("examples", "site", "yaddasht", "yaddasht.mhn"))
	code, body := get(t, base+"/")
	if code != 200 || !strings.Contains(body, "yaddasht-safhe") {
		t.Fatalf("the page was %d %q", code, body)
	}
	code, body = get(t, base+"/api/yaddasht")
	if code != 200 || body != "[]" {
		t.Fatalf("a new notebook was %d %q", code, body)
	}
	resp, err := http.Post(base+"/api/yaddasht", "application/json", strings.NewReader(`{"matn":"salam-az-safhe"}`))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(got), `"matn":"salam-az-safhe"`) {
		t.Fatalf("posting a note answered %d %q", resp.StatusCode, got)
	}
	saved, err := os.ReadFile(note)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(saved), "salam-az-safhe") {
		t.Fatalf("the note was not written down: %s", saved)
	}
}
