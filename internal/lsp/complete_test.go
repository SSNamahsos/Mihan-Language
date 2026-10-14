package lsp

import (
	"strings"
	"testing"
)

func TestADotAfterTextOffersTextMethods(t *testing.T) {
	_, c := newClient(t)
	c.open("esm = \"Sara\"\nesm.\n")
	have := labels(c.complete(1, 4))
	for _, want := range []string{"gooshband", "tab", "upper", "split"} {
		if _, ok := have[want]; !ok {
			t.Errorf("%s should be offered after a matn, got %v", want, keysOf(have))
		}
	}
	if _, ok := have["bast"]; ok {
		t.Error("a list method is offered on a matn")
	}
}

func TestADotAfterAChainFollowsTheResult(t *testing.T) {
	_, c := newClient(t)
	c.open("esm = \"a,b\"\nesm.tab(\",\").\n")
	have := labels(c.complete(1, 13))
	if _, ok := have["bast"]; !ok {
		t.Errorf("tab gives a list, so bast should be offered, got %v", keysOf(have))
	}
}

func TestKhodOffersTheFieldsOfTheClass(t *testing.T) {
	_, c := newClient(t)
	c.open("darghan A\n    funct __sakht__(n)\n        khod.naam = n\n    funct goftan()\n        chap_khat khod.\n")
	have := labels(c.complete(4, 23))
	if _, ok := have["naam"]; !ok {
		t.Errorf("naam should be offered after khod., got %v", keysOf(have))
	}
	if _, ok := have["goftan"]; !ok {
		t.Errorf("goftan should be offered after khod., got %v", keysOf(have))
	}
}

func TestNothingIsOfferedInsideTextOrNotes(t *testing.T) {
	_, c := newClient(t)
	c.open("x = \"sal\"\n# ag\n")
	if n := len(c.complete(0, 8).Items); n != 0 {
		t.Errorf("inside text nothing should be offered, got %d", n)
	}
	if n := len(c.complete(1, 4).Items); n != 0 {
		t.Errorf("inside a note nothing should be offered, got %d", n)
	}
}

func TestBlocksAreOnlyOfferedAtTheStartOfALine(t *testing.T) {
	_, c := newClient(t)
	c.open("x = 1\ny = x + t\n")
	have := labels(c.complete(1, 9))
	if _, ok := have["darghan"]; ok {
		t.Error("darghan cannot follow a +")
	}
	if _, ok := have["tule"]; !ok {
		t.Error("tule should be offered inside a sum")
	}
}

func TestSignatureHelpForAMethodAndAConstructor(t *testing.T) {
	_, c := newClient(t)
	c.open("xs = [1]\nxs.biya(0, \n")
	h := c.signature(1, 11)
	if h == nil || len(h.Signatures) == 0 || !strings.HasPrefix(h.Signatures[0].Label, "biya(") || h.ActiveParameter != 1 {
		t.Fatalf("biya's parameters should be shown, got %+v", h)
	}
	c.open("darghan P\n    funct __sakht__(a, b)\n        khod.a = a\np = P(1, \n")
	h = c.signature(3, 9)
	if h == nil || len(h.Signatures) == 0 || h.Signatures[0].Label != "P(a, b)" {
		t.Fatalf("the constructor's parameters should be shown, got %+v", h)
	}
}
