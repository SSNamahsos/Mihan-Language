package eval

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

type Interp struct {
	framePool []*frame

	Out io.Writer

	In io.Reader

	Args []string

	Globals *rt.Env

	File string

	Trace []rt.Frame

	Steps int

	MaxSteps int

	imps *importState

	classes  map[string]*rt.Class
	builtins map[string]*rt.Function

	reader *bufio.Reader

	Depth    int
	MaxDepth int

	Hooks *Hooks
}

type Hooks struct {
	OnStmt func(node ast.Node)

	OnCall func(node ast.Node)

	OnReturn func(node ast.Node)
}

func New(out io.Writer, in io.Reader) *Interp {
	if out == nil {
		out = os.Stdout
	}
	if in == nil {
		in = os.Stdin
	}
	inp := &Interp{
		Out:      out,
		In:       in,
		Globals:  rt.NewEnv(nil),
		imps:     &importState{loaded: map[string]*module{}},
		classes:  map[string]*rt.Class{},
		MaxDepth: 2000,
	}
	installBuiltins(inp)
	return inp
}

func (in *Interp) Class(name string) (*rt.Class, bool) {
	c, ok := in.classes[name]
	return c, ok
}

func (in *Interp) RegisterClass(c *rt.Class) { in.classes[c.Name] = c }

func (in *Interp) RunFile(f *ast.File) error {

	return in.run(f, false)
}

func (in *Interp) RunTopLevel(f *ast.File) error {
	return in.run(f, true)
}

func (in *Interp) run(f *ast.File, shared bool) (err error) {
	in.File = f.Path

	defer func() {
		if r := recover(); r != nil {
			p, ok := r.(*rt.Panic)
			if !ok {
				panic(r)
			}
			if len(p.Trace) == 0 {
				p.Trace = append(p.Trace, in.Trace...)
			}
			err = p
		}
	}()
	block := &ast.Block{Loc: ast.SpanOf(f.Span().Start, f.Span().End), Stmts: f.Decls}
	var flow rt.Flow
	var e error

	resolveFile(block)
	flow, e = in.execBlockIn(block, in.Globals)
	if e != nil {
		return e
	}
	if flow.Kind == rt.CtrlReturn {

		return nil
	}
	if flow.Kind == rt.CtrlFatal {
		if v, ok := flow.Val.(rt.Value); ok && v != nil {
			return rt.NewPanic(v)
		}
	}
	return nil
}

func (in *Interp) execBlock(b *ast.Block, env *rt.Env) (rt.Flow, error) {

	scope := env
	if !plainBlock(b) {
		scope = rt.NewBlockEnv(env)
	}
	for _, st := range b.Stmts {
		flow, err := in.execStmt(st, scope)
		if err != nil {
			return rt.Fatal(rt.Nil), err
		}
		if flow.Kind != rt.CtrlNone {
			return flow, nil
		}
	}
	return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil
}

func (in *Interp) execBlockIn(b *ast.Block, env *rt.Env) (rt.Flow, error) {
	if b == nil {
		return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil
	}
	for _, st := range b.Stmts {
		flow, err := in.execStmt(st, env)
		if err != nil {
			return rt.Fatal(rt.Nil), err
		}
		if flow.Kind != rt.CtrlNone {
			return flow, nil
		}
	}
	return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil
}

func (in *Interp) execStmt(st ast.Stmt, env *rt.Env) (rt.Flow, error) {
	if in.Hooks != nil && in.Hooks.OnStmt != nil {
		in.Hooks.OnStmt(st)
	}
	in.Steps++
	if in.MaxSteps > 0 && in.Steps > in.MaxSteps {
		return rt.Fatal(rt.Nil), in.faultAt(st, "E4001",
			"barname bish az had kar kard va motevaghef shod", "the program ran for too long and was stopped")
	}

	switch st.(type) {
	case *ast.ForStmt, *ast.WhileStmt, *ast.ForInStmt:

		if env == in.Globals && in.fastEnabled() {
			if cs := compiledStmtFor(st); cs != nil {
				return in.runCompiledStmt(cs, env)
			}
		}
	}

	switch s := st.(type) {
	case *ast.ExprStmt:
		if _, err := in.eval(s.X, env); err != nil {
			return rt.Fatal(rt.Nil), err
		}
		return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil

	case *ast.VarDecl:
		return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, in.execVarDecl(s, env)

	case *ast.Block:
		return in.execBlock(s, env)

	case *ast.IfStmt:
		return in.execIf(s, env)

	case *ast.WhileStmt:
		return in.execWhile(s, env)

	case *ast.ForStmt:
		return in.execFor(s, env)

	case *ast.ForInStmt:
		return in.execForIn(s, env)

	case *ast.FuncDecl:
		fn := &rt.Function{
			Name: s.Name, Params: s.Params, Body: s.Body, Env: env,
		}

		if in.fastEnabled() {
			if cf := compiledFuncFor(s); cf != nil {
				fn.Fast = cf
			}
		}
		if b, _ := env.Lookup(s.Name); b != nil {
			b.Value = fn
		} else {
			env.Declare(s.Name, fn)
		}
		return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil

	case *ast.ReturnStmt:
		if s.Value == nil {
			return rt.Return(rt.Nil), nil
		}
		v, err := in.eval(s.Value, env)
		if err != nil {
			return rt.Fatal(rt.Nil), err
		}
		return rt.Return(v), nil

	case *ast.BreakStmt:
		return rt.Flow{Kind: rt.CtrlBreak, Val: rt.Nil}, nil

	case *ast.ContinueStmt:
		return rt.Flow{Kind: rt.CtrlContinue, Val: rt.Nil}, nil

	case *ast.TalignStmt:

		return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil

	case *ast.ThrowStmt:
		v, err := in.eval(s.Value, env)
		if err != nil {
			return rt.Fatal(rt.Nil), err
		}
		return rt.Fatal(rt.Nil), in.panicAt(s, v)

	case *ast.TryStmt:
		return in.execTry(s, env)

	case *ast.ClassDecl:
		return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, in.execClassDecl(s, env)

	case *ast.StructDecl:
		return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, in.execStructDecl(s, env)

	case *ast.EnumDecl:
		return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, in.execEnumDecl(s, env)

	case *ast.InterfaceDecl:

		return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil

	case *ast.ImportDecl:

		return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, in.execImport(s, env)

	case *ast.ModuleDecl, *ast.ProgramDecl, *ast.EnumValue, *ast.FieldDecl:
		return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil
	}
	return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, in.faultAt(st, "E4002",
		"in dastoor ro hanooz balad nistam", "I do not know how to run this statement yet (%s)", st.NodeKind())
}

func (in *Interp) execVarDecl(s *ast.VarDecl, env *rt.Env) error {
	var v rt.Value = rt.Nil
	if s.Value != nil {
		var err error
		if v, err = in.eval(s.Value, env); err != nil {
			return err
		}
	} else if s.TypeName != "" {
		v = zeroOf(s.TypeName)
	}
	if s.Sabit {
		env.DeclareSabit(s.Name, v, s.TypeName)
		return nil
	}
	if s.TypeName != "" {
		conv, err := in.convertTo(s, v, s.TypeName)
		if err != nil {
			return err
		}
		env.DeclareTyped(s.Name, conv, s.TypeName)
		return nil
	}
	if existing, _ := env.Lookup(s.Name); existing != nil {

		if existing.Sabit && existing.Used {
			return in.faultAt(s, "E4003",
				"`"+s.Name+"` ba `sabit` sakhte shode va dige avaz nemishe",
				"%s was declared sabit and cannot be changed", s.Name)
		}
		env.Set(s.Name, v)
		return nil
	}
	env.Home().Declare(s.Name, v)
	return nil
}

func (in *Interp) execAssign(s *ast.Assign, env *rt.Env) error {
	if s.Op != "=" && s.Op != ":=" {
		cur, err := in.eval(s.Target, env)
		if err != nil {
			return err
		}
		rhs, err := in.eval(s.Value, env)
		if err != nil {
			return err
		}
		merged, err := in.binaryOp(s, s.Op[:len(s.Op)-1], cur, rhs)
		if err != nil {
			return err
		}
		return in.store(s.Target, merged, env)
	}
	v, err := in.eval(s.Value, env)
	if err != nil {
		return err
	}
	return in.store(s.Target, v, env)
}

func (in *Interp) store(target ast.Expr, v rt.Value, env *rt.Env) error {
	switch t := target.(type) {
	case *ast.Ident:
		b, _ := env.Lookup(t.Name)
		if b == nil {

			env.Home().Declare(t.Name, v)
			return nil
		}
		if b.Sabit && b.Used {
			return in.faultAt(target, "E4003",
				"`"+t.Name+"` ba `sabit` sakhte shode va dige avaz nemishe",
				"%s was declared sabit and cannot be changed", t.Name)
		}
		b.Value = v
		b.Used = true
		return nil
	case *ast.Index:
		recv, err := in.eval(t.X, env)
		if err != nil {
			return err
		}
		at, err := in.eval(t.At, env)
		if err != nil {
			return err
		}
		return in.setIndex(target, recv, at, v)
	case *ast.Member:
		recv, err := in.eval(t.X, env)
		if err != nil {
			return err
		}
		return in.setMember(target, recv, t.Name, v)
	}
	return in.faultAt(target, "E4004",
		"inja nemishe meghdar gozasht", "there is nowhere to put a value here")
}

func (in *Interp) execIf(s *ast.IfStmt, env *rt.Env) (rt.Flow, error) {
	cond, err := in.eval(s.Cond, env)
	if err != nil {
		return rt.Fatal(rt.Nil), err
	}
	if rt.Truthy(cond) {
		return in.execBlock(s.Then, env)
	}
	if s.Else != nil {
		return in.execStmt(s.Else, env)
	}
	return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil
}

func (in *Interp) execWhile(s *ast.WhileStmt, env *rt.Env) (rt.Flow, error) {

	shared := canShareScope(s.Body)
	reuse := rt.NewBlockEnv(env)
	for {
		cond, err := in.eval(s.Cond, env)
		if err != nil {
			return rt.Fatal(rt.Nil), err
		}
		if !rt.Truthy(cond) {
			return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil
		}
		body := reuse
		if shared {
			reuse.Clear()
		} else {
			body = rt.NewBlockEnv(env)
		}
		flow, err := in.execBlockIn(s.Body, body)
		if err != nil {
			return rt.Fatal(rt.Nil), err
		}
		switch flow.Kind {
		case rt.CtrlBreak:
			return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil
		case rt.CtrlReturn, rt.CtrlFatal:
			return flow, nil
		}
	}
}

func (in *Interp) execFor(s *ast.ForStmt, env *rt.Env) (rt.Flow, error) {
	loV, err := in.eval(s.Lo, env)
	if err != nil {
		return rt.Fatal(rt.Nil), err
	}
	hiV, err := in.eval(s.Hi, env)
	if err != nil {
		return rt.Fatal(rt.Nil), err
	}
	lo, ok1 := toInt(loV)
	hi, ok2 := toInt(hiV)
	if !ok1 || !ok2 {
		return rt.Fatal(rt.Nil), in.faultAt(s, "E4005",
			"do sar-e baze bayad adad bashan", "both ends of a har range must be whole numbers")
	}

	step := rt.Int(1)
	if lo > hi {
		step = rt.Int(-1)
	}
	last := hi
	if s.Inclusive {
		last = hi + step
	}

	shared := canShareScope(s.Body)
	reuse := rt.NewBlockEnv(env)
	for i := lo; i != last; i += step {
		body := reuse
		if shared {
			reuse.Clear()
		} else {
			body = rt.NewBlockEnv(env)
		}
		body.Declare(s.VarName, rt.Int(i))
		if s.Filter != nil {
			keep, err := in.eval(s.Filter, body)
			if err != nil {
				return rt.Fatal(rt.Nil), err
			}
			if !rt.Truthy(keep) {
				continue
			}
		}
		flow, err := in.execBlockIn(s.Body, body)
		if err != nil {
			return rt.Fatal(rt.Nil), err
		}
		switch flow.Kind {
		case rt.CtrlBreak:
			return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil
		case rt.CtrlReturn, rt.CtrlFatal:
			return flow, nil
		}
	}
	return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil
}

func (in *Interp) execForIn(s *ast.ForInStmt, env *rt.Env) (rt.Flow, error) {
	coll, err := in.eval(s.Iter, env)
	if err != nil {
		return rt.Fatal(rt.Nil), err
	}
	items, err := in.iterate(s.Iter, coll)
	if err != nil {
		return rt.Fatal(rt.Nil), err
	}
	shared := canShareScope(s.Body)
	reuse := rt.NewBlockEnv(env)
	for _, item := range items {
		body := reuse
		if shared {
			reuse.Clear()
		} else {
			body = rt.NewBlockEnv(env)
		}
		body.Declare(s.VarName, item)
		if s.Filter != nil {
			keep, err := in.eval(s.Filter, body)
			if err != nil {
				return rt.Fatal(rt.Nil), err
			}
			if !rt.Truthy(keep) {
				continue
			}
		}
		flow, err := in.execBlockIn(s.Body, body)
		if err != nil {
			return rt.Fatal(rt.Nil), err
		}
		switch flow.Kind {
		case rt.CtrlBreak:
			return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil
		case rt.CtrlReturn, rt.CtrlFatal:
			return flow, nil
		}
	}
	return rt.Flow{Kind: rt.CtrlNone, Val: rt.Nil}, nil
}

func (in *Interp) iterate(node ast.Node, v rt.Value) ([]rt.Value, error) {
	switch t := v.(type) {
	case *rt.List:
		out := make([]rt.Value, len(t.Items))
		copy(out, t.Items)
		return out, nil
	case *rt.Set:
		out := make([]rt.Value, t.Len())
		copy(out, t.Items())
		return out, nil
	case *rt.Dict:

		out := make([]rt.Value, t.Len())
		copy(out, t.Keys())
		return out, nil
	case rt.Str:
		s := string(t)
		n := rt.RuneCount(s)
		out := make([]rt.Value, n)
		for i := 0; i < n; i++ {
			r, _ := rt.RuneAt(s, i)
			out[i] = rt.Str(string(r))
		}
		return out, nil
	case *rt.Range:
		return rangeItems(*t), nil
	case *rt.StepRange:
		return stepRangeItems(*t), nil
	case rt.Hich, nil:
		return nil, in.faultAt(node, "E4006",
			"roye hich nemishe halghe zad", "you cannot walk hich; there is nothing in it")
	}
	return nil, in.faultAt(node, "E4007",
		fmt.Sprintf("%s ro nemishe too halghe charkhoond", rt.TypeName(v)),
		"a %s cannot be walked with har dar", rt.TypeName(v))
}

func rangeItems(r rt.Range) []rt.Value {
	var out []rt.Value
	step := rt.Int(1)
	if r.Lo > r.Hi {
		step = rt.Int(-1)
	}
	if r.Inclusive {
		for i := r.Lo; ; i += step {
			out = append(out, i)
			if i == r.Hi {
				break
			}
		}
		return out
	}
	for i := r.Lo; i != r.Hi; i += step {
		out = append(out, i)
	}
	return out
}

func (in *Interp) execTry(s *ast.TryStmt, env *rt.Env) (rt.Flow, error) {
	runFinally := func() error {
		if s.Finally == nil {
			return nil
		}
		_, err := in.execBlock(s.Finally, env)
		return err
	}

	flow, err := in.execBlock(s.Body, env)
	if err == nil {
		if ferr := runFinally(); ferr != nil {
			return rt.Fatal(rt.Nil), ferr
		}
		return flow, nil
	}
	p, ok := rt.AsPanic(err)
	if !ok || !s.HasCatch {

		if ferr := runFinally(); ferr != nil {
			return rt.Fatal(rt.Nil), ferr
		}
		return rt.Fatal(rt.Nil), err
	}

	catchScope := rt.NewBlockEnv(env)
	catchScope.Declare(s.Catch.Name, p.Value)

	caught, cerr := in.execBlock(s.Catch.Body, catchScope)
	if ferr := runFinally(); ferr != nil {
		return rt.Fatal(rt.Nil), ferr
	}
	if cerr != nil {
		return rt.Fatal(rt.Nil), cerr
	}
	return caught, nil
}

func (in *Interp) execClassDecl(s *ast.ClassDecl, env *rt.Env) error {
	cls := &rt.Class{
		Name:     s.Name,
		Doc:      s.Doc,
		Methods:  map[string]*ast.FuncDecl{},
		Defaults: map[string]rt.Value{},
	}
	for _, base := range s.Bases {
		parent, ok := in.classes[base]
		if !ok {
			return in.faultAt(s, "E4008",
				fmt.Sprintf("paye-ye '%s' ghabl az in sakhte nashode", base),
				"%s must be declared before it is used as a base", base)
		}
		cls.Parent = parent
		break
	}
	for _, f := range s.Fields {
		cls.Fields = append(cls.Fields, f)
		cls.Order = append(cls.Order, f.Name)
		var v rt.Value = rt.Nil
		if f.Default != nil {
			val, err := in.eval(f.Default, env)
			if err != nil {
				return err
			}
			v = val
		} else if f.TypeName != "" {
			v = zeroOf(f.TypeName)
		}
		cls.Defaults[rt.HashKeyMust(f.Name)] = v
	}
	for _, m := range s.Methods {
		cls.Methods[m.Name] = m
	}
	in.RegisterClass(cls)

	env.Declare(s.Name, &rt.ClassValue{Class: cls})
	return nil
}

func (in *Interp) execStructDecl(s *ast.StructDecl, env *rt.Env) error {
	cls := &rt.Class{
		Name:     s.Name,
		Doc:      s.Doc,
		Methods:  map[string]*ast.FuncDecl{},
		Defaults: map[string]rt.Value{},
	}
	for _, base := range s.Bases {
		if parent, ok := in.classes[base]; ok {
			cls.Parent = parent
			break
		}
	}
	for _, f := range s.Fields {
		cls.Fields = append(cls.Fields, f)
		cls.Order = append(cls.Order, f.Name)
		var v rt.Value = rt.Nil
		if f.Default != nil {
			val, err := in.eval(f.Default, env)
			if err != nil {
				return err
			}
			v = val
		} else if f.TypeName != "" {
			v = zeroOf(f.TypeName)
		}
		cls.Defaults[rt.HashKeyMust(f.Name)] = v
	}
	in.RegisterClass(cls)
	env.Declare(s.Name, &rt.ClassValue{Class: cls})
	return nil
}

func (in *Interp) execEnumDecl(s *ast.EnumDecl, env *rt.Env) error {
	cls := &rt.Class{
		Name:       s.Name,
		Doc:        s.Doc,
		Methods:    map[string]*ast.FuncDecl{},
		Defaults:   map[string]rt.Value{},
		Enumerated: true,
	}
	for _, ev := range s.Values {
		v, err := in.eval(ev.Value, env)
		if err != nil {
			return err
		}
		f := &ast.FieldDecl{Loc: ev.Loc, Name: ev.Name, NamePos: ev.NamePos, TypeName: s.Name}
		cls.Fields = append(cls.Fields, f)
		cls.Order = append(cls.Order, ev.Name)
		cls.Defaults[rt.HashKeyMust(ev.Name)] = v
	}
	in.RegisterClass(cls)
	env.Declare(s.Name, &rt.ClassValue{Class: cls})
	return nil
}

func (in *Interp) faultAt(node ast.Node, code, fa, enFormat string, args ...any) error {
	msg := rt.NewError(enFormat, args...)
	msg.Name = code
	msg.FA = persian(fa, args...)
	msg.Message = rt.Str(msg.FA)
	return in.panicAt(node, msg)
}

func persian(fa string, args ...any) string {
	if !strings.Contains(fa, "%") {
		return fa
	}
	return fmt.Sprintf(fa, args...)
}

func (in *Interp) faultFA(node ast.Node, code, fa string, faArgs []any, enFormat string, args ...any) error {

	msg := rt.NewError(enFormat, args...)
	msg.FA = fmt.Sprintf(fa, faArgs...)
	msg.Message = rt.Str(msg.FA)
	msg.Name = code
	return in.panicAt(node, msg)
}

func (in *Interp) faultNote(node ast.Node, code, fa, enFormat, note string, args ...any) error {
	err := in.faultAt(node, code, fa, enFormat, args...)
	if p, ok := rt.AsPanic(err); ok {
		if ev, ok := p.Value.(*rt.ErrorValue); ok {
			ev.Note = note
		}
	}
	return err
}

func native(f func(*Interp, []rt.Value, ast.Node) (rt.Value, error)) rt.NativeFn {
	return func(host any, args []rt.Value, node ast.Node) (rt.Value, error) {
		return f(host.(*Interp), args, node)
	}
}

func (in *Interp) panicAt(node ast.Node, v rt.Value) error {
	p := rt.NewPanic(v)
	frame := rt.Frame{Name: in.currentFunction(), File: in.File}
	if node != nil {
		frame.Pos = node.Span().Start
	}
	p.Trace = append([]rt.Frame{frame}, in.Trace...)
	return p
}

func (in *Interp) currentFunction() string {
	if n := len(in.Trace); n > 0 {
		return in.Trace[n-1].Name
	}
	return ""
}

func (in *Interp) RenderError(err error, src []byte) string {
	p, ok := rt.AsPanic(err)
	if !ok {
		return "khata: " + err.Error() + "\n  in eshkal az khod-e mihan bood, na az barname-ye to\n"
	}
	ev, isErr := p.Value.(*rt.ErrorValue)
	var b strings.Builder
	b.WriteString("khata: ")
	if isErr {
		b.WriteString(string(ev.Message))
	} else {
		b.WriteString(rt.Show(p.Value))
	}
	b.WriteByte('\n')

	if isErr && ev.Data != nil {
		if _, plain := ev.Data.(rt.Hich); !plain && ev.Data != nil {
			b.WriteString("  ")
			b.WriteString(rt.Show(ev.Data))
			b.WriteByte('\n')
		}
	}
	if len(p.Trace) > 0 {
		b.WriteString("  az koji: ")
		var frames []string
		for i, f := range p.Trace {
			if i == 0 {
				frames = append(frames, f.String())
				continue
			}
			frames = append(frames, "  az "+f.String())
		}
		b.WriteString(strings.Join(frames, "\n"))
		b.WriteByte('\n')
	}
	if isErr && src != nil {
		_ = src
	}
	return b.String()
}

func toInt(v rt.Value) (rt.Int, bool) {
	switch t := v.(type) {
	case rt.Int:
		return t, true
	case rt.Bool:
		if t {
			return 1, true
		}
		return 0, true
	case rt.Float:
		if float64(t) == math.Trunc(float64(t)) {
			return rt.Int(int64(t)), true
		}
	}
	return 0, false
}

func toFloat(v rt.Value) (float64, bool) {
	switch t := v.(type) {
	case rt.Int:
		return float64(t), true
	case rt.Float:
		return float64(t), true
	case rt.Bool:
		if t {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

func zeroOf(typeName string) rt.Value {
	base := typeName
	if i := strings.IndexByte(base, '<'); i > 0 {
		base = base[:i]
	}
	switch base {
	case "arz":
		return rt.Int(0)
	case "kasr":
		return rt.Float(0)
	case "matn":
		return rt.Str("")
	case "bool":
		return rt.Bool(false)
	case "harf":
		return rt.Str("")
	case "araye":
		return rt.NewList()
	case "list":
		return rt.NewList()
	case "dict":
		return rt.NewDict()
	case "set":
		return rt.NewSet()
	}
	return rt.Nil
}

func plainBlock(b *ast.Block) bool {
	for _, st := range b.Stmts {
		switch t := st.(type) {
		case *ast.ExprStmt, *ast.IfStmt, *ast.WhileStmt, *ast.ForStmt, *ast.ForInStmt,
			*ast.ReturnStmt, *ast.BreakStmt, *ast.ContinueStmt, *ast.ThrowStmt, *ast.TryStmt:
		case *ast.VarDecl:
			if t.Sabit || t.TypeName != "" {
				return false
			}
		default:
			return false
		}
	}
	return true
}
