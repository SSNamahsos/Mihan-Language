package vm

import (
	"testing"

	"mihan/internal/diag"
	"mihan/internal/ir"
	"mihan/internal/parser"
)

func TestAWholeNumberCostsNothingToHold(t *testing.T) {

	a, b := Int(3), Int(4)
	if allocs := testing.AllocsPerRun(200, func() {
		_, _ = arith(ir.OpAdd, a, b)
	}); allocs != 0 {
		t.Errorf("adding two whole numbers allocated %.1f times, so a loop of them would too", allocs)
	}
}

func TestAMillionArithmeticOperationsAllocateNothing(t *testing.T) {

	allocs := testing.AllocsPerRun(20, func() {
		n := Int(0)
		for i := int64(0); i < 1_000_000; i++ {
			twice, f := arith(ir.OpMul, Int(i), Int(2))
			if f != nil {
				t.Fatal(f.en)
			}
			sum, f := arith(ir.OpAdd, n, twice)
			if f != nil {
				t.Fatal(f.en)
			}
			n, f = arith(ir.OpSub, sum, Int(1))
			if f != nil {
				t.Fatal(f.en)
			}
		}
		if n.n == 0 {
			t.Fatal("the loop did nothing")
		}
	})
	if allocs > 1 {
		t.Errorf("a million multiplies, adds and subtracts allocated %.1f times, "+
			"and the interpreter needs one for every single one of them", allocs)
	}
}

func TestComparingAndAskingAllocatesNothing(t *testing.T) {
	a, b := Int(1), Int(2)
	if allocs := testing.AllocsPerRun(200, func() {
		_, _ = compare(a, b)
		_ = equal(a, b)
		_ = a.Truthy()
		_ = a.TypeName()
	}); allocs != 0 {
		t.Errorf("comparing two whole numbers allocated %.1f times, and a loop's condition does this every turn", allocs)
	}
}

func TestTheBenchmarkShapeStillLowers(t *testing.T) {
	src := `
n = 0
har i az 0 ta 1_000_000
    n = n + i * 2 - 1
chap_khat n
`
	bag := &diag.Bag{}
	file := parser.Parse("t.mhn", []byte(src), bag)
	if bag.HasErrors() {
		t.Fatalf("the program did not read: %s", bag.Sorted()[0].EN)
	}
	p := ir.Lower(file, bag)
	if p == nil || p.Top == nil || len(p.Top.Code) == 0 {
		t.Fatal("the program lowered to nothing")
	}

	var jumps, adds int
	for _, in := range p.Top.Code {
		switch in.Op {
		case ir.OpJump, ir.OpJumpIfFalse:
			jumps++
		case ir.OpAdd, ir.OpMul, ir.OpSub:
			adds++
		}
	}
	if jumps < 2 {
		t.Errorf("the loop lowered to %d jumps, which is not a loop:\n%s", jumps, ir.Disassemble(p))
	}
	if adds < 3 {
		t.Errorf("the loop body lowered to %d arithmetic operations, which is not the three it wrote", adds)
	}
}
