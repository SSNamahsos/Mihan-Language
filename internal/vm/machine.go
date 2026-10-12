package vm

import (
	"io"

	"mihan/internal/eval"
	"mihan/internal/ir"
	rt "mihan/internal/runtime"
)

type frame struct {
	fn     *ir.Func
	slots  int32
	ret    int32
	caller *frame

	upvals []*Value

	cells []*Value

	ctor *rt.Object
}

type iter struct {
	cur  int64
	step int64
	last int64
	done bool
	list *rt.List
	at   int
	str  string
	sat  int
}

type handler struct {
	catch   int32
	frames  int
	stack   int
	slots   int
	scopes  int
	catches bool
}

const maxDepth = 2000

type Machine struct {
	p      *ir.Program
	fns    []*ir.Func
	stack  []Value
	slots  []Value
	frames []*frame
	loops  []int32

	code []ir.Instr
	pc   int32
	f    *frame

	globals []Value
	gslot   []int32

	host *eval.Interp

	hostArgs []string
	hostIn   io.Reader

	scopes []int32

	handlers []handler

	out    io.Writer
	builts []Value

	classes   map[string]*rt.Class
	classInfo map[*rt.Class]*classRuntime
	structs   map[string]bool
}

func New(p *ir.Program, out io.Writer) *Machine {
	m := &Machine{
		p:       p,
		fns:     p.Funcs,
		globals: make([]Value, 0, 64),
		gslot:   make([]int32, len(p.Names)),
		out:     out,
		builts:  make([]Value, len(p.Names)),
	}
	for i := range m.gslot {
		m.gslot[i] = -1
	}
	m.stack = make([]Value, 0, 256)
	m.slots = make([]Value, 0, 1024)
	return m
}

func (m *Machine) SetHostArgs(args []string, in io.Reader) {
	m.hostArgs = args
	if in != nil {
		m.hostIn = in
	}
	if m.host != nil {
		m.host.Args = args
		if in != nil {
			m.host.In = in
		}
	}
}

func (m *Machine) hostInterp() *eval.Interp {
	if m.host == nil {
		m.host = eval.New(m.out, m.hostIn)
		m.host.Args = m.hostArgs
		m.host.MaxSteps = 0

		if m.p != nil {
			m.host.File = m.p.Source
		}
	}
	return m.host
}

func (m *Machine) Run() error {
	if m.p == nil || m.p.Top == nil {
		return &fault{code: "E5001",
			fa: "این برنامه چیزی برای اجرا ندارد",
			en: "this program has nothing to run"}
	}
	m.f = &frame{fn: m.p.Top, slots: int32(len(m.slots))}
	m.frames = append(m.frames, m.f)
	m.slots = growSlots(m.slots, m.p.Top.Locals)
	m.code, m.pc = m.p.Top.Code, 0
	for {
		done, err := m.step()
		if err != nil {

			if handled, herr := m.handle(err); herr != nil {
				return herr
			} else if handled {
				continue
			}
			return err
		}
		if done {
			return nil
		}
	}
}

func (m *Machine) step() (done bool, err error) {
	if int(m.pc) >= len(m.code) {

		return m.doReturn(Nil(), m.f)
	}
	in := m.code[m.pc]
	m.pc++
	f := m.f

	switch in.Op {

	case ir.OpConst:
		m.push(m.constant(in.A))

	case ir.OpGlobalDefine, ir.OpGlobalSet:
		m.setGlobal(in.A, m.pop())
	case ir.OpGlobalGet:
		m.push(m.getGlobal(in.A))
	case ir.OpName:
		m.push(m.nameValue(in.A))
	case ir.OpLocalDefine, ir.OpLocalSet:
		v := m.pop()

		if c := f.cellOf(in.A); c != nil {
			*c = v
		} else {
			m.slots[f.slots+in.A] = v
		}
	case ir.OpLocalGet:
		if c := f.cellOf(in.A); c != nil {
			m.push(*c)
		} else {
			m.push(m.slots[f.slots+in.A])
		}
	case ir.OpUpvalueDefine, ir.OpUpvalueSet:
		m.setUpval(f, in.A, m.pop())
	case ir.OpUpvalueGet:
		m.push(m.getUpval(f, in.A))
	case ir.OpDefineFunc:
		v, fc := m.defineFunc(in.A, in.B)
		if fc != nil {
			return false, m.at(fc)
		}
		if _, isLocal := f.localSlot(in.A); isLocal {

			m.push(v)
			return false, nil
		}

		m.setGlobal(in.A, v)

	case ir.OpMakeFunc:
		v, fc := m.makeFunc(in.A)
		if fc != nil {
			return false, m.at(fc)
		}
		m.push(v)

	case ir.OpDefineClass, ir.OpDefineStruct:
		m.initClasses()
		v, fc := m.defineClass(in.A, in.B, in.Op == ir.OpDefineStruct)
		if fc != nil {
			return false, m.at(fc)
		}
		m.storeDefined(in.A, v)
	case ir.OpDefineEnum:
		m.initClasses()
		v, fc := m.defineEnum(in.A, in.B)
		if fc != nil {
			return false, m.at(fc)
		}
		m.storeDefined(in.A, v)
	case ir.OpDefineInterface:

	case ir.OpGetField:
		recv := m.pop()
		v, fc := m.getField(recv, in.A)
		if fc != nil {
			return false, m.at(fc)
		}
		m.push(v)
	case ir.OpSetField:

		val, recv := m.pop(), m.pop()
		if fc := m.setField(recv, in.A, val); fc != nil {
			return false, m.at(fc)
		}

	case ir.OpTruthy:
		m.push(Bool(m.pop().Truthy()))
	case ir.OpNot:
		m.push(Bool(!m.pop().Truthy()))
	case ir.OpBitNot:
		n, ok := m.pop().AsInt()
		if !ok {
			return false, m.bad(0, "این عدد نیست", "that is not a number")
		}
		m.push(Int(^n))
	case ir.OpNeg:
		v := m.pop()
		u, fc := unaryMinus(v)
		if fc != nil {
			return false, m.at(fc)
		}
		m.push(u)
	case ir.OpToStr:
		m.push(Str(m.pop().Text()))
	case ir.OpTypeOf:
		m.push(Str(m.pop().TypeName()))

	case ir.OpAdd, ir.OpSub, ir.OpMul, ir.OpDiv, ir.OpRem, ir.OpPow,
		ir.OpIDiv, ir.OpBitAnd, ir.OpBitOr, ir.OpBitXor, ir.OpShift:

		n := len(m.stack)
		if n < 2 {
			return false, m.bad(0,
				"این حساب دو عدد می‌خواهد اما روی پشته این‌قدر عدد نبود",
				"this arithmetic needs two numbers but the stack did not have them")
		}
		v, fc := arith(in.Op, m.stack[n-2], m.stack[n-1])
		if fc != nil {
			return false, m.at(fc)
		}
		m.stack[n-2] = v
		m.stack = m.stack[:n-1]

	case ir.OpEq, ir.OpNe, ir.OpLt, ir.OpLe, ir.OpGt, ir.OpGe:
		n := len(m.stack)
		if n < 2 {
			return false, m.bad(0,
				"این مقایسه دو مقدار می‌خواهد اما روی پشته این‌قدر مقدار نبود",
				"this comparison needs two values but the stack did not have them")
		}
		v, fc := compareOp(in.Op, m.stack[n-2], m.stack[n-1])
		if fc != nil {
			return false, m.at(fc)
		}
		m.stack[n-2] = v
		m.stack = m.stack[:n-1]

	case ir.OpIn:
		coll, item := m.pop(), m.pop()
		v, fc := contains(coll, item)
		if fc != nil {
			return false, m.at(fc)
		}
		m.push(v)

	case ir.OpInterpolate:
		m.push(m.interpolate(int(in.A)))

	case ir.OpSlice:
		hi, lo, coll := m.pop(), m.pop(), m.pop()
		v, fc := sliceOf(coll, lo, hi, in.B != 0)
		if fc != nil {
			return false, m.at(fc)
		}
		m.push(v)

	case ir.OpMakeList:
		m.push(List(rt.NewList(toRTs(m.popN(int(in.A)))...)))
	case ir.OpMakeSet:
		s := rt.NewSet()
		for _, v := range m.popN(int(in.A)) {
			s.Add(v.ToRT())
		}
		m.push(Set(s))
	case ir.OpMakeDict:
		m.push(m.makeDict(int(in.A)))
	case ir.OpMakeRange:
		m.push(makeIter(m.pop(), m.pop(), in.B != 0))
	case ir.OpMakeStepRange:
		m.push(m.makeStepIter(m.pop()))
	case ir.OpAppend:
		val, coll := m.pop(), m.pop()
		if fc := appendTo(coll, val); fc != nil {
			return false, m.at(fc)
		}
		m.push(coll)
	case ir.OpIndex:
		key, coll := m.pop(), m.pop()
		v, fc := index(coll, key)
		if fc != nil {
			return false, m.at(fc)
		}
		m.push(v)
	case ir.OpSetIndex:
		val, key, coll := m.pop(), m.pop(), m.pop()
		if fc := setIndex(coll, key, val); fc != nil {
			return false, m.at(fc)
		}

	case ir.OpJump:
		m.pc = in.A
	case ir.OpJumpIfFalse:
		return m.jumpIfFalse(f, in.A)
	case ir.OpJumpIfTrue:
		if m.pop().Truthy() {
			m.pc = in.A
		}
	case ir.OpJumpIfNotNil:
		if !m.pop().IsNil() {
			m.pc = in.A
		}

	case ir.OpPushScope:
		m.scopes = append(m.scopes, int32(len(m.scopes)))
	case ir.OpPopScope:
		m.popScope()
	case ir.OpPushLoop:
		m.loops = append(m.loops, m.pc)
	case ir.OpPopLoop:
		if n := len(m.loops); n > 0 {
			m.loops = m.loops[:n-1]
		}
	case ir.OpPop:
		m.pop()

	case ir.OpReturn:
		return m.doReturn(m.pop(), f)
	case ir.OpReturnNil:
		return m.doReturn(Nil(), f)

	case ir.OpCall:
		v, entered, fc, err := m.call(int(in.A))
		if err != nil {
			return false, err
		}
		if fc != nil {
			return false, m.at(fc)
		}
		if !entered {
			m.push(v)
		}
	case ir.OpCallMethod:

		args := m.popN(int(in.A))
		recv := m.pop()
		v, entered, fc, err := m.callMethod(m.p.Names[in.B], recv, args)
		if err != nil {
			return false, err
		}
		if fc != nil {
			return false, m.at(fc)
		}
		if !entered {
			m.push(v)
		}
	case ir.OpGetMethod:

	case ir.OpCallBuiltin:
		m.push(m.pop())
	case ir.OpNop:

	case ir.OpPushHandler:

		m.handlers = append(m.handlers, handler{
			catch:   in.A,
			frames:  len(m.frames),
			stack:   len(m.stack),
			slots:   len(m.slots),
			scopes:  len(m.scopes),
			catches: in.A >= 0 && int(in.A) < len(m.code),
		})
	case ir.OpPopHandler:
		if n := len(m.handlers); n > 0 {
			m.handlers = m.handlers[:n-1]
		}
	case ir.OpThrow:
		v := m.pop()
		if v.tag != TagError {

			v = FromRT(rt.NewPanic(v.ToRT()).Value)
		}
		return false, m.throw(v)
	case ir.OpCatch:

		if n := len(m.stack); n > 0 {
			m.setLocal(f, in.A, m.pop())
		} else {
			m.setLocal(f, in.A, Nil())
		}

	default:
		return false, m.missing(in.Op, in.Line)
	}
	return false, nil
}

func (m *Machine) setLocal(f *frame, slot int32, v Value) {
	if c := f.cellOf(slot); c != nil {
		*c = v
		return
	}
	if int(f.slots)+int(slot) < len(m.slots) {
		m.slots[f.slots+slot] = v
	}
}

func (m *Machine) throw(v Value) error {
	return &fault{code: "E5001", threw: &v,
		fa: "برنامه در همین‌جا ایستاد",
		en: "the program stopped here"}
}

func (m *Machine) handle(err error) (bool, error) {
	fc, ok := err.(*fault)
	if !ok {
		return false, err
	}
	var v Value
	if fc.threw != nil {
		v = *fc.threw
	} else {

		v = FromRT(rt.NewPanic(rt.Str(fc.en)).Value)
	}
	if caught, herr := m.unwind(v, fc); caught {
		return true, herr
	}

	if ev, ok := v.p.(*rt.ErrorValue); ok {
		fc.en = string(ev.Message)
	}
	return false, fc
}

func (m *Machine) unwind(v Value, fc *fault) (bool, error) {
	for len(m.handlers) > 0 {
		h := m.handlers[len(m.handlers)-1]
		m.handlers = m.handlers[:len(m.handlers)-1]

		if h.frames > len(m.frames) {
			continue
		}
		if h.stack <= len(m.stack) {
			m.stack = m.stack[:h.stack]
		}
		if h.slots <= len(m.slots) {
			m.slots = m.slots[:h.slots]
		}
		if h.scopes <= len(m.scopes) {
			m.scopes = m.scopes[:h.scopes]
		}
		if h.frames > 0 {
			m.frames = m.frames[:h.frames]
			m.f = m.frames[h.frames-1]
			m.code = m.f.fn.Code
		}

		m.pc = h.catch
		m.push(v)
		return true, nil
	}
	return false, fc
}
