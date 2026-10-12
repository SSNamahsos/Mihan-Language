package ir

import (
	"mihan/internal/diag"
)

type Op uint8

const (
	OpInvalid Op = iota

	OpConst

	OpLocalGet
	OpLocalSet
	OpLocalDefine

	OpUpvalueGet
	OpUpvalueSet
	OpUpvalueDefine

	OpGlobalGet
	OpGlobalSet
	OpGlobalDefine

	OpName

	OpSelf

	OpNeg
	OpNot
	OpBitNot
	OpTruthy

	OpAdd
	OpSub
	OpMul
	OpDiv
	OpRem
	OpPow

	OpIDiv

	OpBitAnd
	OpBitOr
	OpBitXor
	OpShift

	OpEq
	OpNe

	OpLt
	OpLe
	OpGt
	OpGe

	OpIn

	OpMakeList
	OpMakeDict
	OpMakeSet

	OpMakeRange

	OpMakeStepRange

	OpIndex
	OpSetIndex

	OpSlice

	OpAppend

	OpInterpolate

	OpToStr

	OpLen

	OpNewClass

	OpNewObject

	OpGetField
	OpSetField

	OpGetMethod

	OpCall

	OpCallBuiltin

	OpCallMethod

	OpSpread

	OpJump
	OpJumpIfFalse
	OpJumpIfTrue

	OpJumpIfNotNil

	OpReturn
	OpReturnNil

	OpPop

	OpPushHandler

	OpPopHandler

	OpThrow

	OpCatch

	OpDefineFunc

	OpMakeFunc

	OpDefineClass

	OpDefineEnum

	OpDefineStruct

	OpDefineInterface

	OpPushScope
	OpPopScope

	OpPushLoop
	OpPopLoop

	OpTalignArranged

	OpNop

	OpTypeOf

	OpIsType

	opCount
)

func (o Op) valid() bool { return o > OpInvalid && o < opCount }

func (o Op) String() string {
	if int(o) < len(opNames) && opNames[o] != "" {
		return opNames[o]
	}
	return "Op(?)"
}

var opNames = [...]string{
	OpInvalid:         "Invalid",
	OpConst:           "Const",
	OpLocalGet:        "LocalGet",
	OpLocalSet:        "LocalSet",
	OpLocalDefine:     "LocalDefine",
	OpUpvalueGet:      "UpvalueGet",
	OpUpvalueSet:      "UpvalueSet",
	OpUpvalueDefine:   "UpvalueDefine",
	OpGlobalGet:       "GlobalGet",
	OpGlobalSet:       "GlobalSet",
	OpGlobalDefine:    "GlobalDefine",
	OpName:            "Name",
	OpSelf:            "Self",
	OpNeg:             "Neg",
	OpNot:             "Not",
	OpBitNot:          "BitNot",
	OpTruthy:          "Truthy",
	OpAdd:             "Add",
	OpSub:             "Sub",
	OpMul:             "Mul",
	OpDiv:             "Div",
	OpRem:             "Rem",
	OpPow:             "Pow",
	OpIDiv:            "IDiv",
	OpBitAnd:          "BitAnd",
	OpBitOr:           "BitOr",
	OpBitXor:          "BitXor",
	OpShift:           "Shift",
	OpEq:              "Eq",
	OpNe:              "Ne",
	OpLt:              "Lt",
	OpLe:              "Le",
	OpGt:              "Gt",
	OpGe:              "Ge",
	OpIn:              "In",
	OpMakeList:        "MakeList",
	OpMakeDict:        "MakeDict",
	OpMakeSet:         "MakeSet",
	OpMakeRange:       "MakeRange",
	OpMakeStepRange:   "MakeStepRange",
	OpIndex:           "Index",
	OpSetIndex:        "SetIndex",
	OpSlice:           "Slice",
	OpAppend:          "Append",
	OpInterpolate:     "Interpolate",
	OpToStr:           "ToStr",
	OpLen:             "Len",
	OpNewClass:        "NewClass",
	OpNewObject:       "NewObject",
	OpGetField:        "GetField",
	OpSetField:        "SetField",
	OpGetMethod:       "GetMethod",
	OpCall:            "Call",
	OpCallBuiltin:     "CallBuiltin",
	OpCallMethod:      "CallMethod",
	OpSpread:          "Spread",
	OpJump:            "Jump",
	OpJumpIfFalse:     "JumpIfFalse",
	OpJumpIfTrue:      "JumpIfTrue",
	OpJumpIfNotNil:    "JumpIfNotNil",
	OpReturn:          "Return",
	OpReturnNil:       "ReturnNil",
	OpPop:             "Pop",
	OpPushHandler:     "PushHandler",
	OpPopHandler:      "PopHandler",
	OpThrow:           "Throw",
	OpCatch:           "Catch",
	OpDefineFunc:      "DefineFunc",
	OpMakeFunc:        "MakeFunc",
	OpDefineClass:     "DefineClass",
	OpDefineEnum:      "DefineEnum",
	OpDefineStruct:    "DefineStruct",
	OpDefineInterface: "DefineInterface",
	OpPushScope:       "PushScope",
	OpPopScope:        "PopScope",
	OpPushLoop:        "PushLoop",
	OpPopLoop:         "PopLoop",
	OpTalignArranged:  "TalignArranged",
	OpNop:             "Nop",
	OpTypeOf:          "TypeOf",
	OpIsType:          "IsType",
}

type Instr struct {
	Op   Op
	A    int32
	B    int32
	Line int32
}

type FuncKind uint8

const (
	FuncTop FuncKind = iota

	FuncPlain

	FuncMethod

	FuncCtor

	FuncLambda
)

func (k FuncKind) String() string {
	switch k {
	case FuncTop:
		return "top"
	case FuncPlain:
		return "funct"
	case FuncMethod:
		return "method"
	case FuncCtor:
		return "__sakht__"
	case FuncLambda:
		return "lambda"
	}
	return "?"
}

type Func struct {
	Name string

	Kind FuncKind

	Params int

	Locals int

	Varargs bool

	Upvals int

	UpvalNames []int32

	LocalNames []int32

	Class *Class

	Code []Instr

	Span diag.Span

	Doc string
}

type Field struct {
	Name string

	TypeName string

	Default int32

	Sabit bool

	Baza, Mahram bool
}

type Method struct {
	Name string

	Fn int32

	Span diag.Span
}

type Class struct {
	Name   string
	Fields []Field

	FieldStart int32
	FieldCount int32

	Methods []Method

	Bases []string

	CtorFn int32

	IsStruct bool

	Span diag.Span
	Doc  string

	Baza, Mahram bool
}

type ConstKind uint8

const (
	ConstInt ConstKind = iota
	ConstFloat
	ConstStr
	ConstBool

	ConstHich

	ConstRune
)

type Const struct {
	Kind ConstKind
	I    int64
	F    float64
	S    string
	B    bool
}

type Program struct {
	Top *Func

	Funcs []*Func

	Classes []*Class

	Consts []Const
	Names  []string

	Nums []int32

	Source string

	ImportPaths []string
	ModulePath  string
	ProgramName string
}

func NewProgram(source string) *Program {
	p := &Program{Source: source}
	p.Funcs = append(p.Funcs, &Func{Kind: FuncTop, Name: "<top>"})
	p.Top = p.Funcs[0]
	return p
}

func (p *Program) AddFunc(f *Func) int32 {
	p.Funcs = append(p.Funcs, f)
	return int32(len(p.Funcs) - 1)
}

func (p *Program) AddClass(c *Class) int32 {
	p.Classes = append(p.Classes, c)
	return int32(len(p.Classes) - 1)
}

func (p *Program) AddConst(c Const) int32 {
	for i, have := range p.Consts {
		if sameConst(have, c) {
			return int32(i)
		}
	}
	p.Consts = append(p.Consts, c)
	return int32(len(p.Consts) - 1)
}

func sameConst(a, b Const) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case ConstInt, ConstRune:
		return a.I == b.I
	case ConstFloat:
		return a.F == b.F
	case ConstStr:
		return a.S == b.S
	case ConstBool:
		return a.B == b.B
	}
	return true
}

func (p *Program) AddName(s string) int32 {
	for i, have := range p.Names {
		if have == s {
			return int32(i)
		}
	}
	p.Names = append(p.Names, s)
	return int32(len(p.Names) - 1)
}

func (p *Program) NumRun(nums ...int32) int32 {
	start := len(p.Nums)
	p.Nums = append(p.Nums, nums...)
	return int32(start)
}

func Disassemble(p *Program) string {
	out := make([]byte, 0, 4096)
	for i, f := range p.Funcs {
		out = append(out, "funct "...)
		out = append(out, f.Name...)
		out = append(out, " ("...)
		out = appendInt(out, int64(f.Kind))
		out = append(out, ") params="...)
		out = appendInt(out, int64(f.Params))
		out = append(out, " locals="...)
		out = appendInt(out, int64(f.Locals))
		out = append(out, '\n')
		for pc, in := range f.Code {
			out = appendInt(out, int64(pc))
			out = append(out, ":  "...)
			out = append(out, in.Op.String()...)
			out = append(out, ' ')
			out = appendInt(out, int64(in.A))
			out = append(out, ' ')
			out = appendInt(out, int64(in.B))
			out = append(out, "   ; line "...)
			out = appendInt(out, int64(in.Line))
			out = append(out, '\n')
		}
		_ = i
	}
	return string(out)
}

func appendInt(b []byte, v int64) []byte {
	if v == 0 {
		return append(b, '0')
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var tmp [20]byte
	i := len(tmp)
	for v > 0 {
		i--
		tmp[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		return append(append(b, '-'), tmp[i:]...)
	}
	return append(b, tmp[i:]...)
}
