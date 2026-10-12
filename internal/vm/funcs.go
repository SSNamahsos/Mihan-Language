package vm

func (m *Machine) defineFunc(name, fn int32) (Value, *fault) {
	if int(fn) < 0 || int(fn) >= len(m.fns) {
		return Nil(), &fault{code: "E4002",
			fa: "دستورهای این تابع پیدا نشد",
			en: "the instructions for this function were not found"}
	}
	f := m.fns[fn]
	cl := &funcVal{ir: f, name: m.nameText(name), home: m}
	if len(f.UpvalNames) == 0 {
		return Value{tag: TagFunc, p: cl}, nil
	}
	cl.upvals = make([]*Value, len(f.UpvalNames))
	for i, n := range f.UpvalNames {
		cell := m.capture(n)
		if cell == nil {
			return Nil(), &fault{code: "E4003",
				fa: "این تابع نامی را می‌خواهد که بیرون از آن پیدا نمی‌شود: " + m.nameText(n),
				en: "this function wants a name that is not in sight: " + m.nameText(n)}
		}
		cl.upvals[i] = cell
	}
	return Value{tag: TagFunc, p: cl}, nil
}

func (m *Machine) makeFunc(fn int32) (Value, *fault) {
	if int(fn) < 0 || int(fn) >= len(m.fns) {
		return Nil(), &fault{code: "E4002",
			fa: "دستورهای این تابع پیدا نشد",
			en: "the instructions for this function were not found"}
	}
	f := m.fns[fn]
	cl := &funcVal{ir: f, name: f.Name, home: m}
	if len(f.UpvalNames) == 0 {
		return Value{tag: TagFunc, p: cl}, nil
	}
	cl.upvals = make([]*Value, len(f.UpvalNames))
	for i, n := range f.UpvalNames {
		cell := m.capture(n)
		if cell == nil {
			return Nil(), &fault{code: "E4003",
				fa: "این تابع نامی را می‌خواهد که بیرون از آن پیدا نمی‌شود: " + m.nameText(n),
				en: "this function wants a name that is not in sight: " + m.nameText(n)}
		}
		cl.upvals[i] = cell
	}
	return Value{tag: TagFunc, p: cl}, nil
}

func (m *Machine) capture(name int32) *Value {
	for i := len(m.frames) - 1; i >= 0; i-- {
		f := m.frames[i]
		if slot, ok := f.localSlot(name); ok {
			return f.cellAt(int32(slot), m)
		}
	}

	if s, ok := m.globalSlotOf(name); ok {
		return &m.globals[s]
	}
	return nil
}

func (f *frame) localSlot(name int32) (int32, bool) {
	for i, n := range f.fn.LocalNames {
		if n == name {
			return int32(i), true
		}
	}
	return 0, false
}

func (f *frame) cellAt(slot int32, m *Machine) *Value {
	if f.cells == nil {
		f.cells = make([]*Value, len(f.fn.LocalNames))
	}
	if int(slot) >= len(f.cells) {
		return nil
	}
	if c := f.cells[slot]; c != nil {
		return c
	}
	c := new(Value)
	if int(f.slots)+int(slot) < len(m.slots) {
		*c = m.slots[f.slots+slot]
	}
	f.cells[slot] = c
	return c
}

func (m *Machine) enterFunc(f *funcVal, args []Value) (Value, *fault, error) {
	if len(m.frames) >= maxDepth {
		return Nil(), nil, m.bad(0,
			"فراخوانی‌ها خیلی عمیق شده‌اند؛ شاید تابعی بازگشتی بی‌پایان است",
			"the calls have gone too deep, perhaps a recursive function never ends")
	}
	if f.ir == nil {
		return Nil(), nil, m.bad(0,
			"این تابع هنوز دستوری ندارد",
			"this function has no instructions")
	}

	nf := &frame{
		fn:     f.ir,
		slots:  int32(len(m.slots)),
		ret:    m.pc,
		caller: m.f,
		upvals: f.upvals,
	}
	m.slots = growSlots(m.slots, f.ir.Locals)
	for i, a := range args {
		if i < f.ir.Params && i < f.ir.Locals {
			m.slots[nf.slots+int32(i)] = a
		}
	}
	m.frames = append(m.frames, nf)
	m.f = nf
	m.code, m.pc = f.ir.Code, 0
	return Nil(), nil, nil
}

func (f *frame) cellOf(slot int32) *Value {
	if int(slot) < len(f.cells) {
		return f.cells[slot]
	}
	return nil
}

func (m *Machine) callGo(f *funcVal, args []Value) (Value, *fault, error) {
	if f == nil || f.ir == nil {
		return Nil(), nil, m.bad(0,
			"این تابع دستوری ندارد که اجرا شود",
			"this function has no instructions to run")
	}

	savedCode, savedPC, savedF := m.code, m.pc, m.f
	savedStack, savedFrames, savedSlots := len(m.stack), len(m.frames), len(m.slots)
	savedLoops, savedScopes := len(m.loops), len(m.scopes)
	defer func() {
		m.code, m.pc, m.f = savedCode, savedPC, savedF
		m.stack = m.stack[:savedStack]
		m.frames = m.frames[:savedFrames]
		m.slots = m.slots[:savedSlots]
		m.loops = m.loops[:savedLoops]
		m.scopes = m.scopes[:savedScopes]
	}()

	if _, fc, err := m.enterFunc(f, args); fc != nil || err != nil {
		return Nil(), fc, err
	}

	depth := len(m.frames)
	for len(m.frames) >= depth {
		done, err := m.step()
		if err != nil {

			return Nil(), nil, err
		}
		if done {
			break
		}
	}

	if len(m.stack) > savedStack {
		v := m.stack[len(m.stack)-1]
		return v, nil, nil
	}
	return Nil(), nil, nil
}
