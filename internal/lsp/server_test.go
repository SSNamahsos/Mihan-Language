package lsp

import (
	"strings"
	"testing"

	"mihan/internal/diag"
	"mihan/internal/format"
	"mihan/internal/parser"
)

func TestSuggestionsOfferTheWordsTheLanguageKnows(t *testing.T) {
	_, c := newClient(t)
	c.open("x = 1\n")
	list := c.complete(1, 0)
	have := labels(list)

	chap, ok := have["chap"]
	if !ok {
		t.Fatal("chap is one of the first words anyone writes, and it is not on offer")
	}
	if chap.Kind != KindFunction {
		t.Errorf("chap should be offered as something that does something, not as kind %d", chap.Kind)
	}
	if chap.Documentation == nil || !strings.Contains(chap.Documentation.Value, "write") {
		t.Errorf("chap should arrive with the line that says what it does, got %+v", chap.Documentation)
	}

	if _, ok := have["har"]; !ok {
		t.Error("har is how a loop is written, and it is not on offer")
	}
	if _, ok := have["agar"]; !ok {
		t.Error("agar is how a branch is written, and it is not on offer")
	}

	if _, ok := have["x"]; !ok {
		t.Error("x is the name in the file, and it is not on offer")
	}

	if _, ok := have["frobnicate"]; ok {
		t.Error("a word that is not in the language is on offer, which means the list has drifted")
	}
}

func TestDefCompletesIntoAFunction(t *testing.T) {
	_, c := newClient(t)
	c.open("x = 1\n")
	have := labels(c.complete(1, 0))
	def, ok := have["def"]
	if !ok {
		t.Fatal("def is how a reader starts a function, and it is not on offer")
	}
	if def.InsertTextFormat != InsertSnippet {
		t.Fatalf("def should insert a snippet, format %d", def.InsertTextFormat)
	}
	if !strings.HasPrefix(def.InsertText, "funct ${1:name}(") {
		t.Errorf("accepting def should write a funct and stop on the name, got %q", def.InsertText)
	}
	if _, ok := have["if"]; !ok {
		t.Error("if should be on offer, as the English way to write agar")
	}
	agar := have["agar"]
	if !strings.Contains(agar.InsertText, "agar ${1:shart}") {
		t.Errorf("agar should bring its block with it, got %q", agar.InsertText)
	}
}

func TestNamesInTheFileComeFirst(t *testing.T) {
	_, c := newClient(t)
	c.open("shanje_man = 1\n")
	list := c.complete(1, 0)
	if len(list.Items) < 2 {
		t.Fatalf("there should be plenty to offer, got %d", len(list.Items))
	}
	if list.Items[0].Label != "shanje_man" {
		t.Errorf("the first suggestion is %q, but the name in the file should be first", list.Items[0].Label)
	}
}

func TestADotOffersTheMethodsOfTheValueInFront(t *testing.T) {
	_, c := newClient(t)
	text := "darghan Adad\n" +
		"    adad = 0\n" +
		"\n" +
		"    __sakht__(adad)\n" +
		"        khod.adad = adad\n" +
		"\n" +
		"    chap_kon()\n" +
		"        chap_khat khod.adad\n" +
		"\n" +
		"    zoj_kon()\n" +
		"        chap_khat khod.adad * 2\n" +
		"\n" +
		"a = Adad(1)\n" +
		"a."

	c.open(text)
	list := c.complete(13, 2)
	have := labels(list)
	for _, want := range []string{"chap_kon", "zoj_kon"} {
		if _, ok := have[want]; !ok {
			t.Errorf("after `a.` the method %q should be on offer, but only %v are", want, keysOf(have))
		}
	}

	if _, ok := have["har"]; ok {
		t.Error("after a dot the whole language should not be offered")
	}

}

func keysOf(m map[string]CompletionItem) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestMistakesAreSaidAsSoonAsAFileIsOpen(t *testing.T) {
	_, c := newClient(t)
	found := c.open("x = \n")
	if len(found) == 0 {
		t.Fatal("a file that assigns nothing should be reported while it is open")
	}
	d := found[0]
	if d.Severity != 1 {
		t.Errorf("a mistake should be marked as one, not as a warning: %d", d.Severity)
	}
	if d.Code == "" {
		t.Error("every mistake carries a code, so a reader can look it up")
	}
	if d.Source != "mihan" {
		t.Errorf("the mistake should say it came from Mihan, not from %q", d.Source)
	}
	if !strings.Contains(d.Message, "khata") && d.Message == "" {
		t.Error("the message is empty, which is the one thing a mistake must not be")
	}
}

func TestAFixedFileIsNoLongerReported(t *testing.T) {
	_, c := newClient(t)
	if got := c.open("x = \n"); len(got) == 0 {
		t.Fatal("the broken file should be reported")
	}
	if got := c.open("x = 1\nchap_khat x\n"); len(got) != 0 {
		t.Errorf("a file with nothing wrong with it should have nothing said about it, got %d", len(got))
	}
}

func TestAJumpLandsWhereTheNameWasWritten(t *testing.T) {
	_, c := newClient(t)

	c.open("x = 1\nchap_khat x\n")
	places := c.definition(1, 10)
	if len(places) == 0 {
		t.Fatal("nothing was said about where x is written, which is the whole point")
	}
	if places[0].URI != "file:///t.mhn" {
		t.Errorf("the jump points at %q, which is not the file being read", places[0].URI)
	}
	if places[0].Range.Start.Line != 0 {
		t.Errorf("x is written on line 1, but the jump points at line %d", places[0].Range.Start.Line)
	}
	if places[0].Range.Start.Character != 0 {
		t.Errorf("x is written at the first character, but the jump points at %d",
			places[0].Range.Start.Character)
	}
}

func TestAJumpFromInsideAFunctLandsInsideIt(t *testing.T) {
	_, c := newClient(t)
	text := "x = 1\nfunct f()\n    x = 2\n    chap_khat x\nf()\n"
	c.open(text)

	places := c.definition(3, 14)
	if len(places) == 0 {
		t.Fatal("the read of x inside f found nothing")
	}
	if places[0].Range.Start.Line != 2 {
		t.Errorf("the read of x inside f means the x on line 3, but the jump points at line %d:\n%s",
			places[0].Range.Start.Line, text)
	}
}

func TestHoveringSaysWhatANameIs(t *testing.T) {
	_, c := newClient(t)
	c.open("funct jam(a, b)\n    bargard a * b\njam(1, 2)\n")

	h := c.hover(2, 1)
	if h == nil {
		t.Fatal("hovering over a funct said nothing")
	}
	if !strings.Contains(h.Contents.Value, "funct") {
		t.Errorf("hovering over a funct should say it is a funct, got %q", h.Contents.Value)
	}
	if h.Range == nil {
		t.Error("hover should say which word it is about, so the editor can mark it")
	}
}

func TestHoveringABuiltinSaysWhatItDoes(t *testing.T) {
	_, c := newClient(t)
	c.open("chap_khat 1\n")
	h := c.hover(0, 2)
	if h == nil {
		t.Fatal("hovering over a builtin said nothing")
	}
	if !strings.Contains(h.Contents.Value, "write") {
		t.Errorf("hovering over chap_khat should say it writes something out, got %q", h.Contents.Value)
	}
}

func TestHoveringAKeywordSaysWhatGroupItIsIn(t *testing.T) {
	_, c := newClient(t)
	c.open("har i az 0 ta 3\n    chap_khat i\n")
	h := c.hover(0, 1)
	if h == nil {
		t.Fatal("hovering over a word of the grammar said nothing")
	}
	if !strings.Contains(h.Contents.Value, "har") {
		t.Errorf("hovering over har should name it, got %q", h.Contents.Value)
	}
	if !strings.Contains(h.Contents.Value, "loop") {
		t.Errorf("hovering over har should say what it is for, got %q", h.Contents.Value)
	}
}

func TestTheOutlineListsWhatAReaderScrollsPast(t *testing.T) {
	_, c := newClient(t)
	text := "darghan N\n    x = 0\n\n    chap_kard()\n        chap_khat 1\n\n" +
		"funct f()\n    chap_khat 1\n\nshenase Rang\n    zaban = 0\n\n" +
		"dastoor Drawable\n    chap_kard\n\nn = N()\nn.chap_kard()\nf()\n"
	c.open(text)
	syms := c.symbols()
	var names []string
	for _, s := range syms {
		names = append(names, s.Name)
	}
	want := []string{"N", "chap_kard", "f", "Rang", "Drawable"}
	if len(names) != len(want) {
		t.Fatalf("the outline lists %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("the outline lists %v, want %v", names, want)
		}
	}

	for _, s := range syms {
		if s.Name == "chap_kard" && s.Kind != SymbolMethod {
			t.Errorf("chap_kard is a method, but the outline says kind %d", s.Kind)
		}
	}
}

func TestTheEditorFormatsTheWayTheCommandDoes(t *testing.T) {
	_, c := newClient(t)
	messy := "darghan N\n  x = 0\n  __sakht__(x)\n    khod.x = x\nn = N(1)\n"
	c.open(messy)
	edits := c.format(messy)
	if len(edits) != 1 {
		t.Fatalf("tidying a file should be one change, not %d", len(edits))
	}
	bag := &diag.Bag{}
	tree := parser.Parse("t.mhn", []byte(messy), bag)
	if bag.HasErrors() {
		t.Fatalf("the file did not read, so the test cannot compare anything:\n%s",
			bag.Sorted()[0].Render([]byte(messy)))
	}
	if want := format.File(tree, []byte(messy)); edits[0].NewText != want {
		t.Errorf("the editor tidied it to:\n%q\nbut `mihan fmt` gives:\n%q", edits[0].NewText, want)
	}

	tidy := edits[0].NewText
	again := c.format(tidy)
	switch {
	case len(again) == 0:

	case len(again) == 1 && again[0].NewText == tidy:

		t.Log("the editor sent an edit that changed nothing for an already-tidy file")
	default:
		t.Errorf("tidying an already-tidy file changed it to %q, which means the two will fight on every save",
			again[0].NewText)
	}
}

func TestAnUnreadableFileIsNotTouched(t *testing.T) {
	_, c := newClient(t)
	broken := "darghan N\n    x = \n"
	c.open(broken)
	if edits := c.format(broken); edits != nil {
		t.Errorf("a file that did not read must be left alone, but it was rewritten to %q", edits[0].NewText)
	}
}

func TestAFileThatIsAlreadyTidyIsLeftAlone(t *testing.T) {
	_, c := newClient(t)
	tidy := "x = 1\nchap_khat x\n"
	c.open(tidy)
	if edits := c.format(tidy); edits != nil {
		t.Errorf("a file with nothing to tidy should not be rewritten, but it became %q", edits[0].NewText)
	}
}

func TestPersianTextDoesNotShiftWhereAMistakeIs(t *testing.T) {
	_, c := newClient(t)

	text := "# یک توضیح فارسی\nx = 1\nchap_khat x\n"
	c.open(text)

	places := c.definition(2, 10)
	if len(places) == 0 {
		t.Fatal("the read of x found nothing, even with Persian text above it")
	}

	if places[0].Range.Start.Line != 1 {
		t.Errorf("x is written on the second line of the file, but the jump points at line %d", places[0].Range.Start.Line)
	}
	if places[0].Range.Start.Character != 0 {
		t.Errorf("x is written at the first character, but the jump points at %d", places[0].Range.Start.Character)
	}
}

func TestSuggestionsDoNotOfferNamesFromOutOfScope(t *testing.T) {
	_, c := newClient(t)
	text := "shanje_man = 1\nfunct f(a)\n    b = 2\n    \n"
	c.open(text)

	have := labels(c.complete(3, 4))
	if _, ok := have["a"]; !ok {
		t.Error("a is in scope inside f, and should be on offer")
	}
	if _, ok := have["b"]; !ok {
		t.Error("b is in scope inside f, and should be on offer")
	}
	if _, ok := have["shanje_man"]; !ok {
		t.Error("shanje_man is written in the file and is visible here, so it should be on offer")
	}

	c.open("shanje_man = 1\nfunct f(a)\n    b = 2\n    chap_khat b\n")
	top := labels(c.complete(0, 6))
	if _, ok := top["a"]; ok {
		t.Error("a belongs to f and is not visible at the top, so it should not be on offer")
	}
	if _, ok := top["b"]; ok {
		t.Error("b belongs to f and is not visible at the top, so it should not be on offer")
	}
}
