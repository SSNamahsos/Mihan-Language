package ir

import (
	"strings"
	"testing"

	"mihan/internal/diag"
	"mihan/internal/parser"
)

func lower(t *testing.T, src string) *Program {
	t.Helper()
	bag := &diag.Bag{}
	f := parser.Parse("t.mhn", []byte(src), bag)
	if bag.HasErrors() {
		t.Fatalf("the program did not read:\n%s\n  %s", src, bag.Sorted()[0].Render([]byte(src)))
	}
	return Lower(f, bag)
}

func ops(p *Program, fnIdx int) []string {
	var out []string
	for _, in := range p.Funcs[fnIdx].Code {
		out = append(out, in.Op.String())
	}
	return out
}

func has(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func countOf(list []string, want string) int {
	n := 0
	for _, s := range list {
		if s == want {
			n++
		}
	}
	return n
}

func findFunc(p *Program, name string) int {
	for i, f := range p.Funcs {
		if f.Name == name {
			return i
		}
	}
	return -1
}

func TestNamesBecomeNumbers(t *testing.T) {
	t.Run("a name in a body is a slot in the frame", func(t *testing.T) {
		p := lower(t, "funct f()\n    x = 1\n    bargard x + 1\nf()")
		fn := p.Funcs[findFunc(p, "f")]
		if fn.Locals != 1 {
			t.Errorf("the body declared one name, so the frame wants one slot, not %d", fn.Locals)
		}
		if fn.Params != 0 {
			t.Errorf("f takes nothing, so it wants 0 parameters, not %d", fn.Params)
		}

		var reads, writes int
		for _, in := range fn.Code {
			switch in.Op {
			case OpLocalGet:
				if in.A == 0 {
					reads++
				}
			case OpLocalSet:
				if in.A == 0 {
					writes++
				}
			}
		}
		if reads == 0 || writes == 0 {
			t.Errorf("both a read and a write of x should name slot 0, got %d reads and %d writes:\n%s",
				reads, writes, Disassemble(p))
		}
	})

	t.Run("a parameter is a slot too", func(t *testing.T) {
		p := lower(t, "funct f(a, b)\n    bargard a + b\nf(1, 2)")
		fn := p.Funcs[findFunc(p, "f")]
		if fn.Params != 2 {
			t.Errorf("f takes two things, so it wants 2 parameters, not %d", fn.Params)
		}
		if fn.Locals != 2 {
			t.Errorf("the two parameters are the two slots, so it wants 2 slots, not %d", fn.Locals)
		}
	})

	t.Run("a name from further out is captured, not copied", func(t *testing.T) {

		p := lower(t, "funct shiran()\n    x = 1\n    funct dakheli()\n        bargard x\n    bargard dakheli\nshiran()")
		inner := p.Funcs[findFunc(p, "dakheli")]
		if inner.Upvals == 0 {
			t.Errorf("dakheli reads x from outside itself, so it must capture it:\n%s", Disassemble(p))
		}
		if inner.Locals != 0 {
			t.Errorf("x belongs to shiran, so dakheli must not give it a slot of its own; it has %d", inner.Locals)
		}
		if !has(ops(p, findFunc(p, "dakheli")), "UpvalueGet") {
			t.Errorf("dakheli must read x as an upvalue:\n%s", Disassemble(p))
		}
	})
}

func TestALoopBecomesOneRangeAndTwoJumps(t *testing.T) {
	p := lower(t, "n = 0\nhar i az 0 ta 10\n    n = n + i\nchap_khat n")
	got := ops(p, 0)
	if countOf(got, "MakeRange") != 1 {
		t.Errorf("a loop needs one range, not %d:\n%s", countOf(got, "MakeRange"), Disassemble(p))
	}
	if countOf(got, "Jump") < 1 {
		t.Errorf("a loop needs a jump back to the top:\n%s", Disassemble(p))
	}
	if !has(got, "JumpIfFalse") {
		t.Errorf("a loop needs a way out when the run is finished:\n%s", Disassemble(p))
	}

	checkJumps(t, p)
}

func checkJumps(t *testing.T, p *Program) {
	t.Helper()
	for i, f := range p.Funcs {
		for pc, in := range f.Code {
			switch in.Op {
			case OpJump, OpJumpIfFalse, OpJumpIfTrue, OpJumpIfNotNil:
				if in.A < 0 || int(in.A) > len(f.Code) {
					t.Errorf("%s: %s at %d goes to %d, which is outside the %d instructions of the function",
						f.Name, in.Op, pc, in.A, len(f.Code))
				}
			}
		}
		_ = i
	}
}

func TestEveryJumpInEveryProgramLandsSomewhereReal(t *testing.T) {
	programs := []string{
		"n = 0\nhar i az 0 ta 10\n    n = n + i\nchap_khat n",
		"i = 0\nta i < 10\n    i = i + 1\n    agar i == 5\n        shek\n",
		"xs = [1, 2, 3]\nhar x dar xs\n    chap_khat x\n",
		"funct f(n)\n    agar n < 2\n        bargard n\n    bargard f(n - 1) + f(n - 2)\nchap_khat f(10)\n",
		"koshish\n    chap_khat \"a\"\ngiroft khata e\n    chap_khat e.message\nnemigir\n    chap_khat \"done\"\n",
		"darghan N\n    x = 0\n    __sakht__(x)\n        khod.x = x\n    chap_kon()\n        chap_khat khod.x\nn = N(7)\nn.chap_kon()\n",
		"agar dorost\n    chap_khat \"a\"\nwarna\n    chap_khat \"b\"\n",
		"x = 5\nagar x > 3 va x < 10\n    chap_khat \"dar\"\nwarna agar x == 1\n    chap_khat \"yek\"\nwarna\n    chap_khat \"digar\"\n",
	}
	for i, src := range programs {
		p := lower(t, src)
		checkJumps(t, p)
		for _, f := range p.Funcs {
			for pc, in := range f.Code {
				if in.Op != OpInvalid && !in.Op.valid() {
					t.Errorf("program %d: %s has an operation that is not real", i, f.Name)
				}
				_ = pc
			}
		}

		lines := strings.Count(src, "\n") + 1
		for _, f := range p.Funcs {
			for pc, in := range f.Code {
				if in.Line < 1 || int(in.Line) > lines {
					t.Errorf("program %d: %s instruction %d claims line %d, but the file has %d lines",
						i, f.Name, pc, in.Line, lines)
				}
			}
		}
	}
}

func TestAndSkipsTheRightSideWhenItMust(t *testing.T) {
	p := lower(t, "xs = [1]\nagar dorost va xs[5] == 1\n    chap_khat \"happend\"\nchap_khat \"not\"\n")
	got := Disassemble(p)
	if !strings.Contains(got, "JumpIfFalse") {
		t.Errorf("`va` must be a jump, not an operation with two operands:\n%s", got)
	}
}

func TestOrSkipsTheRightSideWhenItMust(t *testing.T) {
	p := lower(t, "xs = [1]\nagar ghalat ya xs[5] == 1\n    chap_khat \"happend\"\nchap_khat \"not\"\n")
	if !strings.Contains(Disassemble(p), "JumpIfTrue") {
		t.Errorf("`ya` must be a jump:\n%s", Disassemble(p))
	}
}

func TestOperatorWordsLowerToRealInstructions(t *testing.T) {
	cases := []struct {
		src  string
		want string
	}{
		{"x = 7 ghesmat 2", "IDiv"},
		{"x = 2 dar [1, 2]", "In"},
		{"x = na dorost", "Not"},
		{"x = dorost va ghalat", "JumpIfFalse"},
		{"x = dorost ya ghalat", "JumpIfTrue"},
	}
	for _, c := range cases {
		p := lower(t, c.src)
		got := ops(p, 0)
		if !has(got, c.want) {
			t.Errorf("%q should lower to a %s, but lowered to:\n%s",
				c.src, c.want, Disassemble(p))
		}

		if has(got, "Nop") {
			t.Errorf("%q lowered with a Nop in it, which means a word was lost:\n%s",
				c.src, Disassemble(p))
		}
	}
}

func TestAClassBecomesFieldsAndMethods(t *testing.T) {
	p := lower(t, "darghan Noghat\n    x = 0\n    y = 0\n\n    __sakht__(x, y)\n        khod.x = x\n        khod.y = y\n\n    funct kard()\n        chap_khat khod.x\n\nn = Noghat(1, 2)\nn.kard()\n")
	if len(p.Classes) != 1 {
		t.Fatalf("there is one class here, not %d", len(p.Classes))
	}
	c := p.Classes[0]
	if c.Name != "Noghat" {
		t.Errorf("the class is called %q, not Noghat", c.Name)
	}
	if c.FieldCount != 2 {
		t.Errorf("the class has two fields, so FieldCount should be 2, not %d", c.FieldCount)
	}
	if got := p.Names[c.FieldStart]; got != "x" {
		t.Errorf("the first field name is %q, not x", got)
	}
	if got := p.Names[c.FieldStart+1]; got != "y" {
		t.Errorf("the second field name is %q, not y", got)
	}
	if len(c.Methods) != 2 {
		t.Fatalf("the class has two methods, not %d", len(c.Methods))
	}
	if c.CtorFn < 0 {
		t.Error("the class has a __sakht__, so CtorFn should name it")
	}

	ctor := p.Funcs[c.CtorFn]
	if ctor.Params != 3 {
		t.Errorf("__sakht__ takes the object plus x and y, so 3, not %d", ctor.Params)
	}
	for _, m := range c.Methods {
		if m.Name == "chap" && p.Funcs[m.Fn].Params != 1 {
			t.Errorf("a method with no written arguments still takes the object, so 1, not %d",
				p.Funcs[m.Fn].Params)
		}
	}
}

func TestARecordHasNoMethods(t *testing.T) {
	p := lower(t, "sakht Point\n    x = 0\n    y = 0\np = Point(1, 2)\nchap_khat p.x\n")
	if len(p.Classes) != 1 {
		t.Fatalf("there is one record here, not %d", len(p.Classes))
	}
	if !p.Classes[0].IsStruct {
		t.Error("this is a sakht, so the class must say so")
	}
	if len(p.Classes[0].Methods) != 0 {
		t.Errorf("a sakht has no methods, but %d were lowered", len(p.Classes[0].Methods))
	}
}

func TestCompoundAssignmentReadsBeforeItWrites(t *testing.T) {
	p := lower(t, "n = 0\nn += 1\nchap_khat n")

	reads := countOf(ops(p, 0), "GlobalGet") + countOf(ops(p, 0), "LocalGet")
	writes := countOf(ops(p, 0), "GlobalSet") + countOf(ops(p, 0), "LocalSet")
	if reads < 2 {
		t.Errorf("n += 1 must read n as well as chap_khat reading it, got %d reads:\n%s",
			reads, Disassemble(p))
	}
	if !has(ops(p, 0), "Add") {
		t.Errorf("n += 1 must do the addition:\n%s", Disassemble(p))
	}
	if writes < 2 {
		t.Errorf("n += 1 must write n as well as n = 0 writing it, got %d writes:\n%s",
			writes, Disassemble(p))
	}

	seq := ops(p, 0)
	addAt := -1
	for i, o := range seq {
		if o == "Add" {
			addAt = i
			break
		}
	}
	readAt, writeAt := -1, -1
	for i := addAt - 1; i >= 0; i-- {
		if seq[i] == "GlobalGet" || seq[i] == "LocalGet" {
			readAt = i
			break
		}
	}
	for i := addAt + 1; i < len(seq); i++ {
		if seq[i] == "GlobalSet" || seq[i] == "LocalSet" {
			writeAt = i
			break
		}
	}
	if readAt < 0 || writeAt < 0 {
		t.Fatalf("n += 1 must read and write around the addition, got read=%d add=%d write=%d:\n%s",
			readAt, addAt, writeAt, Disassemble(p))
	}
	if !(readAt < addAt && addAt < writeAt) {
		t.Errorf("n += 1 must read, add, then write; the order was read=%d add=%d write=%d:\n%s",
			readAt, addAt, writeAt, Disassemble(p))
	}
}

func TestAFileWithAMistakeStillLowers(t *testing.T) {
	bag := &diag.Bag{}
	f := parser.Parse("t.mhn", []byte("x = \n"), bag)
	p := Lower(f, bag)
	if p == nil || p.Top == nil || len(p.Top.Code) == 0 {
		t.Fatal("a file that did not read must still produce a program with code in it")
	}
	if p.Top.Code[len(p.Top.Code)-1].Op != OpReturnNil {
		t.Error("a program must always end by returning, so a machine never runs off the end")
	}
}

func TestForInIsNotANumberRange(t *testing.T) {
	p := lower(t, "xs = [1, 2, 3]\nhar x dar xs\n    chap_khat x\n")
	if !has(ops(p, 0), "MakeStepRange") {
		t.Errorf("walking a collection is its own operation:\n%s", Disassemble(p))
	}
}
