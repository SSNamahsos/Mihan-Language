package vm

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"mihan/internal/diag"
	"mihan/internal/eval"
	"mihan/internal/parser"

	"mihan/internal/ir"
	rt "mihan/internal/runtime"
)

func byInterpreter(t *testing.T, src string) string {
	t.Helper()
	bag := &diag.Bag{}
	file := parser.Parse("t.mhn", []byte(src), bag)
	if bag.HasErrors() {
		t.Fatalf("the program did not read:\n%s", src)
	}
	var out bytes.Buffer
	in := eval.New(&out, nil)
	if err := in.RunFile(file); err != nil {
		t.Fatalf("the program stopped: %v", err)
	}
	return out.String()
}

func sameAnswer(t *testing.T, expr string, a, b Value) {
	t.Helper()
	got, f := arith(opOf(expr), a, b)
	if f != nil {
		t.Fatalf("%s: the machine said %s", expr, f.en)
	}
	want := byInterpreter(t, "chap_khat("+expr+")")
	if want != got.Text()+"\n" {
		t.Errorf("%s: the machine said %q, the language said %q", expr, got.Text(), strings.TrimSpace(want))
	}
}

func opOf(expr string) ir.Op {
	for _, p := range []struct {
		text string
		op   ir.Op
	}{
		{" + ", ir.OpAdd}, {" - ", ir.OpSub}, {" * ", ir.OpMul}, {" / ", ir.OpDiv},
		{" % ", ir.OpRem}, {" ** ", ir.OpPow},
	} {
		if strings.Contains(expr, p.text) {
			return p.op
		}
	}
	return ir.OpIDiv
}

func TestWholeNumberArithmeticAgreesWithTheLanguage(t *testing.T) {
	nums := []int64{0, 1, 2, 7, -3, -1, 100, 1000000, -1000000}
	for _, op := range []ir.Op{ir.OpAdd, ir.OpSub, ir.OpMul, ir.OpIDiv, ir.OpRem} {
		sym := map[ir.Op]string{ir.OpAdd: "+", ir.OpSub: "-", ir.OpMul: "*", ir.OpIDiv: "ghesmat", ir.OpRem: "%"}[op]
		for _, a := range nums {
			for _, b := range nums {
				if op != ir.OpAdd && op != ir.OpSub && b == 0 {
					continue
				}
				expr := fmt.Sprintf("%d %s %d", a, sym, b)
				sameAnswer(t, expr, Int(a), Int(b))
			}
		}
	}
}

func TestFractionsAgreeWithTheLanguage(t *testing.T) {

	for _, c := range [][3]any{
		{"7 / 2", 7.0, 2.0}, {"6 / 3", 6.0, 3.0}, {"-7 / 2", -7.0, 2.0},
		{"1 / 3", 1.0, 3.0}, {"0 / 5", 0.0, 5.0},
		{"ghesmat(7, 2)", 7.0, 2.0}, {"ghesmat(-7, 2)", -7.0, 2.0}, {"ghesmat(7, -2)", 7.0, -2.0},
		{"7.5 % 2", 7.5, 2.0}, {"2 ** 10", 2.0, 10.0},
	} {
		expr := c[0].(string)
		op := opOf(expr)
		if op == ir.OpIDiv && strings.HasPrefix(expr, "ghesmat") {
			op = ir.OpIDiv
		}
		sameAnswer(t, expr, Float(c[1].(float64)), Float(c[2].(float64)))
	}
}

func TestAWholeNumberStaysWhole(t *testing.T) {
	v, f := arith(ir.OpAdd, Int(2), Int(3))
	if f != nil {
		t.Fatal(f.en)
	}
	if v.TagOf() != TagInt {
		t.Errorf("2 + 3 is a whole number, so the machine should not have made it a %s", v.TypeName())
	}
	v, f = arith(ir.OpDiv, Int(7), Int(2))
	if f != nil {
		t.Fatal(f.en)
	}
	if v.TagOf() != TagFloat {
		t.Errorf("7 / 2 is 3.5, so the machine should have made it a %s", v.TypeName())
	}
	if v.Float() != 3.5 {
		t.Errorf("7 / 2 gave %v, want 3.5", v.Float())
	}
}

func TestDividingByZeroIsAMistakeNotACrash(t *testing.T) {
	for _, op := range []ir.Op{ir.OpDiv, ir.OpIDiv, ir.OpRem} {
		if _, f := arith(op, Int(1), Int(0)); f == nil {
			t.Errorf("%s by zero should be a mistake", op)
		}
	}
}

func TestAddingTextJoinsIt(t *testing.T) {
	v, f := arith(ir.OpAdd, Str("sa"), Str("lam"))
	if f != nil {
		t.Fatal(f.en)
	}
	if v.s != "salam" {
		t.Errorf(`"sa" + "lam" gave %q, want "salam"`, v.s)
	}
}

func TestComparisonAgreesWithTheLanguage(t *testing.T) {
	pairs := [][2]int64{{1, 2}, {2, 1}, {1, 1}, {-1, 1}, {0, 0}, {5, -5}}
	for _, p := range pairs {
		a, b := Int(p[0]), Int(p[1])
		got, ok := compare(a, b)
		if !ok {
			t.Fatalf("%d against %d: the machine could not put them in order", p[0], p[1])
		}
		want, ok := rt.Compare(a.ToRT(), b.ToRT())
		if !ok {
			t.Fatalf("%d against %d: the interpreter could not put them in order", p[0], p[1])
		}
		if got != want {
			t.Errorf("%d against %d: machine says %d, interpreter says %d", p[0], p[1], got, want)
		}
	}
}

func TestTextsAreOrderedByTheirCharacters(t *testing.T) {
	got, ok := compare(Str("salam"), Str("salt"))
	if !ok || got >= 0 {
		t.Errorf(`"salam" comes before "salt", got %d (ok=%v)`, got, ok)
	}
	got, ok = compare(Str("ab"), Str("abc"))
	if !ok || got >= 0 {
		t.Errorf(`"ab" comes before "abc", got %d (ok=%v)`, got, ok)
	}
}

func TestEqualityFollowsTheLanguageNotTheGoType(t *testing.T) {

	if !equal(Int(1), Float(1)) {
		t.Error("1 and 1.0 are the same number")
	}
	if equal(Int(1), Int(2)) {
		t.Error("1 and 2 are not the same number")
	}
	if equal(Bool(true), Int(1)) {
		t.Error("a truth is not a number, so true and 1 are not the same thing")
	}
	if !equal(Nil(), Nil()) {
		t.Error("two absences are the same absence")
	}

	a := List(rt.NewList(rt.Int(1), rt.Int(2)))
	b := List(rt.NewList(rt.Int(1), rt.Int(2)))
	if !equal(a, b) {
		t.Error("two lists with the same items are the same list")
	}
	c := List(rt.NewList(rt.Int(2), rt.Int(1)))
	if equal(a, c) {
		t.Error("a list and the same items the other way round are not the same list")
	}
}

func TestValuesCrossToTheInterpreterWithoutLosingAnything(t *testing.T) {
	cases := []Value{
		Int(42), Int(-1), Float(3.5), Str("salam"), Bool(true), Bool(false), Nil(),
	}
	for _, v := range cases {
		back := FromRT(v.ToRT())
		if back.TagOf() != v.TagOf() {
			t.Errorf("%s became a %s on the way across", v.TypeName(), back.TypeName())
		}
		if back.Text() != v.Text() {
			t.Errorf("%s printed as %q after a round trip, want %q", v.TypeName(), back.Text(), v.Text())
		}
	}
}
