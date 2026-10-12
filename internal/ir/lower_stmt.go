package ir

import (
	"mihan/internal/ast"
	"mihan/internal/diag"
	"mihan/internal/token"
)

func (l *lowerer) decls(stmts []ast.Stmt) {
	for _, s := range stmts {
		l.stmt(s)
	}
}

func (l *lowerer) build(d ast.Stmt) {
	switch x := d.(type) {
	case *ast.FuncDecl:
		l.funcIdx[x] = l.buildFunc(x.Name, FuncPlain, x.Params, x.Body, x.Span(), x.Doc, false)
	case *ast.ClassDecl:
		l.clsIdx[x] = l.buildClass(x.Name, x.Fields, x.Methods, x.Bases, x.Span(), x.Doc, false, x.CtorName)
	case *ast.StructDecl:
		l.clsIdx[x] = l.buildClass(x.Name, x.Fields, nil, x.Bases, x.Span(), x.Doc, true, "")
	case *ast.EnumDecl:
		l.enumIdx[x] = l.enumRun(x)
	case *ast.InterfaceDecl:
		l.iflIdx[x] = l.nameRun(x.Methods)
	}
}

func (l *lowerer) nameRun(names []string) int32 {
	start := len(l.p.Names)
	l.p.Names = append(l.p.Names, names...)
	return int32(start)
}

func (l *lowerer) enumRun(d *ast.EnumDecl) int32 {
	run := make([]int32, 0, 1+2*len(d.Values))
	run = append(run, int32(len(d.Values)))
	for i, v := range d.Values {
		run = append(run, l.nameOf(v.Name))
		if v.Value != nil {

			run = append(run, l.constOf(v.Value))
			continue
		}

		run = append(run, int32(i))
	}
	return l.p.NumRun(run...)
}

func (l *lowerer) define(d ast.Stmt, line int32) bool {
	switch x := d.(type) {
	case *ast.FuncDecl:
		l.emit(Instr{Op: OpDefineFunc, A: l.nameOf(x.Name), B: l.funcIdx[x], Line: line})
		return true
	case *ast.ClassDecl:
		l.emit(Instr{Op: OpDefineClass, A: l.nameOf(x.Name), B: l.clsIdx[x], Line: line})
		return true
	case *ast.StructDecl:
		l.emit(Instr{Op: OpDefineStruct, A: l.nameOf(x.Name), B: l.clsIdx[x], Line: line})
		return true
	case *ast.EnumDecl:
		l.emit(Instr{Op: OpDefineEnum, A: l.nameOf(x.Name), B: l.enumIdx[x], Line: line})
		return true
	case *ast.InterfaceDecl:
		l.emit(Instr{Op: OpDefineInterface, A: l.nameOf(x.Name), B: l.iflIdx[x], Line: line})
		return true
	}
	return false
}

func (l *lowerer) buildIn(b *ast.Block) {
	if b == nil {
		return
	}
	for _, s := range b.Stmts {
		l.build(s)
		switch x := s.(type) {
		case *ast.Block:
			l.buildIn(x)
		case *ast.ExprStmt:
			l.buildInExpr(x.X)
		case *ast.IfStmt:
			l.buildIn(x.Then)
			if inner, ok := x.Else.(*ast.Block); ok {
				l.buildIn(inner)
			}
			if inner, ok := x.Else.(*ast.IfStmt); ok {
				l.buildIn(inner.Then)
				if nxt, ok := inner.Else.(*ast.IfStmt); ok {
					l.buildIn(nxt.Then)
				}
			}
		case *ast.WhileStmt:
			l.buildIn(x.Body)
		case *ast.ForStmt:
			l.buildIn(x.Body)
		case *ast.ForInStmt:
			l.buildIn(x.Body)
		case *ast.TryStmt:
			l.buildIn(x.Body)
			if x.Catch != nil {
				l.buildIn(x.Catch.Body)
			}
			l.buildIn(x.Finally)
		}
	}
}

func (l *lowerer) buildLambdasIn(b *ast.Block) {
	if b == nil {
		return
	}
	for _, s := range b.Stmts {
		var exprs []ast.Expr
		switch x := s.(type) {
		case *ast.ExprStmt:
			exprs = append(exprs, x.X)
			if a, ok := x.X.(*ast.Assign); ok {
				exprs = append(exprs, a.Value)
			}
		case *ast.VarDecl:
			exprs = append(exprs, x.Value)
		case *ast.ReturnStmt:
			exprs = append(exprs, x.Value)
		case *ast.ThrowStmt:
			exprs = append(exprs, x.Value)
		case *ast.TalignStmt:
			exprs = append(exprs, x.X)
		default:
			continue
		}
		for _, e := range exprs {
			l.buildInExpr(e)
		}
	}
}

func (l *lowerer) buildInExpr(e ast.Expr) {
	switch x := e.(type) {
	case *ast.Lambda:
		l.lambdaIdx[x] = l.buildLambda(x)
	case *ast.Call:
		for _, a := range x.Args {
			l.buildInExpr(a.Expr)
		}
		l.buildInExpr(x.Fn)
	case *ast.Binary:
		l.buildInExpr(x.L)
		l.buildInExpr(x.R)
	case *ast.Group:
		l.buildInExpr(x.X)
	case *ast.Member:
		l.buildInExpr(x.X)
		for _, a := range x.Args {
			l.buildInExpr(a.Expr)
		}
	case *ast.Index:
		l.buildInExpr(x.X)
		l.buildInExpr(x.At)
	}
}

func (l *lowerer) buildClass(name string, fields []*ast.FieldDecl, methods []*ast.FuncDecl,
	bases []string, span diag.Span, doc string, isStruct bool, ctorName string) int32 {

	c := &Class{Name: name, Bases: bases, CtorFn: -1, IsStruct: isStruct, Span: span, Doc: doc}

	start := len(l.p.Names)
	for _, f := range fields {
		def := int32(-1)
		if f.Default != nil {
			def = l.constOf(f.Default)
		}
		l.p.Names = append(l.p.Names, f.Name)
		c.Fields = append(c.Fields, Field{
			Name:     f.Name,
			TypeName: f.TypeName,
			Default:  def,
			Sabit:    f.Sabit,
			Baza:     f.Baza,
			Mahram:   f.Mahram,
		})
	}
	c.FieldStart = int32(start)

	c.FieldCount = int32(len(fields))

	if ctorName == "" {
		for _, m := range methods {
			if m.Name == "__sakht__" {
				ctorName = m.Name
			}
		}
	}
	for _, m := range methods {
		kind := FuncMethod
		if m.Name == ctorName {
			kind = FuncCtor
		}

		fr := &frameInfo{
			locals: map[string]int32{},
			upvals: map[string]int32{},
			parent: l.frame,
			class:  c,
			kind:   kind,
		}
		fr.declareLocal("khod")
		params := l.buildParams(m.Params, fr)
		fn := &Func{Name: name + "." + m.Name, Kind: kind, Params: int(params) + 1, Doc: m.Doc, Span: m.Span()}
		done := l.begin(fn, fr)
		l.blockBody(m.Body)
		l.emit(Instr{Op: OpReturnNil, Line: endOf(m.Body, m.Span())})
		done()
		idx := l.p.AddFunc(fn)
		if kind == FuncCtor {
			c.CtorFn = idx
		}
		c.Methods = append(c.Methods, Method{Name: m.Name, Fn: idx, Span: m.Span()})
	}
	return l.p.AddClass(c)
}

func endOf(b *ast.Block, span diag.Span) int32 {
	if b != nil && len(b.Stmts) > 0 {
		return int32(b.Stmts[len(b.Stmts)-1].Span().End.Line)
	}
	return int32(span.End.Line)
}

func (l *lowerer) buildFunc(name string, kind FuncKind, params []*ast.Param, body *ast.Block,
	span diag.Span, doc string, varargs bool) int32 {

	if l.depth > 64 {
		l.bag.Add(diag.Errorf("E6001", span,
			"تابع در خودش تعریف شده و هیچ‌وقت تمام نمی‌شود",
			"this funct is written inside itself without end"))
		return l.p.AddFunc(&Func{Name: name, Kind: kind, Code: []Instr{{Op: OpReturnNil, Line: 1}}})
	}
	fr := &frameInfo{
		locals: map[string]int32{},
		upvals: map[string]int32{},
		parent: l.frame,
		kind:   kind,
	}
	n := l.buildParams(params, fr)
	fn := &Func{Name: name, Kind: kind, Params: int(n), Varargs: varargs, Doc: doc, Span: span}
	l.depth++
	done := l.begin(fn, fr)
	l.blockBody(body)
	l.buildLambdasIn(body)
	l.flushPending()
	l.emit(Instr{Op: OpReturnNil, Line: endOf(body, span)})
	done()
	l.depth--
	return l.p.AddFunc(fn)
}

func (l *lowerer) buildLambda(x *ast.Lambda) int32 {
	if x.Expr == nil {
		return l.buildFunc("lambda", FuncLambda, x.Params, x.Body, x.Span(), "", false)
	}
	fr := &frameInfo{
		locals: map[string]int32{},
		upvals: map[string]int32{},
		parent: l.frame,
		kind:   FuncLambda,
	}
	n := l.buildParams(x.Params, fr)
	fn := &Func{Name: "lambda", Kind: FuncLambda, Params: int(n), Span: x.Span()}
	l.depth++
	done := l.begin(fn, fr)

	l.expr(x.Expr)
	l.emit(Instr{Op: OpReturn, Line: l.line(x)})
	l.emit(Instr{Op: OpReturnNil, Line: int32(x.Span().End.Line)})
	done()
	l.depth--
	return l.p.AddFunc(fn)
}

func (l *lowerer) buildParams(params []*ast.Param, fr *frameInfo) int32 {
	var n int32
	for _, p := range params {
		fr.declareLocal(p.Name)
		n++
	}
	return n
}

func (l *lowerer) block(b *ast.Block) {
	if b == nil {
		return
	}
	l.emit(Instr{Op: OpPushScope, Line: l.line(b)})
	l.blockBody(b)
	l.emit(Instr{Op: OpPopScope, Line: int32(b.Span().End.Line)})
}

func (l *lowerer) blockBody(b *ast.Block) {
	if b == nil {
		return
	}
	for _, s := range b.Stmts {
		l.stmt(s)
	}
}

func (l *lowerer) stmt(s ast.Stmt) {
	switch x := s.(type) {
	case *ast.ExprStmt:

		if a, ok := x.X.(*ast.Assign); ok {
			l.assign(a)
			return
		}
		l.expr(x.X)
		l.emit(Instr{Op: OpPop, Line: l.line(s)})

	case *ast.VarDecl:
		l.varDecl(x)

	case *ast.Block:
		l.block(x)

	case *ast.IfStmt:
		l.ifStmt(x)

	case *ast.WhileStmt:
		l.whileStmt(x)

	case *ast.ForStmt:
		l.forStmt(x)

	case *ast.ForInStmt:
		l.forInStmt(x)

	case *ast.ReturnStmt:
		if x.Value == nil {
			l.emit(Instr{Op: OpReturnNil, Line: l.line(s)})
			return
		}
		l.expr(x.Value)
		l.emit(Instr{Op: OpReturn, Line: l.line(s)})

	case *ast.BreakStmt:

		if len(l.loops) == 0 {
			l.emit(Instr{Op: OpNop, Line: l.line(s)})
			return
		}
		pc := l.emit(Instr{Op: OpJump, A: -1, Line: l.line(s)})
		l.loops[len(l.loops)-1].noteBreak(pc)

	case *ast.ContinueStmt:
		if len(l.loops) == 0 {
			l.emit(Instr{Op: OpNop, Line: l.line(s)})
			return
		}
		pc := l.emit(Instr{Op: OpJump, A: -1, Line: l.line(s)})
		l.loops[len(l.loops)-1].noteContinue(pc)

	case *ast.TalignStmt:
		l.talignStmt(x)

	case *ast.TryStmt:
		l.tryStmt(x)

	case *ast.ThrowStmt:
		l.expr(x.Value)
		l.emit(Instr{Op: OpThrow, Line: l.line(s)})

	case *ast.FuncDecl, *ast.ClassDecl, *ast.StructDecl, *ast.EnumDecl, *ast.InterfaceDecl:

		if l.frame.kind == FuncTop {
			l.define(s, l.line(s))
			return
		}
		l.nestedDeclare(s)

	case *ast.ImportDecl, *ast.ModuleDecl, *ast.ProgramDecl:

	default:
		l.emit(Instr{Op: OpNop, Line: l.line(s)})
	}
}

func (l *lowerer) nestedDeclare(s ast.Stmt) {
	var name string
	switch x := s.(type) {
	case *ast.FuncDecl:
		name = x.Name
	case *ast.ClassDecl:
		name = x.Name
	case *ast.StructDecl:
		name = x.Name
	case *ast.EnumDecl:
		name = x.Name
	case *ast.InterfaceDecl:
		name = x.Name
	default:
		return
	}
	pc := l.emit(Instr{Op: OpNop, A: l.nameOf(name), B: -1, Line: l.line(s)})
	switch s.(type) {
	case *ast.FuncDecl:
		l.code[pc].Op = OpDefineFunc
	case *ast.ClassDecl:
		l.code[pc].Op = OpDefineClass
	case *ast.StructDecl:
		l.code[pc].Op = OpDefineStruct
	case *ast.EnumDecl:
		l.code[pc].Op = OpDefineEnum
	case *ast.InterfaceDecl:
		l.code[pc].Op = OpDefineInterface
	}
	l.pending = append(l.pending, pending{pc: pc, decl: s, frame: l.frame})
	slot := l.frame.declareLocal(name)
	l.emit(Instr{Op: OpLocalSet, A: slot, Line: l.line(s)})
}

func (l *lowerer) varDecl(d *ast.VarDecl) {
	switch {
	case d.Value != nil:
		l.expr(d.Value)
	case d.TypeName != "":
		l.zeroOfType(d, d.TypeName)
	default:
		l.emit(Instr{Op: OpConst, A: l.constOf(&ast.HichLit{}), Line: l.line(d)})
	}
	l.store(d.Name, l.line(d))
}

func (l *lowerer) zeroOfType(at ast.Node, typeName string) {
	base := typeName
	for i := 0; i < len(base); i++ {
		if base[i] == '<' {
			base = base[:i]
			break
		}
	}
	line := l.line(at)
	switch base {
	case "arz":
		l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstInt, I: 0}), Line: line})
	case "kasr":
		l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstFloat, F: 0}), Line: line})
	case "matn", "harf":
		l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstStr, S: ""}), Line: line})
	case "bool":
		l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstBool, B: false}), Line: line})
	case "araye", "list":
		l.emit(Instr{Op: OpMakeList, A: 0, Line: line})
	case "dict":
		l.emit(Instr{Op: OpMakeDict, A: 0, Line: line})
	case "set":
		l.emit(Instr{Op: OpMakeSet, A: 0, Line: line})
	default:
		l.emit(Instr{Op: OpConst, A: l.constOf(&ast.HichLit{}), Line: line})
	}
}

func (l *lowerer) assign(a *ast.Assign) {

	if a.Op == "=" {
		l.assignTo(a.Target, a.Value, a)
		return
	}
	op, ok := arithOp(a.Op)
	if !ok {
		l.assignTo(a.Target, a.Value, a)
		return
	}
	switch t := a.Target.(type) {
	case *ast.Ident:
		l.load(t.Name, l.line(a))
		l.expr(a.Value)
		l.emit(Instr{Op: op, Line: l.line(a)})
		l.store(t.Name, l.line(a))
	case *ast.Index:
		l.expr(t.X)
		l.expr(t.At)
		l.loadPlace(a.Target, l.line(a))
		l.expr(a.Value)
		l.emit(Instr{Op: op, Line: l.line(a)})
		l.emit(Instr{Op: OpSetIndex, Line: l.line(a)})
	case *ast.Member:
		l.expr(t.X)
		l.loadPlace(a.Target, l.line(a))
		l.expr(a.Value)
		l.emit(Instr{Op: op, Line: l.line(a)})
		l.emit(Instr{Op: OpSetField, A: l.nameOf(t.Name), Line: l.line(a)})
	default:
		l.assignTo(a.Target, a.Value, a)
	}
}

func (l *lowerer) loadPlace(place ast.Expr, line int32) {
	switch t := place.(type) {
	case *ast.Ident:
		l.load(t.Name, line)
	case *ast.Index:
		l.expr(t.X)
		l.expr(t.At)
		l.emit(Instr{Op: OpIndex, Line: line})
	case *ast.Member:
		l.expr(t.X)
		l.emit(Instr{Op: OpGetField, A: l.nameOf(t.Name), Line: line})
	default:
		l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstHich}), Line: line})
	}
}

func (l *lowerer) assignTo(target ast.Expr, value ast.Expr, at ast.Node) {
	switch t := target.(type) {
	case *ast.Ident:
		l.expr(value)
		l.store(t.Name, l.line(at))
	case *ast.Index:
		l.expr(t.X)
		l.expr(t.At)
		l.expr(value)
		l.emit(Instr{Op: OpSetIndex, Line: l.line(at)})
	case *ast.Member:
		l.expr(t.X)
		l.expr(value)
		l.emit(Instr{Op: OpSetField, A: l.nameOf(t.Name), Line: l.line(at)})
	default:

		l.expr(value)
		l.emit(Instr{Op: OpPop, Line: l.line(at)})
	}
}

func (l *lowerer) ifStmt(d *ast.IfStmt) {
	l.expr(d.Cond)
	elseJump := l.emit(Instr{Op: OpJumpIfFalse, A: -1, Line: l.line(d)})
	l.block(d.Then)
	if d.Else == nil {
		l.patch(elseJump, l.here())
		return
	}
	endJump := l.emit(Instr{Op: OpJump, A: -1, Line: l.line(d)})
	l.patch(elseJump, l.here())
	l.stmt(d.Else)
	l.patch(endJump, l.here())
}

type loopJumps struct {
	breaks    []int32
	continues []int32
}

func (lj *loopJumps) noteBreak(pc int32)    { lj.breaks = append(lj.breaks, pc) }
func (lj *loopJumps) noteContinue(pc int32) { lj.continues = append(lj.continues, pc) }

func (l *lowerer) whileStmt(d *ast.WhileStmt) {
	start := l.here()
	l.emit(Instr{Op: OpPushLoop, Line: l.line(d)})
	l.expr(d.Cond)
	out := l.emit(Instr{Op: OpJumpIfFalse, A: -1, Line: l.line(d)})
	lj := l.pushLoop()
	l.block(d.Body)
	l.popLoop()
	l.emit(Instr{Op: OpJump, A: start, Line: l.line(d)})
	l.patch(out, l.here())
	l.emit(Instr{Op: OpPopLoop, Line: l.line(d)})

	for _, pc := range lj.continues {
		l.patch(pc, start)
	}
	for _, pc := range lj.breaks {
		l.patch(pc, l.here())
	}
}

func (l *lowerer) forStmt(d *ast.ForStmt) {
	l.expr(d.Lo)
	l.expr(d.Hi)
	if d.Inclusive {
		l.emit(Instr{Op: OpToStr, Line: l.line(d)})
		l.emit(Instr{Op: OpPop, Line: l.line(d)})
	}
	slot := l.frame.declareLocal(d.VarName)
	l.emit(Instr{Op: OpPushLoop, Line: l.line(d)})
	l.emit(Instr{Op: OpMakeRange, A: slot, B: boolBit(d.Inclusive), Line: l.line(d)})

	start := l.here()

	skip := l.emit(Instr{Op: OpJumpIfFalse, A: -1, Line: l.line(d)})
	lj := l.pushLoop()
	l.emit(Instr{Op: OpLocalSet, A: slot, Line: l.line(d)})
	if d.Filter != nil {

		l.expr(d.Filter)
		drop := l.emit(Instr{Op: OpJumpIfFalse, A: -1, Line: l.line(d)})
		l.block(d.Body)

		l.emit(Instr{Op: OpJump, A: start, Line: l.line(d)})
		l.patch(drop, l.here())

		l.emit(Instr{Op: OpJump, A: start, Line: l.line(d)})
	} else {
		l.block(d.Body)
		l.emit(Instr{Op: OpJump, A: start, Line: l.line(d)})
	}
	l.popLoop()
	l.patch(skip, l.here())
	l.emit(Instr{Op: OpPopLoop, Line: l.line(d)})
	for _, pc := range lj.continues {
		l.patch(pc, start)
	}
	for _, pc := range lj.breaks {
		l.patch(pc, l.here())
	}
}

func (l *lowerer) pushLoop() *loopJumps {
	lj := &loopJumps{}
	l.loops = append(l.loops, lj)
	return lj
}

func (l *lowerer) popLoop() {
	if len(l.loops) > 0 {
		l.loops = l.loops[:len(l.loops)-1]
	}
}

func boolBit(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func (l *lowerer) forInStmt(d *ast.ForInStmt) {
	l.expr(d.Iter)
	slot := l.frame.declareLocal(d.VarName)
	l.emit(Instr{Op: OpPushLoop, Line: l.line(d)})

	l.emit(Instr{Op: OpMakeStepRange, A: slot, Line: l.line(d)})
	start := l.here()
	skip := l.emit(Instr{Op: OpJumpIfFalse, A: -1, Line: l.line(d)})
	lj := l.pushLoop()
	l.emit(Instr{Op: OpLocalSet, A: slot, Line: l.line(d)})
	if d.Filter != nil {

		l.expr(d.Filter)
		drop := l.emit(Instr{Op: OpJumpIfFalse, A: -1, Line: l.line(d)})
		l.block(d.Body)
		l.emit(Instr{Op: OpJump, A: start, Line: l.line(d)})
		l.patch(drop, l.here())
		l.emit(Instr{Op: OpJump, A: start, Line: l.line(d)})
	} else {
		l.block(d.Body)
		l.emit(Instr{Op: OpJump, A: start, Line: l.line(d)})
	}
	l.popLoop()
	l.patch(skip, l.here())
	l.emit(Instr{Op: OpPopLoop, Line: l.line(d)})
	for _, pc := range lj.continues {
		l.patch(pc, start)
	}
	for _, pc := range lj.breaks {
		l.patch(pc, l.here())
	}
}

func (l *lowerer) talignStmt(d *ast.TalignStmt) {

	l.emit(Instr{Op: OpTalignArranged, A: -1, Line: l.line(d)})
	l.expr(d.X)
	l.emit(Instr{Op: OpPop, Line: l.line(d)})
}

func (l *lowerer) tryStmt(d *ast.TryStmt) {
	flag := l.frame.declareLocal("__koshish_raise__")
	fail := l.frame.declareLocal("__koshish_khata__")

	push := l.emit(Instr{Op: OpPushHandler, A: -1, Line: l.line(d)})
	l.block(d.Body)
	l.emit(Instr{Op: OpPopHandler, Line: l.line(d)})

	l.emit(Instr{Op: OpConst, A: l.p.AddConst(Const{Kind: ConstBool, B: false}), Line: l.line(d)})
	l.emit(Instr{Op: OpLocalSet, A: flag, Line: l.line(d)})
	toFinally := l.emit(Instr{Op: OpJump, A: -1, Line: l.line(d)})

	l.patch(push, l.here())
	l.emit(Instr{Op: OpLocalSet, A: fail, Line: l.line(d)})
	if d.HasCatch {

		slot := l.frame.declareLocal(d.Catch.Name)
		l.emit(Instr{Op: OpLocalGet, A: fail, Line: l.line(d)})
		l.emit(Instr{Op: OpCatch, A: slot, Line: l.line(d)})
		l.block(d.Catch.Body)
	}

	raised := Const{Kind: ConstBool, B: true}
	if d.HasCatch {
		raised = Const{Kind: ConstBool, B: false}
	}
	l.emit(Instr{Op: OpConst, A: l.p.AddConst(raised), Line: l.line(d)})
	l.emit(Instr{Op: OpLocalSet, A: flag, Line: l.line(d)})

	finallyAt := l.here()
	l.patch(toFinally, finallyAt)
	if d.Finally != nil {
		l.block(d.Finally)
	}

	l.emit(Instr{Op: OpLocalGet, A: flag, Line: l.line(d)})
	skipThrow := l.emit(Instr{Op: OpJumpIfFalse, A: -1, Line: l.line(d)})
	l.emit(Instr{Op: OpLocalGet, A: fail, Line: l.line(d)})
	l.emit(Instr{Op: OpThrow, Line: l.line(d)})
	l.patch(skipThrow, l.here())
}

func arithOp(k token.Kind) (Op, bool) {
	switch k {
	case "+=":
		return OpAdd, true
	case "-=":
		return OpSub, true
	case "*=":
		return OpMul, true
	case "/=":
		return OpDiv, true
	case "%=":
		return OpRem, true
	case "**=":
		return OpPow, true
	}
	return OpInvalid, false
}
