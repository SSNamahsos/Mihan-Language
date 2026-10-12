package ir

import "testing"

func TestEveryOperationHasAName(t *testing.T) {

	ops := []Op{
		OpInvalid, OpConst,
		OpLocalGet, OpLocalSet, OpLocalDefine,
		OpUpvalueGet, OpUpvalueSet, OpUpvalueDefine,
		OpGlobalGet, OpGlobalSet, OpGlobalDefine,
		OpName, OpSelf,
		OpNeg, OpNot, OpBitNot, OpTruthy,
		OpAdd, OpSub, OpMul, OpDiv, OpRem, OpPow, OpIDiv,
		OpBitAnd, OpBitOr, OpBitXor, OpShift,
		OpEq, OpNe, OpLt, OpLe, OpGt, OpGe, OpIn,
		OpMakeList, OpMakeDict, OpMakeSet, OpMakeRange, OpMakeStepRange,
		OpIndex, OpSetIndex, OpSlice, OpAppend,
		OpInterpolate, OpToStr, OpLen,
		OpNewClass, OpNewObject, OpGetField, OpSetField, OpGetMethod,
		OpCall, OpCallBuiltin, OpCallMethod, OpSpread,
		OpJump, OpJumpIfFalse, OpJumpIfTrue, OpJumpIfNotNil,
		OpReturn, OpReturnNil, OpPop,
		OpPushHandler, OpPopHandler, OpThrow, OpCatch,
		OpDefineFunc, OpMakeFunc, OpDefineClass, OpDefineEnum, OpDefineStruct,
		OpDefineInterface, OpPushScope, OpPopScope, OpPushLoop, OpPopLoop,
		OpTalignArranged, OpNop,
		OpTypeOf, OpIsType,
	}
	seen := map[Op]bool{}
	for _, o := range ops {
		if seen[o] {
			t.Errorf("%s is listed twice in the test, so one row is being checked twice", o)
		}
		seen[o] = true
		if o != OpInvalid && !o.valid() {
			t.Errorf("%d is outside the valid range", int(o))
		}
		if o.String() == "Op(?)" || o.String() == "Invalid" && o != OpInvalid {
			t.Errorf("operation %d has no name in opNames", int(o))
		}
	}

	for i, n := range opNames {
		if n == "" {
			continue
		}
		if !seen[Op(i)] {
			t.Errorf("opNames has %q at %d, which is not in the const block", n, i)
		}
	}
	if len(opNames) != int(opCount) {
		t.Errorf("opNames has %d entries but there are %d operations", len(opNames), int(opCount))
	}
}

func TestTheInvalidOperationIsTheZeroValue(t *testing.T) {

	var in Instr
	if in.Op != OpInvalid {
		t.Fatalf("an empty instruction is %s, not OpInvalid", in.Op)
	}
	if in.Op.valid() {
		t.Error("OpInvalid must not be a valid operation")
	}
}

func TestPoolsDoNotHoldTheSameThingTwice(t *testing.T) {
	p := NewProgram("t.mhn")
	if p.Top != p.Funcs[0] {
		t.Fatal("the top level must be the first function, so index 0 names it")
	}
	a := p.AddName("naam")
	if b := p.AddName("naam"); a != b {
		t.Errorf("the same name was given two numbers, %d and %d", a, b)
	}
	x := p.AddConst(Const{Kind: ConstInt, I: 7})
	if y := p.AddConst(Const{Kind: ConstInt, I: 7}); x != y {
		t.Errorf("the same number was given two numbers, %d and %d", x, y)
	}

	if z := p.AddConst(Const{Kind: ConstRune, I: 7}); z == x {
		t.Error("a character must not be the same constant as the number of the same value")
	}

	if z := p.AddConst(Const{Kind: ConstFloat, F: 7}); z == x {
		t.Error("a fraction must not be the same constant as the whole number of the same value")
	}
}

func TestDisassembleSaysWhereItIs(t *testing.T) {
	p := NewProgram("t.mhn")
	p.Top.Code = []Instr{
		{Op: OpConst, A: p.AddConst(Const{Kind: ConstInt, I: 1}), Line: 3},
		{Op: OpGlobalDefine, A: p.AddName("n"), Line: 3},
		{Op: OpReturnNil, Line: 4},
	}
	got := Disassemble(p)
	for _, want := range []string{"Const 0 0", "GlobalDefine 0 0", "ReturnNil", "line 3"} {
		if !contains(got, want) {
			t.Errorf("the disassembly does not say %q:\n%s", want, got)
		}
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
