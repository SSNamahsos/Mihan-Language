package eval

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"unsafe"

	"mihan/internal/ast"
	"mihan/internal/diag"
	rt "mihan/internal/runtime"
	"mihan/internal/token"
)

type ctl uint8

const (
	cNone ctl = iota
	cReturn
	cBreak
	cContinue
)

type frame struct {
	in      *Interp
	globals *rt.Env
	slots   []rt.Value
	ret     rt.Value
	inline  [6]rt.Value
	big     []rt.Value
}

func newFrame(in *Interp, globals *rt.Env, n int) *frame {
	var fr *frame
	if k := len(in.framePool); k > 0 {
		fr = in.framePool[k-1]
		in.framePool = in.framePool[:k-1]
		fr.globals = globals
	} else {
		fr = &frame{in: in, globals: globals}
	}
	if n <= len(fr.inline) {
		fr.slots = fr.inline[:n]
	} else {
		if cap(fr.big) < n {
			fr.big = make([]rt.Value, n)
		}
		fr.slots = fr.big[:n]
	}
	return fr
}

func freeFrame(in *Interp, fr *frame) {
	clear(fr.slots)
	fr.ret = nil
	fr.globals = nil
	if len(in.framePool) < 1024 {
		in.framePool = append(in.framePool, fr)
	}
}

type cexpr func(fr *frame) (rt.Value, error)
type cstmt func(fr *frame) (ctl, error)

type compiledFunc struct {
	n      int
	body   cstmt
	method bool
	expr   bool
}

type compiledStmt struct {
	n    int
	body cstmt
}

const (
	boxLo = -1024
	boxHi = 1 << 15
)

var (
	boxOnce sync.Once
	boxed   []rt.Value
)

func boxInt(i rt.Int) rt.Value {
	if i >= boxLo && i < boxHi {
		return boxed[i-boxLo]
	}
	return i
}

func initBoxes() {
	boxOnce.Do(func() {
		boxed = make([]rt.Value, boxHi-boxLo)
		for i := range boxed {
			boxed[i] = rt.Int(i + boxLo)
		}
	})
}

var (
	funcCache sync.Map
	stmtCache sync.Map
)

func (in *Interp) fastEnabled() bool { return in.Hooks == nil && in.MaxSteps == 0 }

type funcKey struct {
	node   ast.Node
	method bool
}

func compiledFuncFor(d *ast.FuncDecl) *compiledFunc {
	return cachedFunc(funcKey{d, false}, d.Params, d.Body, nil)
}

func compiledMethodFor(d *ast.FuncDecl) *compiledFunc {
	return cachedFunc(funcKey{d, true}, d.Params, d.Body, nil)
}

func compiledLambdaFor(x *ast.Lambda) *compiledFunc {
	return cachedFunc(funcKey{x, false}, x.Params, x.Body, x.Expr)
}

func cachedFunc(k funcKey, params []*ast.Param, body *ast.Block, expr ast.Expr) *compiledFunc {
	if v, ok := funcCache.Load(k); ok {
		cf, _ := v.(*compiledFunc)
		return cf
	}
	initBoxes()
	cf := compileFunc(params, body, expr, k.method)
	funcCache.Store(k, cf)
	return cf
}

func compiledStmtFor(st ast.Stmt) *compiledStmt {
	if v, ok := stmtCache.Load(st); ok {
		cs, _ := v.(*compiledStmt)
		return cs
	}
	initBoxes()
	c := &compiler{}
	body := c.stmt(st)
	var cs *compiledStmt
	if !c.bad && body != nil {
		cs = &compiledStmt{n: c.n, body: body}
	}
	stmtCache.Store(st, cs)
	return cs
}

func (in *Interp) runCompiledStmt(cs *compiledStmt, env *rt.Env) (rt.Flow, error) {
	fr := newFrame(in, env, cs.n)
	c, err := cs.body(fr)
	ret := fr.ret
	freeFrame(in, fr)
	if err != nil {
		return rt.Fatal(rt.Nil), err
	}
	switch c {
	case cReturn:
		return rt.Return(ret), nil
	case cBreak:
		return rt.Flow{Kind: rt.CtrlBreak, Val: rt.Nil}, nil
	case cContinue:
		return rt.Flow{Kind: rt.CtrlContinue, Val: rt.Nil}, nil
	}
	return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil
}

func (in *Interp) callFast(node ast.Node, pos diag.Pos, fn *rt.Function, cf *compiledFunc, fr *frame) (rt.Value, error) {
	if cf.expr {
		c, err := cf.body(fr)
		ret := fr.ret
		freeFrame(in, fr)
		if err != nil {
			return nil, err
		}
		if c == cReturn {
			return ret, nil
		}
		return rt.Nil, nil
	}
	in.Depth++
	in.Trace = append(in.Trace, rt.Frame{Name: fn.Name, File: in.File, Pos: pos, Callee: fn.Name})
	c, err := cf.body(fr)
	in.Trace = in.Trace[:len(in.Trace)-1]
	in.Depth--
	ret := fr.ret
	freeFrame(in, fr)
	if err != nil {
		return nil, err
	}
	if c == cReturn {
		return ret, nil
	}
	return rt.Nil, nil
}

func (in *Interp) tryFastCall(node ast.Node, fn *rt.Function, recv *rt.Object, args []rt.Value) (rt.Value, bool, error) {
	cf, _ := fn.Fast.(*compiledFunc)
	if cf == nil || (recv != nil) != cf.method || !in.fastEnabled() || len(args) > len(fn.Params) {
		return nil, false, nil
	}
	for _, a := range args {
		if _, ok := a.(rt.NamedArg); ok {
			return nil, false, nil
		}
	}
	for i, p := range fn.Params {
		if i >= len(args) && p.Default == nil {
			return nil, false, nil
		}
	}
	fr := newFrame(in, fn.Env, cf.n)
	off := 0
	if cf.method {
		fr.slots[0] = recv
		off = 1
	}
	for i, p := range fn.Params {
		if i < len(args) {
			fr.slots[off+i] = args[i]
			continue
		}
		v, err := in.eval(p.Default, fn.Env)
		if err != nil {
			freeFrame(in, fr)
			return nil, true, err
		}
		fr.slots[off+i] = v
	}
	v, err := in.callFast(node, posOf(node), fn, cf, fr)
	return v, true, err
}

type compiler struct {
	method bool
	params int
	fn     bool
	locals map[string]int
	lex    []map[string]int
	n      int
	bad    bool
}

func (c *compiler) newSlot() int { c.n++; return c.n - 1 }

func compileFunc(params []*ast.Param, body *ast.Block, expr ast.Expr, method bool) *compiledFunc {
	c := &compiler{fn: true, method: method, locals: map[string]int{}}
	if method {
		c.locals["__khod__"] = c.newSlot()
	}
	for _, p := range params {
		if _, dup := c.locals[p.Name]; dup || p.Name == "khod" {
			return nil
		}
		c.locals[p.Name] = c.newSlot()
	}
	c.params = c.n
	ok := true
	note := func(name string) {
		if name == "khod" || name == "__khod__" {
			ok = false
		}
		if _, have := c.locals[name]; !have {
			c.locals[name] = c.newSlot()
		}
	}
	var run cstmt
	if expr != nil {
		collectAssignedIn(expr, note)
		if !ok {
			return nil
		}
		e := c.expr(expr)
		if c.bad {
			return nil
		}
		run = func(fr *frame) (ctl, error) {
			v, err := e(fr)
			if err != nil {
				return cNone, err
			}
			fr.ret = v
			return cReturn, nil
		}
	} else {
		if body == nil {
			return nil
		}
		collectAssigned(body, note)
		if !ok {
			return nil
		}
		run = c.block(body)
		if c.bad {
			return nil
		}
	}
	return &compiledFunc{n: c.n, body: run, method: method, expr: expr != nil}
}

func collectAssignedIn(e ast.Expr, f func(string)) {
	walkNodes(reflect.ValueOf(e), func(v reflect.Value) bool {
		if n, ok := v.Interface().(*ast.Assign); ok {
			if id, ok := n.Target.(*ast.Ident); ok {
				f(id.Name)
			}
		}
		return true
	})
}

func collectAssigned(b *ast.Block, f func(string)) {
	walkNodes(reflect.ValueOf(b), func(v reflect.Value) bool {
		switch n := v.Interface().(type) {
		case *ast.Assign:
			if id, ok := n.Target.(*ast.Ident); ok {
				f(id.Name)
			}
		case *ast.VarDecl:
			f(n.Name)
		}
		return true
	})
}

type nameRef struct {
	kind int
	slot int
}

func (c *compiler) resolve(name string) nameRef {
	for i := len(c.lex) - 1; i >= 0; i-- {
		if s, ok := c.lex[i][name]; ok {
			return nameRef{0, s}
		}
	}
	if s, ok := c.locals[name]; ok {
		return nameRef{1, s}
	}
	return nameRef{2, -1}
}

func (c *compiler) fail() cstmt { c.bad = true; return nil }

func (c *compiler) block(b *ast.Block) cstmt {
	if b == nil {
		return func(*frame) (ctl, error) { return cNone, nil }
	}
	if !plainBlock(b) {
		return c.fail()
	}
	stmts := make([]cstmt, 0, len(b.Stmts))
	for _, st := range b.Stmts {
		s := c.stmt(st)
		if c.bad {
			return nil
		}
		stmts = append(stmts, s)
	}
	switch len(stmts) {
	case 1:
		return stmts[0]
	case 2:
		a, b := stmts[0], stmts[1]
		return func(fr *frame) (ctl, error) {
			if k, err := a(fr); k != cNone || err != nil {
				return k, err
			}
			return b(fr)
		}
	}
	return func(fr *frame) (ctl, error) {
		for _, s := range stmts {
			if k, err := s(fr); k != cNone || err != nil {
				return k, err
			}
		}
		return cNone, nil
	}
}

func (c *compiler) stmt(st ast.Stmt) cstmt {
	switch s := st.(type) {
	case *ast.ExprStmt:
		e := c.expr(s.X)
		if c.bad {
			return nil
		}
		return func(fr *frame) (ctl, error) {
			_, err := e(fr)
			return cNone, err
		}
	case *ast.Block:
		return c.block(s)
	case *ast.VarDecl:
		if s.Sabit || s.TypeName != "" {
			return c.fail()
		}
		return c.varDecl(s)
	case *ast.IfStmt:
		cond := c.cond(s.Cond)
		then := c.block(s.Then)
		var els cstmt
		if s.Else != nil {
			els = c.stmt(s.Else)
		}
		if c.bad {
			return nil
		}
		return func(fr *frame) (ctl, error) {
			t, err := cond(fr)
			if err != nil {
				return cNone, err
			}
			if t {
				return then(fr)
			}
			if els != nil {
				return els(fr)
			}
			return cNone, nil
		}
	case *ast.WhileStmt:
		cond := c.cond(s.Cond)
		body := c.block(s.Body)
		if c.bad {
			return nil
		}
		return func(fr *frame) (ctl, error) {
			for {
				t, err := cond(fr)
				if err != nil {
					return cNone, err
				}
				if !t {
					return cNone, nil
				}
				k, err := body(fr)
				if err != nil {
					return cNone, err
				}
				switch k {
				case cBreak:
					return cNone, nil
				case cReturn:
					return k, nil
				}
			}
		}
	case *ast.ForStmt:
		return c.forRange(s)
	case *ast.ForInStmt:
		return c.forIn(s)
	case *ast.ReturnStmt:
		if s.Value == nil {
			return func(fr *frame) (ctl, error) { fr.ret = rt.Nil; return cReturn, nil }
		}
		e := c.expr(s.Value)
		if c.bad {
			return nil
		}
		return func(fr *frame) (ctl, error) {
			v, err := e(fr)
			if err != nil {
				return cNone, err
			}
			fr.ret = v
			return cReturn, nil
		}
	case *ast.TryStmt:
		return c.try(s)
	case *ast.BreakStmt:
		return func(*frame) (ctl, error) { return cBreak, nil }
	case *ast.ContinueStmt:
		return func(*frame) (ctl, error) { return cContinue, nil }
	case *ast.ThrowStmt:
		e := c.expr(s.Value)
		if c.bad {
			return nil
		}
		return func(fr *frame) (ctl, error) {
			v, err := e(fr)
			if err != nil {
				return cNone, err
			}
			return cNone, fr.in.panicAt(s, v)
		}
	}
	return c.fail()
}

func (c *compiler) forRange(s *ast.ForStmt) cstmt {
	lo := c.expr(s.Lo)
	hi := c.expr(s.Hi)
	slot := c.newSlot()
	c.lex = append(c.lex, map[string]int{s.VarName: slot})
	var filter cexpr
	if s.Filter != nil {
		filter = c.expr(s.Filter)
	}
	body := c.block(s.Body)
	c.lex = c.lex[:len(c.lex)-1]
	if c.bad {
		return nil
	}
	inclusive := s.Inclusive
	return func(fr *frame) (ctl, error) {
		loV, err := lo(fr)
		if err != nil {
			return cNone, err
		}
		hiV, err := hi(fr)
		if err != nil {
			return cNone, err
		}
		l, ok1 := toInt(loV)
		h, ok2 := toInt(hiV)
		if !ok1 || !ok2 {
			return cNone, fr.in.faultAt(s, "E4005",
				"do sar-e baze bayad adad bashan", "both ends of a har range must be whole numbers")
		}
		step := rt.Int(1)
		if l > h {
			step = -1
		}
		last := h
		if inclusive {
			last = h + step
		}
		for i := l; i != last; i += step {
			fr.slots[slot] = boxInt(i)
			if filter != nil {
				keep, err := filter(fr)
				if err != nil {
					return cNone, err
				}
				if !rt.Truthy(keep) {
					continue
				}
			}
			k, err := body(fr)
			if err != nil {
				return cNone, err
			}
			switch k {
			case cBreak:
				return cNone, nil
			case cReturn:
				return k, nil
			}
		}
		return cNone, nil
	}
}

func (c *compiler) try(s *ast.TryStmt) cstmt {
	body := c.block(s.Body)
	var fin, catch cstmt
	if s.Finally != nil {
		fin = c.block(s.Finally)
	}
	slot := -1
	if s.HasCatch && s.Catch != nil {
		slot = c.newSlot()
		c.lex = append(c.lex, map[string]int{s.Catch.Name: slot})
		catch = c.block(s.Catch.Body)
		c.lex = c.lex[:len(c.lex)-1]
	}
	if c.bad {
		return nil
	}
	return func(fr *frame) (ctl, error) {
		k, err := body(fr)
		if err == nil {
			if fin != nil {
				if _, ferr := fin(fr); ferr != nil {
					return cNone, ferr
				}
			}
			return k, nil
		}
		p, ok := rt.AsPanic(err)
		if !ok || catch == nil {
			if fin != nil {
				if _, ferr := fin(fr); ferr != nil {
					return cNone, ferr
				}
			}
			return cNone, err
		}
		fr.slots[slot] = p.Value
		k, cerr := catch(fr)
		if fin != nil {
			if _, ferr := fin(fr); ferr != nil {
				return cNone, ferr
			}
		}
		if cerr != nil {
			return cNone, cerr
		}
		return k, nil
	}
}

func (c *compiler) forIn(s *ast.ForInStmt) cstmt {
	iter := c.expr(s.Iter)
	slot := c.newSlot()
	c.lex = append(c.lex, map[string]int{s.VarName: slot})
	var filter cexpr
	if s.Filter != nil {
		filter = c.expr(s.Filter)
	}
	body := c.block(s.Body)
	c.lex = c.lex[:len(c.lex)-1]
	if c.bad {
		return nil
	}
	return func(fr *frame) (ctl, error) {
		coll, err := iter(fr)
		if err != nil {
			return cNone, err
		}
		items, err := fr.in.iterate(s.Iter, coll)
		if err != nil {
			return cNone, err
		}
		for _, item := range items {
			fr.slots[slot] = item
			if filter != nil {
				keep, err := filter(fr)
				if err != nil {
					return cNone, err
				}
				if !rt.Truthy(keep) {
					continue
				}
			}
			k, err := body(fr)
			if err != nil {
				return cNone, err
			}
			switch k {
			case cBreak:
				return cNone, nil
			case cReturn:
				return k, nil
			}
		}
		return cNone, nil
	}
}

func (c *compiler) varDecl(s *ast.VarDecl) cstmt {
	var val cexpr
	if s.Value != nil {
		val = c.expr(s.Value)
	}
	if c.bad {
		return nil
	}
	id := &ast.Ident{Loc: s.Loc, Name: s.Name}
	store := c.storeIdent(id, func(fr *frame, b *rt.Binding) error {
		if b.Sabit && b.Used {
			return fr.in.faultAt(s, "E4003",
				"`"+s.Name+"` ba `sabit` sakhte shode va dige avaz nemishe",
				"%s was declared sabit and cannot be changed", s.Name)
		}
		return nil
	})
	return func(fr *frame) (ctl, error) {
		var v rt.Value = rt.Nil
		if val != nil {
			var err error
			if v, err = val(fr); err != nil {
				return cNone, err
			}
		}
		return cNone, store(fr, v)
	}
}

type gslot struct {
	root *rt.Env
	b    *rt.Binding
}

type gref struct {
	name string
	c    atomic.Pointer[gslot]
}

func (g *gref) get(fr *frame) *rt.Binding {
	if s := g.c.Load(); s != nil && s.root == fr.globals {
		return s.b
	}
	b, at := fr.globals.Lookup(g.name)

	if b != nil && at == fr.globals && !at.IsBlock() {
		g.c.Store(&gslot{root: fr.globals, b: b})
	}
	return b
}

func (c *compiler) ident(x *ast.Ident) cexpr {
	if c.method && (x.Name == "khod" || x.Name == "__khod__") {
		return func(fr *frame) (rt.Value, error) { return fr.slots[0], nil }
	}
	if x.Name == "khod" || x.Name == "__khod__" {
		return func(fr *frame) (rt.Value, error) { return fr.in.evalIdent(x, fr.globals) }
	}
	ref := c.resolve(x.Name)
	g := &gref{name: x.Name}
	global := func(fr *frame) (rt.Value, error) {
		if b := g.get(fr); b != nil && (b.Used || b.Value != nil) {
			return b.Value, nil
		}
		return fr.in.evalIdent(x, fr.globals)
	}
	if ref.kind == 1 && ref.slot < c.params {
		ref.kind = 0
	}
	switch ref.kind {
	case 0:
		s := ref.slot
		return func(fr *frame) (rt.Value, error) { return fr.slots[s], nil }
	case 1:
		s := ref.slot
		return func(fr *frame) (rt.Value, error) {
			if v := fr.slots[s]; v != nil {
				return v, nil
			}
			return global(fr)
		}
	}
	return global
}

func (c *compiler) storeIdent(x *ast.Ident, check func(*frame, *rt.Binding) error) func(*frame, rt.Value) error {
	ref := c.resolve(x.Name)
	g := &gref{name: x.Name}
	writeGlobal := func(fr *frame, v rt.Value, b *rt.Binding) error {
		if check != nil {
			if err := check(fr, b); err != nil {
				return err
			}
		} else if b.Sabit && b.Used {
			return fr.in.store(x, v, fr.globals)
		}
		b.Value = v
		b.Used = true
		return nil
	}
	switch ref.kind {
	case 0:
		s := ref.slot
		return func(fr *frame, v rt.Value) error { fr.slots[s] = v; return nil }
	case 1:
		s := ref.slot
		return func(fr *frame, v rt.Value) error {
			if fr.slots[s] != nil {
				fr.slots[s] = v
				return nil
			}
			if b := g.get(fr); b != nil {
				return writeGlobal(fr, v, b)
			}
			fr.slots[s] = v
			return nil
		}
	}
	return func(fr *frame, v rt.Value) error {
		if b := g.get(fr); b != nil {
			return writeGlobal(fr, v, b)
		}
		fr.globals.Declare(x.Name, v)
		return nil
	}
}

func constant(v rt.Value) cexpr { return func(*frame) (rt.Value, error) { return v, nil } }

func (c *compiler) failE() cexpr { c.bad = true; return nil }

func (c *compiler) expr(e ast.Expr) cexpr {
	if c.bad {
		return nil
	}
	switch x := e.(type) {
	case *ast.IntLit:
		return constant(rt.Value(rt.Int(x.Value)))
	case *ast.FloatLit:
		return constant(rt.Value(rt.Float(x.Value)))
	case *ast.StrLit:
		return constant(rt.Value(rt.Str(x.Value)))
	case *ast.CharLit:
		return constant(rt.Value(rt.Str(string(x.Value))))
	case *ast.BoolLit:
		return constant(rt.Value(rt.Bool(x.Value)))
	case *ast.HichLit:
		return constant(rt.Value(rt.Nil))
	case *ast.Ident:
		return c.ident(x)
	case *ast.Group:
		return c.expr(x.X)
	case *ast.InterpLit:
		return c.interp(x)
	case *ast.ListLit:
		elems := c.exprs(x.Elems)
		if c.bad {
			return nil
		}
		return func(fr *frame) (rt.Value, error) {
			items := make([]rt.Value, len(elems))
			for i, el := range elems {
				v, err := el(fr)
				if err != nil {
					return nil, err
				}
				items[i] = v
			}
			return rt.NewList(items...), nil
		}
	case *ast.DictLit:
		return c.dict(x)
	case *ast.Unary:
		v := c.expr(x.X)
		if c.bad {
			return nil
		}
		if x.Op == "na" || x.Op == "!" {
			return func(fr *frame) (rt.Value, error) {
				a, err := v(fr)
				if err != nil {
					return nil, err
				}
				return rt.Bool(!rt.Truthy(a)), nil
			}
		}
		return func(fr *frame) (rt.Value, error) {
			a, err := v(fr)
			if err != nil {
				return nil, err
			}
			return fr.in.unaryOp(x, a)
		}
	case *ast.Binary:
		return c.binary(x)
	case *ast.Ternary:
		cond, then, els := c.expr(x.Cond), c.expr(x.Then), c.expr(x.Else)
		if c.bad {
			return nil
		}
		return func(fr *frame) (rt.Value, error) {
			v, err := cond(fr)
			if err != nil {
				return nil, err
			}
			if rt.Truthy(v) {
				return then(fr)
			}
			return els(fr)
		}
	case *ast.Assign:
		return c.assign(x)
	case *ast.RangeExpr:
		lo, hi := c.expr(x.Lo), c.expr(x.Hi)
		if c.bad {
			return nil
		}
		return func(fr *frame) (rt.Value, error) {
			l, err := lo(fr)
			if err != nil {
				return nil, err
			}
			h, err := hi(fr)
			if err != nil {
				return nil, err
			}
			li, ok1 := toInt(l)
			hiI, ok2 := toInt(h)
			if !ok1 || !ok2 {
				return nil, fr.in.faultAt(x, "E4103",
					"do sar-e baze bayad adad bashan", "both ends of a range must be whole numbers")
			}
			return &rt.Range{Lo: li, Hi: hiI, Inclusive: x.Inclusive}, nil
		}
	case *ast.Index:
		recv, at := c.expr(x.X), c.expr(x.At)
		if c.bad {
			return nil
		}
		return func(fr *frame) (rt.Value, error) {
			r, err := recv(fr)
			if err != nil {
				return nil, err
			}
			a, err := at(fr)
			if err != nil {
				return nil, err
			}
			if l, ok := r.(*rt.List); ok {
				if i, ok := a.(rt.Int); ok && i >= 0 && int(i) < len(l.Items) {
					return l.Items[i], nil
				}
			}
			return fr.in.getIndex(x, r, a)
		}
	case *ast.Member:
		return c.member(x)
	case *ast.Call:
		return c.call(x)
	case *ast.ListComp:
		return c.listComp(x)
	}
	return c.failE()
}

type ccond func(fr *frame) (bool, error)

func (c *compiler) cond(e ast.Expr) ccond {
	if g, ok := e.(*ast.Group); ok {
		return c.cond(g.X)
	}
	if x, ok := e.(*ast.Binary); ok {
		var cmp func(a, b rt.Int) bool
		switch x.Op {
		case "<":
			cmp = func(a, b rt.Int) bool { return a < b }
		case "<=":
			cmp = func(a, b rt.Int) bool { return a <= b }
		case ">":
			cmp = func(a, b rt.Int) bool { return a > b }
		case ">=":
			cmp = func(a, b rt.Int) bool { return a >= b }
		case "==":
			cmp = func(a, b rt.Int) bool { return a == b }
		case "!=":
			cmp = func(a, b rt.Int) bool { return a != b }
		}
		if cmp != nil {
			l, r := c.expr(x.L), c.expr(x.R)
			if c.bad {
				return nil
			}
			op := x.Op
			return func(fr *frame) (bool, error) {
				a, err := l(fr)
				if err != nil {
					return false, err
				}
				b, err := r(fr)
				if err != nil {
					return false, err
				}
				if ai, ok := a.(rt.Int); ok {
					if bi, ok := b.(rt.Int); ok {
						return cmp(ai, bi), nil
					}
				}
				v, err := fr.in.binaryOp(x, op, a, b)
				if err != nil {
					return false, err
				}
				return rt.Truthy(v), nil
			}
		}
	}
	v := c.expr(e)
	if c.bad {
		return nil
	}
	return func(fr *frame) (bool, error) {
		a, err := v(fr)
		if err != nil {
			return false, err
		}
		return rt.Truthy(a), nil
	}
}

func (c *compiler) exprs(es []ast.Expr) []cexpr {
	out := make([]cexpr, len(es))
	for i, e := range es {
		out[i] = c.expr(e)
	}
	return out
}

func (c *compiler) interp(x *ast.InterpLit) cexpr {
	type seg struct {
		text string
		e    cexpr
	}
	segs := make([]seg, len(x.Segs))
	for i, s := range x.Segs {
		if s.Expr == nil {
			segs[i] = seg{text: s.Text}
		} else {
			segs[i] = seg{e: c.expr(s.Expr)}
		}
	}
	if c.bad {
		return nil
	}
	return func(fr *frame) (rt.Value, error) {
		var b []byte
		for _, s := range segs {
			if s.e == nil {
				b = append(b, s.text...)
				continue
			}
			v, err := s.e(fr)
			if err != nil {
				return nil, err
			}
			b = append(b, rt.Show(v)...)
		}
		return rt.Str(b), nil
	}
}

func (c *compiler) dict(x *ast.DictLit) cexpr {
	keys := make([]cexpr, len(x.Entries))
	vals := make([]cexpr, len(x.Entries))
	for i, en := range x.Entries {
		keys[i] = c.expr(en.Key)
		vals[i] = c.expr(en.Value)
	}
	if c.bad {
		return nil
	}
	return func(fr *frame) (rt.Value, error) {
		d := rt.NewDict()
		for i := range keys {
			k, err := keys[i](fr)
			if err != nil {
				return nil, err
			}
			v, err := vals[i](fr)
			if err != nil {
				return nil, err
			}
			if _, ok := rt.HashKey(k); !ok {
				return nil, fr.in.faultAt(x.Entries[i].Key, "E4101",
					"kelid-e dict bayad adad, matn, bool ya hich bashe",
					"a dict key has to be a number, matn, bool or hich, not %s", rt.TypeName(k))
			}
			d.Set(k, v)
		}
		return d, nil
	}
}

func (c *compiler) listComp(x *ast.ListComp) cexpr {
	iter := c.expr(x.Iter)
	slot := c.newSlot()
	c.lex = append(c.lex, map[string]int{x.VarName: slot})
	var filter cexpr
	if x.Filter != nil {
		filter = c.expr(x.Filter)
	}
	elem := c.expr(x.Elem)
	c.lex = c.lex[:len(c.lex)-1]
	if c.bad {
		return nil
	}
	return func(fr *frame) (rt.Value, error) {
		coll, err := iter(fr)
		if err != nil {
			return nil, err
		}
		items, err := fr.in.iterate(x.Iter, coll)
		if err != nil {
			return nil, err
		}
		out := make([]rt.Value, 0, len(items))
		for _, item := range items {
			fr.slots[slot] = item
			if filter != nil {
				keep, err := filter(fr)
				if err != nil {
					return nil, err
				}
				if !rt.Truthy(keep) {
					continue
				}
			}
			v, err := elem(fr)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return rt.NewList(out...), nil
	}
}

func intOp(op token.Kind) (func(a, b rt.Int) (rt.Value, bool), bool) {
	switch op {
	case "+":
		return func(a, b rt.Int) (rt.Value, bool) {
			r := a + b
			if (a >= 0) == (b >= 0) && (r >= 0) != (a >= 0) {
				return nil, false
			}
			return boxInt(r), true
		}, true
	case "-":
		return func(a, b rt.Int) (rt.Value, bool) {
			r := a - b
			if (a >= 0) != (b >= 0) && (r >= 0) != (a >= 0) {
				return nil, false
			}
			return boxInt(r), true
		}, true
	case "*":
		return func(a, b rt.Int) (rt.Value, bool) {
			if a == 0 || b == 0 {
				return boxInt(0), true
			}
			r := a * b
			if r/b != a || (a == -1 && b == math.MinInt64) || (b == -1 && a == math.MinInt64) {
				return nil, false
			}
			return boxInt(r), true
		}, true
	case "<":
		return func(a, b rt.Int) (rt.Value, bool) { return rt.Bool(a < b), true }, true
	case "<=":
		return func(a, b rt.Int) (rt.Value, bool) { return rt.Bool(a <= b), true }, true
	case ">":
		return func(a, b rt.Int) (rt.Value, bool) { return rt.Bool(a > b), true }, true
	case ">=":
		return func(a, b rt.Int) (rt.Value, bool) { return rt.Bool(a >= b), true }, true
	case "==":
		return func(a, b rt.Int) (rt.Value, bool) { return rt.Bool(a == b), true }, true
	case "!=":
		return func(a, b rt.Int) (rt.Value, bool) { return rt.Bool(a != b), true }, true
	case "%":
		return func(a, b rt.Int) (rt.Value, bool) {
			if b == 0 || b == -1 {
				return nil, false
			}
			m := a % b
			if m != 0 && ((m < 0) != (b < 0)) {
				m += b
			}
			return boxInt(m), true
		}, true
	}
	return nil, false
}

func (c *compiler) binary(x *ast.Binary) cexpr {
	l, r := c.expr(x.L), c.expr(x.R)
	if c.bad {
		return nil
	}
	switch x.Op {
	case "va", "&&":
		return func(fr *frame) (rt.Value, error) {
			a, err := l(fr)
			if err != nil {
				return nil, err
			}
			if !rt.Truthy(a) {
				return rt.Bool(false), nil
			}
			b, err := r(fr)
			if err != nil {
				return nil, err
			}
			return rt.Bool(rt.Truthy(b)), nil
		}
	case "ya", "||":
		return func(fr *frame) (rt.Value, error) {
			a, err := l(fr)
			if err != nil {
				return nil, err
			}
			if rt.Truthy(a) {
				return rt.Bool(true), nil
			}
			b, err := r(fr)
			if err != nil {
				return nil, err
			}
			return rt.Bool(rt.Truthy(b)), nil
		}
	case "dar":
		return func(fr *frame) (rt.Value, error) {
			a, err := l(fr)
			if err != nil {
				return nil, err
			}
			b, err := r(fr)
			if err != nil {
				return nil, err
			}
			return fr.in.inOp(x, a, b)
		}
	}
	op := x.Op
	fast, _ := intOp(op)
	if op == "+" {
		return func(fr *frame) (rt.Value, error) {
			a, err := l(fr)
			if err != nil {
				return nil, err
			}
			b, err := r(fr)
			if err != nil {
				return nil, err
			}
			if ai, ok := a.(rt.Int); ok {
				if bi, ok := b.(rt.Int); ok {
					if v, ok := fast(ai, bi); ok {
						return v, nil
					}
				}
			} else if as, ok := a.(rt.Str); ok {
				switch bv := b.(type) {
				case rt.Str:
					return as + bv, nil
				case rt.Int:
					return rt.Str(strconv.AppendInt([]byte(as), int64(bv), 10)), nil
				}
			}
			return fr.in.binaryOp(x, op, a, b)
		}
	}
	if lit, ok := x.R.(*ast.IntLit); ok && fast != nil {
		k := rt.Int(lit.Value)
		kv := rt.Value(k)
		if id, ok := x.L.(*ast.Ident); ok {
			if ref := c.resolve(id.Name); ref.kind == 0 || (ref.kind == 1 && ref.slot < c.params) {
				sl := ref.slot
				return func(fr *frame) (rt.Value, error) {
					a := fr.slots[sl]
					if ai, ok := a.(rt.Int); ok {
						if v, ok := fast(ai, k); ok {
							return v, nil
						}
					}
					return fr.in.binaryOp(x, op, a, kv)
				}
			}
		}
		return func(fr *frame) (rt.Value, error) {
			a, err := l(fr)
			if err != nil {
				return nil, err
			}
			if ai, ok := a.(rt.Int); ok {
				if v, ok := fast(ai, k); ok {
					return v, nil
				}
			}
			return fr.in.binaryOp(x, op, a, kv)
		}
	}
	return func(fr *frame) (rt.Value, error) {
		a, err := l(fr)
		if err != nil {
			return nil, err
		}
		b, err := r(fr)
		if err != nil {
			return nil, err
		}
		if fast != nil {
			if ai, ok := a.(rt.Int); ok {
				if bi, ok := b.(rt.Int); ok {
					if v, ok := fast(ai, bi); ok {
						return v, nil
					}
				}
			}
		}
		return fr.in.binaryOp(x, op, a, b)
	}
}

func (c *compiler) assign(x *ast.Assign) cexpr {
	val := c.expr(x.Value)
	var cur cexpr
	compound := x.Op != "=" && x.Op != ":="
	if compound {
		cur = c.expr(x.Target)
	}
	var store func(*frame, rt.Value) error
	switch t := x.Target.(type) {
	case *ast.Ident:
		store = c.storeIdent(t, nil)
	case *ast.Index:
		recv, at := c.expr(t.X), c.expr(t.At)
		store = func(fr *frame, v rt.Value) error {
			r, err := recv(fr)
			if err != nil {
				return err
			}
			a, err := at(fr)
			if err != nil {
				return err
			}
			return fr.in.setIndex(t, r, a, v)
		}
	case *ast.Member:
		if t.Call {
			return c.failE()
		}
		recv := c.expr(t.X)
		store = func(fr *frame, v rt.Value) error {
			r, err := recv(fr)
			if err != nil {
				return err
			}
			return fr.in.setMember(t, r, t.Name, v)
		}
	default:
		return c.failE()
	}
	if c.bad {
		return nil
	}
	if !compound {
		return func(fr *frame) (rt.Value, error) {
			v, err := val(fr)
			if err != nil {
				return nil, err
			}
			if err := store(fr, v); err != nil {
				return nil, err
			}
			return rt.Nil, nil
		}
	}
	op := x.Op[:len(x.Op)-1]
	fast, _ := intOp(op)
	var app appender
	return func(fr *frame) (rt.Value, error) {
		a, err := cur(fr)
		if err != nil {
			return nil, err
		}
		b, err := val(fr)
		if err != nil {
			return nil, err
		}
		var m rt.Value
		ok := false
		if op == "+" {
			if as, isS := a.(rt.Str); isS {
				if bs, isS := b.(rt.Str); isS {
					m, ok = app.join(as, bs), true
				}
			}
		}
		if !ok && fast != nil {
			if ai, isI := a.(rt.Int); isI {
				if bi, isI := b.(rt.Int); isI {
					m, ok = fast(ai, bi)
				}
			}
		}
		if !ok {
			if m, err = fr.in.binaryOp(x, op, a, b); err != nil {
				return nil, err
			}
		}
		if err := store(fr, m); err != nil {
			return nil, err
		}
		return rt.Nil, nil
	}
}

type appender struct {
	buf []byte
}

func (ap *appender) join(a, b rt.Str) rt.Value {
	if len(b) == 0 {
		return a
	}
	if len(ap.buf) > 0 && len(a) == len(ap.buf) && unsafe.StringData(string(a)) == unsafe.SliceData(ap.buf) &&
		cap(ap.buf)-len(ap.buf) >= len(b) {
		ap.buf = append(ap.buf, b...)
		return rt.Str(unsafe.String(unsafe.SliceData(ap.buf), len(ap.buf)))
	}
	if len(a)+len(b) < 64 {
		return a + b
	}
	nb := make([]byte, 0, 2*(len(a)+len(b)))
	nb = append(nb, a...)
	nb = append(nb, b...)
	ap.buf = nb
	return rt.Str(unsafe.String(unsafe.SliceData(nb), len(nb)))
}

type methodCache struct {
	class   *rt.Class
	globals *rt.Env
	fn      *rt.Function
	cf      *compiledFunc
}

func (mc *methodCache) lookup(in *Interp, cl *rt.Class, name string) (*rt.Function, *compiledFunc) {
	if mc.class == cl && mc.globals == in.Globals {
		return mc.fn, mc.cf
	}
	m, owner := cl.FindMethod(name)
	if m == nil {
		return nil, nil
	}
	cf := compiledMethodFor(m)
	fn := &rt.Function{Name: m.Name, Params: m.Params, Body: m.Body, Env: in.Globals, Class: owner, Fast: cf}
	mc.class, mc.globals, mc.fn, mc.cf = cl, in.Globals, fn, cf
	return fn, cf
}

type carg struct {
	name string
	e    cexpr
}

func (c *compiler) args(as []ast.Arg) ([]carg, bool) {
	out := make([]carg, len(as))
	named := false
	for i, a := range as {
		out[i] = carg{a.Name, c.expr(a.Expr)}
		if a.Name != "" {
			named = true
		}
	}
	return out, named
}

func evalCArgs(fr *frame, as []carg) ([]rt.Value, error) {
	out := make([]rt.Value, len(as))
	for i, a := range as {
		v, err := a.e(fr)
		if err != nil {
			return nil, err
		}
		if a.name != "" {
			v = rt.NewNamedArg(a.name, v)
		}
		out[i] = v
	}
	return out, nil
}

func (c *compiler) member(x *ast.Member) cexpr {
	recv := c.expr(x.X)
	args, named := c.args(x.Args)
	if c.bad {
		return nil
	}

	direct := x.Call && !named && x.Name != "tule" && x.Name != "len"
	pos := x.Span().Start
	var mc methodCache
	return func(fr *frame) (rt.Value, error) {
		r, err := recv(fr)
		if err != nil {
			return nil, err
		}
		if direct {
			switch t := r.(type) {
			case *rt.Object:

				if _, isField := t.Get(x.Name); !isField && fr.in.fastEnabled() {
					if fn, cf := mc.lookup(fr.in, t.Class, x.Name); cf != nil && len(fn.Params) == len(args) {
						if fr.in.Depth > maxCallDepth {
							return nil, depthFault(fr.in, x)
						}
						nf := newFrame(fr.in, fn.Env, cf.n)
						nf.slots[0] = t
						for i, a := range args {
							v, err := a.e(fr)
							if err != nil {
								freeFrame(fr.in, nf)
								return nil, err
							}
							nf.slots[1+i] = v
						}
						return fr.in.callFast(x, pos, fn, cf, nf)
					}
				}
			case *rt.Dict:
				if dictMethods[x.Name] {
					vs, err := evalCArgs(fr, args)
					if err != nil {
						return nil, err
					}
					return fr.in.dictMethod(t, x.Name, vs)
				}
			case *rt.List:
				if listMethods[x.Name] {
					vs, err := evalCArgs(fr, args)
					if err != nil {
						return nil, err
					}
					return fr.in.listMethod(t, x.Name, vs)
				}
			case rt.Str:
				if stringMethods[x.Name] {
					vs, err := evalCArgs(fr, args)
					if err != nil {
						return nil, err
					}
					return fr.in.strMethod(t, x.Name, vs)
				}
			}
		}
		target, err := fr.in.readMember(x, r)
		if err != nil {
			return nil, err
		}
		if !x.Call {
			return target, nil
		}
		vs, err := evalCArgs(fr, args)
		if err != nil {
			return nil, err
		}
		switch target.(type) {
		case *rt.Function, *rt.BoundMethod, *rt.ClassValue:
		default:
			return nil, fr.in.faultAt(x, "E4154",
				fmt.Sprintf("'%s' yek %s ast, na funct-i ke beshe seda zad", x.Name, rt.TypeName(target)),
				"%s is a %s, not something to call; write %s without brackets", x.Name, rt.TypeName(target), x.Name)
		}
		return fr.in.callValue(x, target, nil, vs)
	}
}

func depthFault(in *Interp, x ast.Node) error {
	return in.faultAt(x, "E4150",
		"funct-ha bish az had hamdige ro seda zadan va man istadam",
		"the calls went on deeper than %d and I stopped, so this is probably a function that never ends", maxCallDepth)
}

func (c *compiler) call(x *ast.Call) cexpr {
	args, named := c.args(x.Args)
	id, isIdent := x.Fn.(*ast.Ident)
	if !isIdent {
		fn := c.expr(x.Fn)
		if c.bad {
			return nil
		}
		return func(fr *frame) (rt.Value, error) {
			if fr.in.Depth > maxCallDepth {
				return nil, depthFault(fr.in, x)
			}
			f, err := fn(fr)
			if err != nil {
				return nil, err
			}
			vs, err := evalCArgs(fr, args)
			if err != nil {
				return nil, err
			}
			return fr.in.callValue(x, f, nil, vs)
		}
	}
	if c.bad {
		return nil
	}
	ref := c.resolve(id.Name)
	if id.Name == "khod" || id.Name == "__khod__" {
		return c.failE()
	}
	g := &gref{name: id.Name}
	pos := x.Span().Start
	return func(fr *frame) (rt.Value, error) {
		in := fr.in
		if in.Depth > maxCallDepth {
			return nil, depthFault(in, x)
		}

		var small [6]rt.Value
		var vs []rt.Value
		if len(args) <= len(small) {
			vs = small[:len(args)]
		} else {
			vs = make([]rt.Value, len(args))
		}
		for i, a := range args {
			v, err := a.e(fr)
			if err != nil {
				return nil, err
			}
			if a.name != "" {
				v = rt.NewNamedArg(a.name, v)
			}
			vs[i] = v
		}
		var callee rt.Value
		found := false
		if ref.kind != 2 {
			if callee = fr.slots[ref.slot]; callee != nil {
				found = true
			}
		}
		if !found {
			if b := g.get(fr); b != nil {
				callee, found = b.Value, true
			}
		}
		if !found {
			own := append([]rt.Value(nil), vs...)
			if cl, ok := in.classes[id.Name]; ok {
				return in.constructNode(x, cl, own)
			}
			if fn, ok := in.builtins[id.Name]; ok {
				if !named {
					return fn.Native(in, own, x)
				}
				return in.callNative(x, fn, own)
			}
			return nil, in.faultAt(x, "E4151",
				fmt.Sprintf("'%s' ro hich ja nasakhtam", id.Name),
				"%s is not a name I know; it is not a variable, a funct, a darghan or a builtin", id.Name)
		}

		if fn, ok := callee.(*rt.Function); ok && !named && fn.Native == nil && fn.Expr == nil && len(fn.Params) == len(vs) {
			if cf, ok := fn.Fast.(*compiledFunc); ok && cf != nil && !cf.method && in.fastEnabled() {
				nf := newFrame(in, fn.Env, cf.n)
				copy(nf.slots, vs)
				return in.callFast(x, pos, fn, cf, nf)
			}
		}
		return in.callValue(x, callee, nil, append([]rt.Value(nil), vs...))
	}
}
