package eval

import (
	"strings"
	"testing"
)

func TestFinglishStyleWords(t *testing.T) {
	out := run(t, `chap_khat sabk_khat({"rang": "siah", "gooshe_gerd": 30, "fasele_dakheli": "8px 12px", "zamine": "zard", "tarazbandi": "markaz", "pahna": "por"})`)
	want := "color:black;border-radius:30px;padding:8px 12px;background:#eab308;text-align:center;width:100%;"
	if strings.TrimSpace(out) != want {
		t.Errorf("got %q want %q", out, want)
	}
}

func TestEnglishStyleNamesStillWork(t *testing.T) {
	out := run(t, `chap_khat sabk_khat({"color": "black", "border_radius": 30, "border-radius": "1rem", "z-index": 3})`)
	want := "color:black;border-radius:30px;border-radius:1rem;z-index:3;"
	if strings.TrimSpace(out) != want {
		t.Errorf("got %q want %q", out, want)
	}
}

func TestStyleValuesTranslateInsideText(t *testing.T) {
	out := run(t, `chap_khat sabk_khat({"hashie": "2px khat_chin ghermez", "tabdil": "charkhesh(45deg) bozorgi(1.2)", "enteghal": "rang 0.3s narm", "zamine": "aks(zard.png)", "harekat": "biya 1s yeknavakht bi_payan"})`)
	want := "border:2px dashed #dc2626;transform:rotate(45deg) scale(1.2);transition:color 0.3s ease;background:url(zard.png);animation:biya 1s linear infinite;"
	if strings.TrimSpace(out) != want {
		t.Errorf("got %q want %q", out, want)
	}
}

func TestStyleSheetWithAnimationAndMedia(t *testing.T) {
	out := run(t, `chap_khat sabk({"@harekat biya": {"az": {"shafafiat": 0}, "ta": {"shafafiat": 1}}, "onvan": {"harekat": "biya 1s narm"}, ".karte:rooye": {"tabdil": "bozorgi(1.05)"}, "@kochak_tar 600": {"jabe": {"fasele_dakheli": 8}}})`)
	for _, want := range []string{"@keyframes biya {", "from {opacity:0;}", "to {opacity:1;}", "h1 {animation:biya 1s ease;}", ".karte:hover {transform:scale(1.05);}", "@media (max-width:600px) {", "div {padding:8px;}"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %q", want, out)
		}
	}
}

func TestUnknownStyleWordSuggests(t *testing.T) {
	err := runFail(t, `chap_khat sabk_khat({"gooshe_gard": 3})`)
	if !strings.Contains(err, "gooshe_gerd") || !strings.Contains(err, "nemishnasam") {
		t.Errorf("got %q", err)
	}
}

func TestPageTakesAStyleSheetAndClasses(t *testing.T) {
	out := run(t, `chap_khat safhe_kon("x", jabe([band("salam", {"kelas": "kart", "rang": "abi"})]), {".kart": {"saye": "0 2px 8px khakestari"}})`)
	for _, want := range []string{".kart {box-shadow:0 2px 8px #6b7280;}", `class="kart"`, "color:#2563eb"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestRangYaKhataCountsToTheEnd(t *testing.T) {
	out := run(t, "chap_khat rang_ya_khata(2, 5)\nchap_khat rang_ya_khata(3)")
	if strings.TrimSpace(out) != "[2, 3, 4]\n[0, 1, 2]" {
		t.Errorf("got %q", out)
	}
}
