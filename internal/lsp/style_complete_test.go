package lsp

import "testing"

func TestStyleKeysAreOfferedInADict(t *testing.T) {
	_, c := newClient(t)
	c.open("s = {\"\n")
	have := labels(c.complete(0, 6))
	for _, want := range []string{"rang", "gooshe_gerd", "fasele_dakheli", "tabdil"} {
		if _, ok := have[want]; !ok {
			t.Errorf("%s should be offered as a style word, got %d items", want, len(have))
		}
	}
}

func TestStyleValuesAreOfferedAfterAColon(t *testing.T) {
	_, c := newClient(t)
	c.open("s = {\"rang\": \"\n")
	have := labels(c.complete(0, 14))
	for _, want := range []string{"siah", "ghermez", "markaz"} {
		if _, ok := have[want]; !ok {
			t.Errorf("%s should be offered as a style value", want)
		}
	}
}

func TestPlainTextOffersNothing(t *testing.T) {
	_, c := newClient(t)
	c.open("chap_khat \"sal\n")
	if have := labels(c.complete(0, 14)); len(have) != 0 {
		t.Errorf("nothing should be offered inside plain text, got %d", len(have))
	}
}
