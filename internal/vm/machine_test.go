package vm

import (
	"bytes"
	"strings"
	"testing"

	"mihan/internal/diag"
	"mihan/internal/eval"
	"mihan/internal/ir"
	"mihan/internal/parser"
)

func runOnBoth(t *testing.T, src string) {
	t.Helper()

	bag := &diag.Bag{}
	file := parser.Parse("t.mhn", []byte(src), bag)
	if bag.HasErrors() {
		t.Fatalf("the program did not read: %s", bag.Sorted()[0].EN)
	}

	var want bytes.Buffer
	in := eval.New(&want, nil)
	if err := in.RunFile(file); err != nil {
		t.Fatalf("the interpreter stopped: %v", err)
	}

	lbag := &diag.Bag{}
	p := ir.Lower(parser.Parse("t.mhn", []byte(src), lbag), lbag)
	if p == nil {
		t.Fatalf("the program did not lower: %s", lbag.Sorted()[0].EN)
	}

	var got bytes.Buffer
	m := New(p, &got)
	if err := m.Run(); err != nil {
		t.Fatalf("the machine stopped: %v\nit had printed: %q", err, got.String())
	}
	if got.String() != want.String() {
		t.Errorf("the two disagree.\n  program:     %s\n  interpreter: %q\n  machine:     %q",
			strings.TrimSpace(src), want.String(), got.String())
	}
}

func TestCountingIsTheSameLanguage(t *testing.T) {
	runOnBoth(t, `
n = 0
har i az 0 ta 10
    n = n + i
chap_khat n
`)
}

func TestAMillionCountsAgree(t *testing.T) {
	runOnBoth(t, `
n = 0
har i az 0 ta 1_000_000
    n = n + i * 2 - 1
chap_khat n
`)
}

func TestTheFourKindsOfNumberAgree(t *testing.T) {
	runOnBoth(t, `
chap_khat 7 / 2
chap_khat 6 / 3
chap_khat ghesmat(7, 2)
chap_khat 7 ghesmat 2
chap_khat 2 ** 10
chap_khat 7 % 3
chap_khat -3 % 7
chap_khat 1.5 + 2.5
`)
}

func TestWordOperatorsAgree(t *testing.T) {
	runOnBoth(t, `
chap_khat 7 ghesmat 2
chap_khat -7 ghesmat 2
chap_khat 2 dar [1, 2, 3]
chap_khat 9 dar [1, 2, 3]
chap_khat "sa" dar "salam"
chap_khat na dorost
chap_khat na 0
chap_khat dorost va ghalat
chap_khat dorost ya ghalat
`)
}

func TestConditionsAndJumpingAgree(t *testing.T) {
	runOnBoth(t, `
har x az 0 ta 5
    agar x % 2 == 0
        chap_khat "even " + x
    warna
        chap_khat "odd " + x
`)
}

func TestAWhileLoopAgrees(t *testing.T) {
	runOnBoth(t, `
i = 0
ta i < 5
    chap_khat i
    i += 1
`)
}

func TestTextAgrees(t *testing.T) {
	runOnBoth(t, `
s = "salam"
chap_khat s
chap_khat tule s
chap_khat s + " " + "dunya"
`)
}

func TestInterpolatedTextAgrees(t *testing.T) {
	runOnBoth(t, `
naam = "donya"
chap_khat $"salam {naam}!"
chap_khat $"2 + 2 = {2 + 2} ast"
chap_khat $"6 / 3 = {6 / 3} ast"
chap_khat $"dorost ast? {1 < 2}"
chap_khat $"hich ast? {hich}"
chap_khat $"{1}{2}{3} = {1} + {2} + {3}"
chap_khat $"{{{naam}}}"
`)
}

func TestInterpolationOfInterpolationAgrees(t *testing.T) {
	runOnBoth(t, `
a = $"x = {1 + 1}"
b = $"and {a}, truly"
chap_khat b
chap_khat $"{a}"
chap_khat $"{{{a}}}"
`)
}

func TestSlicingAgrees(t *testing.T) {
	runOnBoth(t, `
xs = [1, 2, 3, 4, 5]
chap_khat xs[1:3]
chap_khat xs[1..3]
chap_khat xs[:2]
chap_khat xs[2:]
chap_khat xs[-2:]
chap_khat xs[:-1]
chap_khat xs[5:]
chap_khat xs[:99]
`)
}

func TestSlicingTextAgrees(t *testing.T) {
	runOnBoth(t, `
s = "salam donya"
chap_khat s[1:3]
chap_khat s[1..3]
chap_khat s[:5]
chap_khat s[6:]
chap_khat s[-5:]
chap_khat s[99:]
chap_khat s[..4]
`)
}

func TestSlicingSetsAndRangesAgrees(t *testing.T) {
	runOnBoth(t, `
st = {1, 2, 3, 4}
chap_khat st[1:3]
r = 1..6
chap_khat r[1:3]
chap_khat r[2..4]
`)
}

func TestSlicingTheWrongThingIsRefused(t *testing.T) {
	check := func(src string) {
		t.Helper()
		p := lowerFor(t, src)
		err := New(p, &devNull{}).Run()
		if err == nil {
			t.Fatalf("the machine sliced something unsliceable and said nothing:\n%s", src)
		}
		if !strings.Contains(err.Error(), "bresh") {
			t.Errorf("the mistake should say a slice was refused, and it said: %v", err)
		}
	}
	check("chap_khat 5[1:2]")
	check("chap_khat hich[1:2]")
}

func TestAListAgrees(t *testing.T) {
	runOnBoth(t, `
xs = [1, 2, 3]
chap_khat xs
chap_khat tule xs
chap_khat xs[1]
xs[0] = 10
chap_khat xs
`)
}

func TestADictAgrees(t *testing.T) {
	runOnBoth(t, `
d = {"a": 1, "b": 2}
chap_khat d
chap_khat d["a"]
chap_khat tule d
chap_khat "a" dar d
d["c"] = 3
chap_khat d
har k dar d
    chap_khat k
`)
}

func TestASetAgrees(t *testing.T) {
	runOnBoth(t, `
s = {1, 2, 3, 2, 1}
chap_khat s
chap_khat tule s
chap_khat 2 dar s
`)
}

func TestADictWithTextKeysAgrees(t *testing.T) {
	runOnBoth(t, `
d = {}
d["salam"] = "donya"
d["khodahafez"] = "hamishe"
chap_khat d
chap_khat d["salam"]
chap_khat tule d
`)
}

func TestWhatTheMachineCannotDoIsNamedNotGuessed(t *testing.T) {
	runOnBoth(t, `
x = 1
x += 1
chap_khat x
`)
}

func TestWhatTheMachineCannotDoIsNamedRatherThanNumbered(t *testing.T) {
	p := ir.Lower(parser.Parse("t.mhn", []byte("x = 1"), &diag.Bag{}), &diag.Bag{})
	m := New(p, &bytes.Buffer{})
	for _, op := range []ir.Op{ir.OpNewClass, ir.OpThrow, ir.OpCatch} {
		err := m.missing(op, 0)
		if err == nil {
			t.Fatalf("%s: the machine said it could do something it cannot", op)
		}
		msg := err.Error()
		if strings.Contains(msg, "operation ") {
			t.Errorf("%s was reported as a number rather than a name: %s", op, msg)
		}
		if !strings.Contains(msg, op.String()) {
			t.Errorf("%s was not reported by its own name: %s", op, msg)
		}
	}
}

func TestAClassNowRunsRatherThanBeingRefused(t *testing.T) {
	runOnBoth(t, `
darghan Box
    tula = 0
b = Box()
chap_khat b.tula
`)
}

func TestFailuresNowRunRatherThanBeingRefused(t *testing.T) {
	runOnBoth(t, `
funct f(n)
    agar n == 1
        portab khata("one")
    agar n == 2
        portab khata("two", n)
    chap_khat "ok"

koshish
    f(1)
giroft khata e
    chap_khat $"got {e.message}"

koshish
    f(2)
giroft khata e
    chap_khat $"got {e.message} data {e.data}"

koshish
    f(3)
giroft khata e
    chap_khat "not reached"

koshish
    chap_khat "clean"
nemigir
    chap_khat "after"
`)
}

func TestAFailureNobodyCatchesStopsTheMachine(t *testing.T) {
	bag := &diag.Bag{}
	src := `
funct bad()
    portab khata("worse than that")
bad()
`
	p := ir.Lower(parser.Parse("t.mhn", []byte(src), bag), bag)
	if p == nil {
		t.Fatalf("the program did not lower: %s", bag.Sorted()[0].EN)
	}
	err := New(p, &devNull{}).Run()
	if err == nil {
		t.Fatal("a failure nobody caught did not stop the machine")
	}
	if !strings.Contains(err.Error(), "worse than that") {
		t.Errorf("the failure did not carry its message out: %v", err)
	}
}

func TestAnOperationStillMissingIsRefusedNotGuessed(t *testing.T) {
	p := ir.Lower(parser.Parse("t.mhn", []byte("x = 1"), &diag.Bag{}), &diag.Bag{})
	m := New(p, &bytes.Buffer{})
	for _, op := range []ir.Op{ir.OpSelf, ir.OpIsType, ir.OpNewObject, ir.OpTypeOf} {
		err := m.missing(op, 0)
		if err == nil {
			t.Fatalf("%s: the machine said it could do something it cannot", op)
		}
		msg := err.Error()
		if strings.Contains(msg, "operation ") {
			t.Errorf("%s was reported as a number rather than a name: %s", op, msg)
		}
		if !strings.Contains(msg, op.String()) {
			t.Errorf("%s was not reported by its own name: %s", op, msg)
		}
	}
}

func TestUnfinishedWorkIsNamed(t *testing.T) {

	p := ir.Lower(parser.Parse("t.mhn", []byte("x = 1"), &diag.Bag{}), &diag.Bag{})
	err := New(p, &devNull{}).missing(ir.OpSelf, 3)
	if err == nil {
		t.Fatal("an operation the machine cannot do was not refused")
	}
	if !strings.Contains(err.Error(), "in mashin hanooz") {
		t.Errorf("the refusal did not name what it cannot do: %v", err)
	}
	if !strings.Contains(err.Error(), "Self") {
		t.Errorf("the refusal did not name the operation itself: %v", err)
	}
}

func lowerFor(t *testing.T, src string) *ir.Program {
	t.Helper()
	bag := &diag.Bag{}
	p := ir.Lower(parser.Parse("t.mhn", []byte(src), bag), bag)
	if p == nil {
		t.Fatalf("the program did not lower: %s", bag.Sorted()[0].EN)
	}
	return p
}

type devNull struct{}

func (devNull) Write(p []byte) (int, error) { return len(p), nil }

func (m *Machine) usedSlots() int { return len(m.slots) }
