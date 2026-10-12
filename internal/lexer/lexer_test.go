package lexer

import (
	"strings"
	"testing"

	"mihan/internal/diag"
	"mihan/internal/token"
)

func kinds(toks []token.Token) string {
	var b strings.Builder
	for i, tk := range toks {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(string(tk.Kind))
		if showLit[tk.Kind] {
			b.WriteString("(" + tk.Lit + ")")
		}
	}
	return b.String()
}

var showLit = map[token.Kind]bool{
	token.IDENT: true,
	token.INT:   true,
	token.FLOAT: true,
}

func lexAll(t *testing.T, src string) ([]token.Token, *diag.Bag) {
	t.Helper()
	bag := &diag.Bag{}
	lx := New("test.mhn", []byte(src), bag)
	return lx.Tokens(), bag
}

func TestBasicKinds(t *testing.T) {
	toks, bag := lexAll(t, "naam = 42\n")
	if bag.HasErrors() {
		t.Fatalf("unexpected errors: %v", bag.Sorted()[0].Message())
	}
	got := kinds(toks)
	want := "ident(naam) = int(42) newline eof"
	if got != want {
		t.Fatalf("kinds\n got: %s\nwant: %s", got, want)
	}
}

func TestIndentationEmitsIndentDedent(t *testing.T) {
	src := "agar shefte\n    chap 1\n    chap 2\n"
	toks, bag := lexAll(t, src)
	if bag.HasErrors() {
		t.Fatalf("unexpected errors: %v", bag.Sorted()[0].Message())
	}
	got := kinds(toks)
	want := "agar ident(shefte) newline indent chap int(1) newline chap int(2) newline dedent eof"
	if got != want {
		t.Fatalf("kinds\n got: %s\nwant: %s", got, want)
	}
}

func TestNestedIndentation(t *testing.T) {
	src := "har i az 0 ta 2\n    agar i\n        chap i\n"
	toks, _ := lexAll(t, src)
	got := kinds(toks)
	want := "har ident(i) az int(0) ta int(2) newline indent agar ident(i) newline indent chap ident(i) newline dedent dedent eof"
	if got != want {
		t.Fatalf("kinds\n got: %s\nwant: %s", got, want)
	}
}

func TestAliasesResolveToCanonicalKind(t *testing.T) {
	toks, _ := lexAll(t, "if x\n    print 1\n")
	got := kinds(toks)
	want := "agar ident(x) newline indent chap int(1) newline dedent eof"
	if got != want {
		t.Fatalf("kinds\n got: %s\nwant: %s", got, want)
	}
}

func TestBracesSuspendIndentation(t *testing.T) {

	src := "funct f() {\nchap 1\nchap 2\n}\n"
	toks, bag := lexAll(t, src)
	if bag.HasErrors() {
		t.Fatalf("unexpected errors: %v", bag.Sorted()[0].Message())
	}
	got := kinds(toks)
	want := "funct ident(f) ( ) { newline chap int(1) newline chap int(2) newline } newline eof"
	if got != want {
		t.Fatalf("kinds\n got: %s\nwant: %s", got, want)
	}
}

func TestParensSuppressNewlines(t *testing.T) {
	src := "chap(\n  1,\n  2\n)\n"
	toks, bag := lexAll(t, src)
	if bag.HasErrors() {
		t.Fatalf("unexpected errors: %v", bag.Sorted()[0].Message())
	}
	if strings.Contains(kinds(toks), "newline )") {
		t.Fatalf("newlines inside parens should vanish: %s", kinds(toks))
	}
	if strings.Contains(kinds(toks), "indent") || strings.Contains(kinds(toks), "dedent") {
		t.Fatalf("parens must not produce indentation: %s", kinds(toks))
	}
}

func TestStringInterpolation(t *testing.T) {
	toks, bag := lexAll(t, `chap $"salam ${naam + 1}!"`+"\n")
	if bag.HasErrors() {
		t.Fatalf("unexpected errors: %v", bag.Sorted()[0].Message())
	}
	got := kinds(toks)
	want := "chap interp interp_text ident(naam) + int(1) interp_end interp_text string_end newline eof"
	if got != want {
		t.Fatalf("kinds\n got: %s\nwant: %s", got, want)
	}
}

func TestInterpolationKeepsRealPositions(t *testing.T) {

	src := "chap 1\nx = $\"a ${ naam } b\"\n"
	toks, bag := lexAll(t, src)
	if bag.HasErrors() {
		t.Fatalf("unexpected errors: %v", bag.Sorted()[0].Message())
	}
	found := false
	for _, tk := range toks {
		if tk.Lit == "naam" {
			found = true
			if tk.Line != 2 {
				t.Errorf("naam on line %d, want 2", tk.Line)
			}
			if tk.Col != 12 {
				t.Errorf("naam at column %d, want 12", tk.Col)
			}
		}
	}
	if !found {
		t.Fatalf("the hole's name never reached the stream: %s", kinds(toks))
	}
}

func TestInterpolationWithDictInsideAHole(t *testing.T) {
	toks, bag := lexAll(t, `$"${ {"a": 1} }"`+"\n")
	if bag.HasErrors() {
		t.Fatalf("a dict inside a hole must parse: %v", bag.Sorted()[0].Message())
	}

	closes := 0
	for _, tk := range toks {
		if tk.Kind == token.INTERP_END {
			closes++
		}
	}
	if closes != 1 {
		t.Fatalf("want exactly one hole terminator, got %d: %s", closes, kinds(toks))
	}
}

func TestShortDollarInterpolation(t *testing.T) {
	toks, _ := lexAll(t, `x = "hi $user!"`+"\n")
	got := kinds(toks)
	want := "ident(x) = interp interp_text ident(user) interp_end interp_text string_end newline eof"
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestPlainStringWithNoDollarIsJustAString(t *testing.T) {

	toks, _ := lexAll(t, `x = "salam $ oops"`+"\n")
	for _, tk := range toks {
		if tk.Kind == token.INTERP {
			t.Fatalf("a lone $ that is not followed by a name must stay literal")
		}
	}
	var s string
	for _, tk := range toks {
		if tk.Kind == token.STRING {
			s = tk.Text
		}
	}
	if s != "salam $ oops" {
		t.Fatalf("got %q", s)
	}
}

func TestEscapes(t *testing.T) {
	cases := map[string]string{
		`"a\nb"`:        "a\nb",
		`"tab\there"`:   "tab\there",
		`"q\"q"`:        `q"q`,
		`"back\\slash"`: `back\slash`,
		`"dollar\$x"`:   "dollar$x",
		`"\x41"`:        "A",
		`"\u06A9"`:      "\u06a9",
		"'x'":           "x",
		"'\\n'":         "\n",
		"raw`a\\nb`":    `a\nb`,
		"'\\\\'":        "\\",
		"'\\u06A9'":     "\u06a9",
	}
	for src, want := range cases {
		toks, bag := lexAll(t, src+"\n")
		if bag.HasErrors() {
			t.Errorf("%s: unexpected error %s", src, bag.Sorted()[0].Message())
			continue
		}
		var got string
		for _, tk := range toks {
			if tk.Kind == token.STRING || tk.Kind == token.CHAR {
				got = tk.Text
			}
		}
		if got != want {
			t.Errorf("%s: got %q, want %q", src, got, want)
		}
	}
}

func TestNumberForms(t *testing.T) {
	cases := map[string]string{
		"42":        "int",
		"1_000_000": "int",
		"0xFF":      "int",
		"0b1011":    "int",
		"0o17":      "int",
		"3.5":       "float",
		"1e3":       "float",
		"2.5e-3":    "float",
		"1_0.5_0":   "float",
	}
	for src, kind := range cases {
		toks, bag := lexAll(t, src+"\n")
		if bag.HasErrors() {
			t.Errorf("%s: unexpected error %s", src, bag.Sorted()[0].Message())
			continue
		}
		if string(toks[0].Kind) != kind {
			t.Errorf("%s: kind %s, want %s", src, toks[0].Kind, kind)
		}
	}
}

func TestUnicodeIdentifiers(t *testing.T) {
	toks, bag := lexAll(t, "سلام = 1\n")
	if bag.HasErrors() {
		t.Fatalf("unexpected error: %s", bag.Sorted()[0].Message())
	}
	if toks[0].Kind != token.IDENT || toks[0].Lit != "سلام" {
		t.Fatalf("got %+v", toks[0])
	}
}

func TestCommentsAreCollectedNotEmitted(t *testing.T) {
	src := "# a comment\nchap 1  // another\n/* block */\n"
	bag := &diag.Bag{}
	lx := New("test.mhn", []byte(src), bag)
	toks := lx.Tokens()
	if bag.HasErrors() {
		t.Fatalf("unexpected error: %s", bag.Sorted()[0].Message())
	}
	for _, tk := range toks {
		if tk.Kind == token.COMMENT {
			t.Fatalf("comments must not enter the token stream")
		}
	}
	if len(lx.Comments()) != 3 {
		t.Fatalf("want 3 comments, got %d: %+v", len(lx.Comments()), lx.Comments())
	}
}

func TestUnclosedStringIsReported(t *testing.T) {
	_, bag := lexAll(t, "x = \"oops\n")
	if !bag.HasErrors() {
		t.Fatal("an unclosed string must be an error")
	}
	if code := bag.Sorted()[0].Code; code != "E1005" {
		t.Fatalf("code = %s, want E1005", code)
	}
}

func TestMisalignedIndentationIsReported(t *testing.T) {

	_, bag := lexAll(t, "agar x\n        chap 1\n    chap 2\n")
	if !bag.HasErrors() {
		t.Fatal("inconsistent indentation must be an error")
	}
	if code := bag.Sorted()[0].Code; code != "E1002" {
		t.Fatalf("code = %s, want E1002", code)
	}
}

func TestUnmatchedBraceIsReported(t *testing.T) {
	_, bag := lexAll(t, "chap 1)\n")
	if !bag.HasErrors() {
		t.Fatal("an unmatched ) must be an error")
	}
}

func TestEmptyInput(t *testing.T) {
	toks, bag := lexAll(t, "")
	if bag.HasErrors() {
		t.Fatalf("empty input must not error")
	}
	if len(toks) != 1 || toks[0].Kind != token.EOF {
		t.Fatalf("got %s", kinds(toks))
	}
}

func TestBlankLinesAreIgnored(t *testing.T) {
	toks, bag := lexAll(t, "chap 1\n\n\n   \nchap 2\n")
	if bag.HasErrors() {
		t.Fatalf("blank lines must not error: %s", bag.Sorted()[0].Message())
	}
	if got, want := kinds(toks), "chap int(1) newline chap int(2) newline eof"; got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestSemicolonsSeparateStatements(t *testing.T) {
	toks, _ := lexAll(t, "chap 1; chap 2\n")
	if got, want := kinds(toks), "chap int(1) ; chap int(2) newline eof"; got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestEveryKeywordAndAliasLexes(t *testing.T) {
	for _, kw := range token.KeywordList() {
		for _, spelling := range append([]string{kw.Name}, kw.Aliases...) {
			toks, bag := lexAll(t, spelling+"\n")
			if bag.HasErrors() {
				t.Errorf("%q: %s", spelling, bag.Sorted()[0].Message())
				continue
			}
			if toks[0].Kind != token.Kind(kw.Name) {
				t.Errorf("%q lexed as %s, want %s", spelling, toks[0].Kind, kw.Name)
			}
			if toks[0].Lit != spelling {
				t.Errorf("%q kept literal %q", spelling, toks[0].Lit)
			}
		}
	}
}

func TestEveryOperatorIsLongestMatched(t *testing.T) {
	cases := map[string]string{
		"a = b":        "=",
		"a += b":       "+=",
		"a **= b":      "**=",
		"a ** b":       "**",
		"a ghesmat b":  "ghesmat",
		"a floordiv b": "ghesmat",
		"a < b":        "<",
		"a <= b":       "<=",
		"a <=> b":      "<=>",
		"a << b":       "<<",
		"a >>= b":      ">>=",
		"a .. b":       "..",
		"a ..+ b":      "..+",
		"a -> b":       "->",
		"a => b":       "=>",
		"a ? b":        "?",
		"a :: b":       "::",
		"a && b":       "&&",
		"a and b":      "va",
	}
	for src, want := range cases {
		toks, _ := lexAll(t, src+"\n")
		if got := string(toks[1].Kind); got != want {
			t.Errorf("%q: got %q want %q", src, got, want)
		}
	}
}

func TestFuzzNeverPanics(t *testing.T) {
	seeds := []string{
		"agar", "chap", "$", "${", "\"", "`", "'", "..", "0x", "\t", "\n\n",
		"سلام", "#", "/*", "a.b", "f(x", "}", "]", ")",
	}
	for _, s := range seeds {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic on %q: %v", s, r)
				}
			}()
			lexAll(t, s)
		}()
	}
}
