package ir

import (
	"mihan/internal/ast"
	"mihan/internal/token"
)

func (l *lowerer) expr(e ast.Expr) {
	switch x := e.(type) {
	case *ast.IntLit, *ast.FloatLit, *ast.StrLit, *ast.CharLit, *ast.BoolLit, *ast.HichLit:
		l.emit(Instr{Op: OpConst, A: l.constOf(x), Line: l.line(x)})

	case *ast.Ident:
		l.load(x.Name, l.line(x))

	case *ast.Group:
		l.expr(x.X)

	case *ast.Unary:
		l.expr(x.X)
		op, ok := unaryOp(x.Op)
		if !ok {
			l.emit(Instr{Op: OpNop, Line: l.line(x)})
			return
		}
		l.emit(Instr{Op: op, Line: l.line(x)})

	case *ast.Binary:
		l.binary(x)

	case *ast.Ternary:
		l.expr(x.Cond)
		other := l.emit(Instr{Op: OpJumpIfFalse, A: -1, Line: l.line(x)})
		l.expr(x.Then)
		end := l.emit(Instr{Op: OpJump, A: -1, Line: l.line(x)})
		l.patch(other, l.here())
		l.expr(x.Else)
		l.patch(end, l.here())

	case *ast.ListLit:
		for _, item := range x.Elems {
			l.expr(item)
		}
		l.emit(Instr{Op: OpMakeList, A: int32(len(x.Elems)), Line: l.line(x)})

	case *ast.SetLit:
		for _, item := range x.Elems {
			l.expr(item)
		}
		l.emit(Instr{Op: OpMakeSet, A: int32(len(x.Elems)), Line: l.line(x)})

	case *ast.DictLit:
		for _, e := range x.Entries {
			l.expr(e.Key)
			l.expr(e.Value)
		}
		l.emit(Instr{Op: OpMakeDict, A: int32(len(x.Entries)), Line: l.line(x)})

	case *ast.RangeExpr:
		l.expr(x.Lo)
		l.expr(x.Hi)
		l.emit(Instr{Op: OpMakeRange, B: boolBit(x.Inclusive), Line: l.line(x)})

	case *ast.InterpLit:
		l.interp(x)

	case *ast.Index:
		l.expr(x.X)
		l.expr(x.At)
		l.emit(Instr{Op: OpIndex, Line: l.line(x)})

	case *ast.Slice:
		l.expr(x.X)
		if x.Lo == nil {
			l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstHich}), Line: l.line(x)})
		} else {
			l.expr(x.Lo)
		}
		if x.Hi == nil {
			l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstHich}), Line: l.line(x)})
		} else {
			l.expr(x.Hi)
		}
		l.emit(Instr{Op: OpSlice, B: boolBit(x.Inclusive), Line: l.line(x)})

	case *ast.Call:
		l.call(x)

	case *ast.Member:
		l.member(x)

	case *ast.Lambda:
		fn := l.buildLambda(x)
		l.emit(Instr{Op: OpMakeFunc, A: fn, Line: l.line(x)})

	case *ast.ErrorExpr:
		l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstHich}), Line: l.line(x)})

	default:
		l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstHich}), Line: int32(1)})
	}
}

func (l *lowerer) binary(b *ast.Binary) {

	switch b.Op {
	case "&&", "and", "va":
		l.shortCircuit(b, false)
		return
	case "||", "or", "ya":
		l.shortCircuit(b, true)
		return
	}
	l.expr(b.L)
	l.expr(b.R)
	op, ok := binaryOp(b.Op)
	if !ok {
		l.emit(Instr{Op: OpNop, Line: l.line(b)})
		return
	}
	l.emit(Instr{Op: op, Line: l.line(b)})
}

func (l *lowerer) shortCircuit(b *ast.Binary, or bool) {

	jumpIf := OpJumpIfFalse
	answer := true
	other := false
	if or {
		jumpIf = OpJumpIfTrue
		answer = false
		other = true
	}
	l.expr(b.L)
	first := l.emit(Instr{Op: jumpIf, A: -1, Line: l.line(b)})
	l.expr(b.R)
	second := l.emit(Instr{Op: jumpIf, A: -1, Line: l.line(b)})

	l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstBool, B: answer}), Line: l.line(b)})
	end := l.emit(Instr{Op: OpJump, A: -1, Line: l.line(b)})
	l.patch(first, l.here())
	l.patch(second, l.here())
	l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstBool, B: other}), Line: l.line(b)})
	l.patch(end, l.here())
}

func (l *lowerer) interp(x *ast.InterpLit) {

	if len(x.Segs) == 1 && x.Segs[0].Expr == nil {
		l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstStr, S: x.Segs[0].Text}), Line: l.line(x)})
		return
	}

	n := int32(0)
	for _, seg := range x.Segs {
		if seg.Expr == nil {
			l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstStr, S: seg.Text}), Line: l.line(x)})
		} else {
			l.expr(seg.Expr)
			l.emit(Instr{Op: OpToStr, Line: l.line(x)})
		}
		n++
	}
	l.emit(Instr{Op: OpInterpolate, A: n, Line: l.line(x)})
}

func (l *lowerer) call(c *ast.Call) {
	switch fn := c.Fn.(type) {
	case *ast.Ident:

		op, slot := l.resolve(fn.Name)
		l.emit(Instr{Op: op, A: slot, Line: l.line(c)})
		l.emit(Instr{Op: OpGetMethod, A: l.nameOf(fn.Name), B: -1, Line: l.line(c)})
	default:
		l.expr(c.Fn)
	}
	l.args(c.Args, c.Spread, c)
	l.emit(Instr{Op: OpCall, A: int32(len(c.Args)), B: boolBit(c.Bare), Line: l.line(c)})
}

func (l *lowerer) member(m *ast.Member) {
	l.expr(m.X)
	if m.Call {

		l.emit(Instr{Op: OpGetMethod, A: l.nameOf(m.Name), B: -1, Line: l.line(m)})
		l.args(m.Args, m.Spread, m)
		l.emit(Instr{Op: OpCallMethod, A: int32(len(m.Args)), B: l.nameOf(m.Name), Line: l.line(m)})
		return
	}
	l.emit(Instr{Op: OpGetField, A: l.nameOf(m.Name), Line: l.line(m)})
}

func (l *lowerer) args(args []ast.Arg, spread bool, at ast.Node) {
	for _, a := range args {
		l.expr(a.Expr)
	}
	if spread {
		l.emit(Instr{Op: OpSpread, Line: l.line(at)})
	}
}

func unaryOp(k token.Kind) (Op, bool) {
	switch k {
	case "-":
		return OpNeg, true
	case "+":
		return OpInvalid, false
	case "!", "na":

		return OpNot, true
	case "~":
		return OpBitNot, true
	}
	return OpInvalid, false
}

func binaryOp(k token.Kind) (Op, bool) {
	switch k {
	case "+":
		return OpAdd, true
	case "-":
		return OpSub, true
	case "*":
		return OpMul, true
	case "/":
		return OpDiv, true
	case "%":
		return OpRem, true
	case "**":
		return OpPow, true
	case "==":
		return OpEq, true
	case "!=":
		return OpNe, true
	case "<":
		return OpLt, true
	case "<=":
		return OpLe, true
	case ">":
		return OpGt, true
	case ">=":
		return OpGe, true
	case "&&", "and", "||", "or":

		return OpInvalid, false
	case "&":
		return OpBitAnd, true
	case "|":
		return OpBitOr, true
	case "^":
		return OpBitXor, true
	case "<<", ">>":
		return OpShift, true
	case "in", "dar":

		return OpIn, true
	case "ghesmat":

		return OpIDiv, true
	}
	return OpInvalid, false
}
