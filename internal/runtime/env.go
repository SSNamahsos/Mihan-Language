package runtime

import "mihan/internal/diag"

type Env struct {
	small []entry

	ie     [3]entry
	ib     [3]Binding
	nb     uint8
	vars   map[string]*Binding
	parent *Env

	turn uint64

	Name string

	Pos diag.Pos

	block bool
}

type entry struct {
	name string
	b    *Binding
}

const smallMax = 8

func (e *Env) find(name string) *Binding {
	if e.vars != nil {
		return e.vars[name]
	}
	for i := range e.small {
		if e.small[i].name == name {
			return e.small[i].b
		}
	}
	return nil
}

func (e *Env) put(name string, b *Binding) {
	if e.small == nil && e.vars == nil {
		e.small = e.ie[:0]
	}
	if e.vars == nil && len(e.small) < smallMax {
		e.small = append(e.small, entry{name, b})
		return
	}
	if e.vars == nil {
		e.vars = make(map[string]*Binding, 2*smallMax)
		for _, en := range e.small {
			e.vars[en.name] = en.b
		}
		e.small = nil
	}
	e.vars[name] = b
}

type Binding struct {
	Value Value

	Sabit bool

	TypeName string

	Used bool

	turn uint64
}

func NewEnv(parent *Env) *Env {
	return &Env{parent: parent}
}

func NewBlockEnv(parent *Env) *Env {
	e := NewEnv(parent)
	e.block = true
	return e
}

func (e *Env) Home() *Env {
	for e.block && e.parent != nil {
		e = e.parent
	}
	return e
}

func (e *Env) IsBlock() bool { return e.block }

func (e *Env) Root() *Env {
	for e.parent != nil {
		e = e.parent
	}
	return e
}

func (e *Env) Lookup(name string) (*Binding, *Env) {
	for s := e; s != nil; s = s.parent {
		if b := s.find(name); b != nil && b.turn == s.turn {
			return b, s
		}
	}
	return nil, nil
}

func (e *Env) declare(name string, v Value, typeName string, used, sabit bool) *Binding {
	if b := e.find(name); b != nil {
		b.Value = v
		b.TypeName = typeName
		b.Used = used
		b.Sabit = sabit
		b.turn = e.turn
		return b
	}
	var b *Binding
	if int(e.nb) < len(e.ib) {
		b = &e.ib[e.nb]
		e.nb++
	} else {
		b = &Binding{}
	}
	*b = Binding{Value: v, TypeName: typeName, Used: used, Sabit: sabit, turn: e.turn}
	e.put(name, b)
	return b
}

func (e *Env) Clear() { e.turn++ }

func (e *Env) Declare(name string, v Value) *Binding {
	return e.declare(name, v, "", true, false)
}

func (e *Env) DeclareUnset(name, typeName string) *Binding {
	return e.declare(name, Nil, typeName, false, false)
}

func (e *Env) DeclareTyped(name string, v Value, typeName string) *Binding {
	return e.declare(name, v, typeName, true, false)
}

func (e *Env) DeclareSabit(name string, v Value, typeName string) *Binding {
	return e.declare(name, v, typeName, true, true)
}

func (e *Env) Set(name string, v Value) bool {
	if b, _ := e.Lookup(name); b != nil {
		b.Value = v
		b.Used = true
		return true
	}
	return false
}

func (e *Env) Names() []string {
	out := make([]string, 0, len(e.vars)+len(e.small))
	for n, b := range e.vars {
		if b.turn == e.turn {
			out = append(out, n)
		}
	}
	for _, en := range e.small {
		if en.b.turn == e.turn {
			out = append(out, en.name)
		}
	}
	return out
}

type Ctrl uint8

const (
	CtrlNone Ctrl = iota

	CtrlReturn

	CtrlBreak

	CtrlContinue

	CtrlFatal
)

type Flow struct {
	Kind Ctrl
	Val  Value
}

var (
	flowNone     = Flow{Kind: CtrlNone, Val: Nil}
	flowBreak    = Flow{Kind: CtrlBreak, Val: Nil}
	flowContinue = Flow{Kind: CtrlContinue, Val: Nil}
	flowFatal    = Flow{Kind: CtrlFatal, Val: Nil}
)

func Return(v Value) Flow { return Flow{Kind: CtrlReturn, Val: v} }

func Fatal(v Value) Flow { return Flow{Kind: CtrlFatal, Val: v} }

type Panic struct {
	Value Value

	Trace []Frame
}

func (p *Panic) Error() string {
	if e, ok := p.Value.(*ErrorValue); ok {
		return string(e.Message)
	}
	return Show(p.Value)
}

func NewPanic(v Value) *Panic {
	if e, ok := v.(*ErrorValue); ok {
		_ = e
		return &Panic{Value: v}
	}
	return &Panic{Value: &ErrorValue{Message: Str(Show(v)), Data: v}}
}

func AsPanic(err error) (*Panic, bool) {
	p, ok := err.(*Panic)
	return p, ok
}

type NamedArg struct {
	name    string
	value   Value
	nameSet bool
}

func NewNamedArg(name string, v Value) NamedArg {
	return NamedArg{name: name, value: v, nameSet: name != ""}
}

func (a NamedArg) Name() string { return a.name }

func (a NamedArg) Value() Value { return a.value }

func (a NamedArg) IsNamed() bool { return a.nameSet }

type Frame struct {
	Name string
	File string
	Pos  diag.Pos

	Callee string
}

func (f Frame) String() string {
	if f.Name == "" {
		return fmtPos(f)
	}
	return f.Name + " at " + fmtPos(f)
}

func fmtPos(f Frame) string {
	if f.File == "" {
		return f.Pos.LineString()
	}
	return f.File + ":" + f.Pos.LineString()
}
