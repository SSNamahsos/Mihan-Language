package eval

import (
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func TestAPageIsWrittenWithFinglishWords(t *testing.T) {
	got := run(t, `
s = safhe_kon("khane", {
    "gone": "jabe",
    "rang": "black",
    "khat_dor": 20,
    "pas_zamine": "meshki",
    "matn": "salam <donya>"
})
chap_khat s
`)
	for _, want := range []string{
		"color:black;",
		"border-radius:20px;",
		"background:black;",
		"salam &lt;donya&gt;",
		"<title>khane</title>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the page should contain %q\n%s", want, got)
		}
	}
	if !strings.Contains(got, "addEventListener") {
		t.Error("the page should carry the language's own click script")
	}
	if strings.Contains(got, "<donya>") {
		t.Error("text the writer put on the page was treated as a tag")
	}
}

func TestAnUnknownStyleWordIsNamed(t *testing.T) {
	checkFail(t, `safhe_kon({"ghalat_style": "1"})`, "nemishnasam")
}

func TestAFormComesBackAsADict(t *testing.T) {
	check(t, `
f = form_bekhan("naam=Sara+Khane&sen=20")
chap_khat f["naam"]
chap_khat f["sen"]
`, "Sara Khane\n20\n")
}

func TestPiecesAndRoutesAreThePage(t *testing.T) {
	check(t, `
funct khane(darakht)
    bargard dokme("bekhar", "bekhar", {"khat_dor": 20})
javab = rah({"path": "/", "method": "GET", "matn": ""}, [
    {"path": "/", "onvan": "khane", "tabe": khane}
])
chap_khat javab["vaziat"]
chap_khat "data-kar=\"bekhar\"" dar javab["matn"]
chap_khat "border-radius:20px" dar javab["matn"]
nist = rah({"path": "/no", "method": "GET", "matn": ""}, [
    {"path": "/", "onvan": "khane", "tabe": khane}
])
chap_khat nist["vaziat"]
`, "200\ndorost\ndorost\n404\n")
}

func TestTheBarnameSiteIsThreePages(t *testing.T) {
	base := startServingFile(t, filepath.Join("examples", "site", "barname", "barname.mhn"))
	code, home := get(t, base+"/")
	if code != 200 || !strings.Contains(home, "foroosh") || !strings.Contains(home, "<nav") {
		t.Fatalf("the home page was %d and did not carry the bar of pages", code)
	}
	code, shop := get(t, base+"/foroosh")
	if code != 200 || !strings.Contains(shop, "sefaresh: 0") || !strings.Contains(shop, "data-kar=\"bekhar\"") {
		t.Fatalf("the shop was %d:\n%s", code, shop)
	}
	resp, err := http.Post(base+"/foroosh", "application/x-www-form-urlencoded", strings.NewReader("kar=bekhar"))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(body), "sefaresh: 1") {
		t.Fatalf("the button did not reach the program: %d %s", resp.StatusCode, body)
	}
	code, about := get(t, base+"/darbare")
	if code != 200 || !strings.Contains(about, "darbare") {
		t.Fatalf("the about page was %d", code)
	}
	code, _ = get(t, base+"/nist")
	if code != 404 {
		t.Errorf("a missing page answered %d", code)
	}
}

func TestTheRangSiteIsAMihanPage(t *testing.T) {
	base := startServingFile(t, filepath.Join("examples", "site", "rang", "rang.mhn"))
	code, body := get(t, base+"/")
	if code != 200 {
		t.Fatalf("the page answered %d", code)
	}
	if !strings.Contains(body, "border-radius:20px;") || !strings.Contains(body, "salam doost") {
		t.Fatalf("the page was not built from rang and khat_dor:\n%s", body)
	}
	resp, err := http.PostForm(base+"/", url.Values{"naam": {"Sara"}})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(got), "salam Sara") {
		t.Fatalf("the greeting was %d %q", resp.StatusCode, got)
	}
}
