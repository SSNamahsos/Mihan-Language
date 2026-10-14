package vm

import (
	"fmt"
	"strconv"

	"mihan/internal/eval"
	"mihan/internal/ir"
	rt "mihan/internal/runtime"
)

func growSlots(s []Value, n int) []Value {
	need := len(s) + n
	if need <= cap(s) {
		return s[:need]
	}
	out := make([]Value, need, need*2)
	copy(out, s)
	for i := len(s); i < need; i++ {
		out[i] = Nil()
	}
	return out
}

func (m *Machine) push(v Value) { m.stack = append(m.stack, v) }

func (m *Machine) pop() Value {
	n := len(m.stack)
	if n == 0 {
		return Nil()
	}
	v := m.stack[n-1]
	m.stack = m.stack[:n-1]
	return v
}

func (m *Machine) popN(n int) []Value {
	if n <= 0 {
		return nil
	}
	base := len(m.stack) - n
	if base < 0 {
		base = 0
	}
	out := make([]Value, len(m.stack)-base)
	copy(out, m.stack[base:])
	m.stack = m.stack[:base]
	return out
}

func toRTs(vs []Value) []rt.Value {
	out := make([]rt.Value, len(vs))
	for i, v := range vs {
		out[i] = v.ToRT()
	}
	return out
}

func (m *Machine) constant(i int32) Value {
	if int(i) < 0 || int(i) >= len(m.p.Consts) {
		return Nil()
	}
	c := m.p.Consts[i]
	switch c.Kind {
	case ir.ConstInt:
		return Int(c.I)
	case ir.ConstRune:
		return Str(string(rune(c.I)))
	case ir.ConstFloat:
		return Float(c.F)
	case ir.ConstBool:
		return Bool(c.B)
	case ir.ConstStr:
		return Str(c.S)
	}
	return Nil()
}

func (m *Machine) setGlobal(name int32, v Value) {
	i := int(name)
	if i < 0 || i >= len(m.gslot) {
		return
	}
	if m.gslot[i] < 0 {
		m.gslot[i] = int32(len(m.globals))
		m.globals = append(m.globals, v)
		return
	}
	m.globals[m.gslot[i]] = v
}

func (m *Machine) getGlobal(name int32) Value {
	i := int(name)
	if i < 0 || i >= len(m.gslot) {
		return Nil()
	}
	if s := m.gslot[i]; s >= 0 {
		return m.globals[s]
	}
	if b := m.builtin(name); b != nil {
		if b.tag == TagNone {
			return Nil()
		}
		return *b
	}
	return Nil()
}

func (m *Machine) builtin(name int32) *Value {
	if int(name) < 0 || int(name) >= len(m.p.Names) {
		return nil
	}
	if int(name) >= len(m.builts) {
		return nil
	}

	if m.builts[name].tag != TagNil {
		return &m.builts[name]
	}
	n := m.p.Names[name]
	if _, ok := builtins[n]; !ok {

		f := aliasOf(n)
		if f != "" {
			n = f
		}
		if _, machineHas := builtins[n]; !machineHas {
			if !m.hostKnows(n) {
				m.builts[name] = Value{tag: TagNone}
				return nil
			}
		}
	}

	m.builts[name] = Value{tag: TagBuiltin, s: n}
	return &m.builts[name]
}

func (m *Machine) hostKnows(name string) bool {
	for _, b := range eval.Builtins() {
		if b.Name == name || b.Alias == name {
			return true
		}
	}
	return false
}

func (m *Machine) nameValue(i int32) Value {
	if int(i) < 0 || int(i) >= len(m.p.Names) {
		return Nil()
	}
	return Str(m.p.Names[i])
}

func (m *Machine) setUpval(f *frame, i int32, v Value) {
	if c := f.upvalue(i); c != nil {
		*c = v
	}
}

func (f *frame) upvalue(i int32) *Value {
	if int(i) < len(f.upvals) {
		return f.upvals[i]
	}
	return nil
}

func (m *Machine) getUpval(f *frame, i int32) Value {
	if c := f.upvalue(i); c != nil {
		return *c
	}
	return Nil()
}

func (m *Machine) globalSlotOf(name int32) (int32, bool) {
	if int(name) < 0 || int(name) >= len(m.gslot) {
		return 0, false
	}
	s := m.gslot[name]
	return s, s >= 0
}

func (m *Machine) nameText(i int32) string {
	if int(i) < 0 || int(i) >= len(m.p.Names) {
		return ""
	}
	return m.p.Names[i]
}

func (m *Machine) popScope() {
	if n := len(m.scopes); n > 0 {
		m.scopes = m.scopes[:n-1]
	}
}

func (m *Machine) doReturn(v Value, f *frame) (bool, error) {
	m.slots = m.slots[:f.slots]
	if len(m.slots) < int(f.slots) {
		m.slots = m.slots[:0]
	}
	if f.ctor != nil {
		v = Object(f.ctor)
	}

	for len(m.handlers) > 0 && m.handlers[len(m.handlers)-1].frames > len(m.frames)-1 {
		m.handlers = m.handlers[:len(m.handlers)-1]
	}
	if f.caller == nil {
		return true, nil
	}
	m.frames = m.frames[:len(m.frames)-1]
	m.f = f.caller
	m.code, m.pc = f.caller.fn.Code, f.ret
	m.push(v)
	return false, nil
}

func (m *Machine) at(f *fault) error {
	if f.line == 0 {
		f.line = m.line()
	}
	return f
}

func (m *Machine) bad(line int32, fa, en string) error {
	return &fault{code: "E4000", fa: fa, en: en, line: m.line()}
}

func (m *Machine) line() int32 {
	if m.f == nil {
		return 0
	}
	if p := int(m.pc) - 1; p >= 0 && p < len(m.code) {
		return m.code[p].Line
	}
	return 0
}

func (m *Machine) missing(op ir.Op, line int32) error {
	name := op.String()
	if name == "" {
		name = strconv.Itoa(int(op))
	}
	return &fault{code: "E5900", line: line,
		fa: "in mashin hanooz '" + name + "' ro balad nist",
		en: "this machine does not do " + name + " yet"}
}

func itoa(v int) string { return fmt.Sprint(v) }
