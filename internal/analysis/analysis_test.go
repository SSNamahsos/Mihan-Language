package analysis

import (
	"sort"
	"strings"
	"testing"

	"mihan/internal/diag"
	"mihan/internal/parser"
)

func (a *File) declByName(name string) *Decl {
	for _, d := range a.Decls {
		if d.Name == name {
			return d
		}
	}
	return nil
}

func build(t *testing.T, src string) *File {
	t.Helper()

	Builtins["chap_khat"] = "chap_khat x — write a value out, then start a new line"
	Builtins["tule"] = "tule x — how many items are in x"
	Builtins["taghib"] = "taghib xs — the list, sorted"
	Builtins["neshan"] = "neshan x — the readable form of a value, as matn"
	bag := &diag.Bag{}
	f := parser.Parse("t.mhn", []byte(src), bag)
	if bag.HasErrors() {
		t.Fatalf("the program did not read:\n%s\n  %s", src, bag.Sorted()[0].Render([]byte(src)))
	}
	return Build(f)
}

func TestNamesAreFoundWhereTheyAreWritten(t *testing.T) {
	cases := []struct {
		what  string
		src   string
		name  string
		kind  DeclKind
		pos   diag.Pos
		depth int
	}{
		{"a variable at the top", "x = 1\n", "x", DeclVar, diag.Pos{Line: 1, Col: 1, Off: 0}, 0},
		{"a parameter", "funct f(a, b)\n    bargard a\nf(1, 2)\n", "a", DeclParam, diag.Pos{Line: 1, Col: 9, Off: 8}, 1},
		{"a funct", "funct f()\n    chap_khat 1\nf()\n", "f", DeclFunc, diag.Pos{Line: 1, Col: 7, Off: 6}, 0},
		{"a darghan", "darghan N\n    x = 0\nn = N()\n", "N", DeclClass, diag.Pos{Line: 1, Col: 9, Off: 8}, 0},
		{"a shenase", "shenase Rang\n    zaban = 0\n", "Rang", DeclEnum, diag.Pos{Line: 1, Col: 9, Off: 8}, 0},
		{"a loop variable", "har i az 0 ta 3\n    chap_khat i\n", "i", DeclFor, diag.Pos{Line: 1, Col: 5, Off: 4}, 1},
	}
	for _, c := range cases {
		t.Run(c.what, func(t *testing.T) {
			a := build(t, c.src)
			d := a.declByName(c.name)
			if d == nil {
				var have []string
				for _, n := range a.Root.Names() {
					have = append(have, n)
				}
				t.Fatalf("%q is not in the file, which holds %v", c.name, have)
			}
			if d.Kind != c.kind {
				t.Errorf("%q is a %s, but the walk says it is a %s", c.name, c.kind, d.Kind)
			}
			if d.Pos.Off != c.pos.Off {
				t.Errorf("%q is written at offset %d, but the walk says %d (line %d col %d)",
					c.name, c.pos.Off, d.Pos.Off, d.Pos.Line, d.Pos.Col)
			}
		})
	}
}

func TestANameFindsTheOneThatIsNearest(t *testing.T) {
	a := build(t, "x = 1\nfunct f()\n    x = 2\n    chap_khat x\nf()\n")

	var read *Use
	for _, u := range a.Uses {
		if u.Name == "x" {
			read = u
		}
	}
	if read == nil {
		t.Fatal("the file never reads x")
	}
	if read.Pos.Line != 4 {
		t.Fatalf("the read of x should be on line 4, but the walk says line %d:\n%s", read.Pos.Line, dump(a))
	}
	if read.Decl == nil {
		t.Fatalf("the read of x on line 4 finds nothing:\n%s", dump(a))
	}
	if read.Decl.Pos.Line != 3 {
		t.Errorf("the read of x on line 4 means the x on line 3, but the walk says line %d:\n%s",
			read.Decl.Pos.Line, dump(a))
	}
}

func TestEveryReadFindsSomething(t *testing.T) {
	srcs := []string{
		"x = 1\nchap_khat x\n",
		"funct f(a)\n    bargard a * 2\nchap_khat f(3)\n",
		"n = 0\nhar i az 0 ta 3\n    n = n + i\nchap_khat n\n",
		"darghan N\n    x = 0\n    __sakht__(x)\n        khod.x = x\n    funct kard()\n        chap_khat khod.x\nn = N(1)\nn.kard()\n",
		"xs = [1, 2, 3]\nhar x dar xs\n    chap_khat x\n",
		"koshish\n    chap_khat 1\ngiroft khata e\n    chap_khat e.message\n",
	}
	for _, src := range srcs {
		a := build(t, src)
		for _, u := range a.Uses {
			if u.Decl == nil && u.Builtin == "" {
				t.Errorf("%q is read on line %d but nothing in the file gives it a value:\n%s",
					u.Name, u.Pos.Line, dump(a))
			}
		}
	}
}

func TestTheOutlineListsWhatARepiderLooksFor(t *testing.T) {
	a := build(t, "darghan N\n    x = 0\n\n    funct kard()\n        chap_khat 1\n\nfunct f()\n    chap_khat 1\n\nshenase Rang\n    zaban = 0\n\ndastoor Drawable\n    chap_kard\n\nn = N()\nn.kard()\nf()\n")
	var names []string
	for _, s := range a.Symbols() {
		names = append(names, s.Kind.String()+" "+s.Name)
	}

	want := []string{"darghan N", "method kard", "funct f", "shenase Rang", "dastoor Drawable"}
	if len(names) != len(want) {
		t.Fatalf("the outline has %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("the outline has %v, want %v", names, want)
			break
		}
	}
}

func posAt(src string, line, col int) diag.Pos {
	off := 0
	l := 1
	for l < line {
		nl := strings.IndexByte(src[off:], '\n')
		if nl < 0 {
			break
		}
		off += nl + 1
		l++
	}
	return diag.Pos{Line: l, Col: col, Off: off + col - 1}
}

func TestNamesOffersWhatIsVisible(t *testing.T) {
	src := "g = 10\nfunct f(a)\n    b = a + 1\n    bargard b\nf(1)\n"
	a := build(t, src)

	s := a.ScopeAt(posAt(src, 4, 12))
	if s == a.Root {
		t.Errorf("line 4 is inside f, so the scope there should be f's, not the file's")
	}
	names := map[string]bool{}
	for _, n := range s.Names() {
		names[n] = true
	}
	for _, want := range []string{"a", "b", "g", "f"} {
		if !names[want] {
			t.Errorf("from inside f, %q should be on offer, but only %v are", want, keysOf(names))
		}
	}

	top := map[string]bool{}
	for _, n := range a.ScopeAt(posAt(src, 1, 1)).Names() {
		top[n] = true
	}
	if top["a"] || top["b"] {
		t.Errorf("from the top of the file, the names inside f should not be on offer, but %v are", keysOf(top))
	}
}

func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func dump(a *File) string {
	out := "decls:"
	for _, d := range a.Decls {
		out += "\n  " + d.Kind.String() + " " + d.Name + " at line " + itoa(d.Pos.Line)
	}
	out += "\nuses:"
	for _, u := range a.Uses {
		out += "\n  " + u.Name + " at line " + itoa(u.Pos.Line)
		if u.Decl != nil {
			out += " -> line " + itoa(u.Decl.Pos.Line)
		} else {
			out += " -> nothing"
		}
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}
