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

func mhnPath(t *testing.T, p string) string {
	t.Helper()
	s := filepath.ToSlash(p)
	if strings.ContainsAny(s, "\"\n") {
		t.Fatalf("this path cannot sit in a Mihan string: %s", s)
	}
	return s
}

func TestAFileRoundTripKeepsTheText(t *testing.T) {
	dir := t.TempDir()
	note := mhnPath(t, filepath.Join(dir, "note.txt"))
	check(t, `
benevis "`+note+`", "سلام"
chap_khat bekhan("`+note+`")
`, "سلام\n")
}

func TestAMissingFileIsNamed(t *testing.T) {
	dir := t.TempDir()
	note := mhnPath(t, filepath.Join(dir, "nist.txt"))
	checkFail(t, `bekhan "`+note+`"`, "peyda nashod")
}

func TestAFolderIsNotReadAsText(t *testing.T) {
	dir := mhnPath(t, t.TempDir())
	checkFail(t, `bekhan "`+dir+`"`, "pooshe")
}

func TestAFolderListsNamesInOrder(t *testing.T) {
	dir := t.TempDir()
	folder := mhnPath(t, filepath.Join(dir, "daftar"))
	check(t, `
pooshe "`+folder+`"
benevis "`+folder+`/b.txt", "b"
benevis "`+folder+`/a.txt", "a"
chap_khat fehrest("`+folder+`")
hazf "`+folder+`/a.txt"
hazf "`+folder+`/b.txt"
hazf "`+folder+`"
chap_khat masir_hast("`+folder+`")
`, "[\"a.txt\", \"b.txt\"]\nghalat\n")
}

func TestHazfWillNotEmptyAFolder(t *testing.T) {
	dir := t.TempDir()
	folder := mhnPath(t, filepath.Join(dir, "por"))
	checkFail(t, `
pooshe "`+folder+`"
benevis "`+folder+`/a.txt", "a"
hazf "`+folder+`"
`, "hanooz chizi too khodesh dare")
}

func TestJSONKeepsOrderAndTheLanguageRules(t *testing.T) {
	check(t, `
chap_khat json_benevis({"b": 1, "a": "سلام"})
chap_khat json_bekhan("2")
chap_khat json_bekhan("2.0")
chap_khat json_bekhan("true")
`, "{\"b\":1,\"a\":\"سلام\"}\n2\n2.0\ndorost\n")
}

func TestJSONTruthRoundTripsAsTheWordTrue(t *testing.T) {
	check(t, `
chap_khat json_bekhan("true")
chap_khat json_benevis(dorost)
chap_khat json_benevis([dorost, ghalat, hich])
`, "dorost\ntrue\n[true,false,null]\n")
}

func TestAJSONKeyThatIsNotTextIsRefused(t *testing.T) {
	checkFail(t, `chap_khat json_benevis({1: "a"})`, "kelid-e JSON faghat mitoone matn bashe")
}

func TestAnEnvironmentVariableIsTextOrNothing(t *testing.T) {
	t.Setenv("MIHAN_YADDASHT", "salam")
	check(t, `
chap_khat mohit("MIHAN_YADDASHT")
chap_khat mohit("MIHAN_NO_SUCH_VAR_ZZ")
`, "salam\nhich\n")
}

func TestAPageCannotLeaveItsFolder(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "safhe")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("salam-safhe"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "style.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("no"), 0o644); err != nil {
		t.Fatal(err)
	}
	page := mhnPath(t, root)
	check(t, `
chap_khat parvande_pasokh("`+page+`", "/")["matn"]
chap_khat parvande_pasokh("`+page+`", "/../secret.txt")["vaziat"]
chap_khat parvande_pasokh("`+page+`", "/../secret.txt")["matn"]
chap_khat parvande_pasokh("`+page+`", "\\..\\secret.txt")["matn"]
chap_khat parvande_pasokh("`+page+`", "/style.css")["sarlawah"]["Content-Type"]
chap_khat parvande_pasokh("`+page+`", "/nist")["vaziat"]
`, "salam-safhe\n404\nnist\nnist\ntext/css; charset=utf-8\n404\n")
}

func TestDarkhastReadsWhatAServerSaid(t *testing.T) {
	url := strings.TrimRight(startTestServer(t, `
funct pasokh(darakht)
    bargard { "vaziat": 201, "matn": darakht["method"] + ":" + darakht["matn"] }
`), "/")
	check(t, `
d = darkhast("`+url+`", {"method": "POST", "matn": "salam"})
chap_khat d["vaziat"]
chap_khat d["matn"]
`, "201\nPOST:salam\n")
}

func TestDarkhastRefusesAFileAddress(t *testing.T) {
	checkFail(t, `darkhast "file:///c:/notes.txt"`, "http va https")
}

func TestTheShopPageAndItsAPI(t *testing.T) {
	stopTestServer()
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "site", "dokan", "dokan.mhn")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	bag := &diag.Bag{}
	file := parser.Parse(path, src, bag)
	if bag.HasErrors() {
		t.Fatalf("the shop did not read: %s", bag.Sorted()[0].EN)
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
	var base string
	for time.Now().Before(deadline) {
		if u := serverURL(); u != "" {
			base = strings.TrimRight(u, "/")
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if base == "" {
		t.Fatalf("the shop never came up; it said:\n%s", said.String())
	}

	code, body := get(t, base+"/")
	if code != 200 || !strings.Contains(body, "dokan-safhe") {
		t.Fatalf("the page was %d %q", code, body)
	}
	if ct := headerOf(t, base+"/"); !strings.Contains(ct, "text/html") {
		t.Errorf("the page type was %s", ct)
	}
	_, js := get(t, base+"/app.js")
	if !strings.Contains(js, "/api/kala") {
		t.Fatalf("the page's script does not ask the program:\n%s", js)
	}
	_, css := get(t, base+"/style.css")
	if !strings.Contains(css, "dokan") && !strings.Contains(css, "--bg") {
		t.Fatalf("the stylesheet did not come back:\n%s", css)
	}

	code, body = get(t, base+"/api/kala")
	want := `[{"nam":"nan","gheymat":2},{"nam":"shir","gheymat":5}]`
	if code != 200 || body != want {
		t.Fatalf("the list was %d %q, want %q", code, body, want)
	}

	resp, err := http.Post(base+"/api/kala", "application/json", strings.NewReader(`{"nam":"chay","gheymat":3}`))
	if err != nil {
		t.Fatal(err)
	}
	posted, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || string(posted) != `{"nam":"chay","gheymat":3}` {
		t.Fatalf("adding an item answered %d %q", resp.StatusCode, posted)
	}
	_, body = get(t, base+"/api/kala")
	if !strings.Contains(body, `"nam":"chay"`) || !strings.HasPrefix(body, `[{"nam":"nan"`) {
		t.Fatalf("the list forgot its order or the new item: %s", body)
	}

	code, body = get(t, base+"/nist")
	if code != 404 || body != "nist" {
		t.Fatalf("a missing page was %d %q", code, body)
	}
}

func headerOf(t *testing.T, url string) string {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.Header.Get("Content-Type")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find the module root")
		}
		dir = parent
	}
}
