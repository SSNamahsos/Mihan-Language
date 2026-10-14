package vm

import (
	"strconv"

	"mihan/internal/eval"
	"mihan/internal/ir"
	rt "mihan/internal/runtime"
)

type funcVal struct {
	ir     *ir.Func
	upvals []*Value
	bound  Value
	name   string

	home *Machine
}

type classVal struct {
	ir      *ir.Class
	methods map[string]*ir.Func
}

type builtinFn func(m *Machine, args []rt.Value) (rt.Value, error)

var builtins = map[string]builtinFn{}

var aliasOf = func(name string) string {
	for _, b := range eval.Builtins() {
		if b.Alias == name {
			return b.Name
		}
	}
	return ""
}

func init() {
	builtins["chap"] = func(m *Machine, args []rt.Value) (rt.Value, error) {
		for _, a := range args {
			if _, err := m.out.Write([]byte(rt.Show(a))); err != nil {
				return rt.Nil, err
			}
		}
		return rt.Nil, nil
	}
	builtins["chap_khat"] = func(m *Machine, args []rt.Value) (rt.Value, error) {
		for _, a := range args {
			if _, err := m.out.Write([]byte(rt.Show(a))); err != nil {
				return rt.Nil, err
			}
		}
		if _, err := m.out.Write([]byte("\n")); err != nil {
			return rt.Nil, err
		}
		return rt.Nil, nil
	}
	builtins["neshan"] = func(m *Machine, args []rt.Value) (rt.Value, error) {
		if len(args) != 1 {
			return rt.Nil, arity(1, len(args))
		}
		return rt.Str(rt.Show(args[0])), nil
	}
	builtins["tule"] = func(m *Machine, args []rt.Value) (rt.Value, error) {
		if len(args) != 1 {
			return rt.Nil, arity(1, len(args))
		}
		return length(FromRT(args[0])).ToRT(), nil
	}
	builtins["ghesmat"] = func(m *Machine, args []rt.Value) (rt.Value, error) {
		if len(args) != 2 {
			return rt.Nil, arity(2, len(args))
		}
		v, fc := arith(ir.OpIDiv, FromRT(args[0]), FromRT(args[1]))
		if fc != nil {
			return rt.Nil, fc
		}
		return v.ToRT(), nil
	}
	builtins["matn"] = func(m *Machine, args []rt.Value) (rt.Value, error) {
		if len(args) != 1 {
			return rt.Nil, arity(1, len(args))
		}
		return rt.Str(rt.Show(args[0])), nil
	}
	builtins["adad"] = func(m *Machine, args []rt.Value) (rt.Value, error) {
		if len(args) != 1 {
			return rt.Nil, arity(1, len(args))
		}
		return rt.Float(float64(toNum(args[0]))), nil
	}
	builtins["bool"] = func(m *Machine, args []rt.Value) (rt.Value, error) {
		if len(args) != 1 {
			return rt.Nil, arity(1, len(args))
		}
		return rt.Bool(FromRT(args[0]).Truthy()), nil
	}
}

func toNum(v rt.Value) float64 {
	switch x := v.(type) {
	case rt.Int:
		return float64(x)
	case rt.Float:
		return float64(x)
	}
	return 0
}

func arity(want, got int) error {
	return &fault{code: "E4120",
		fa: "in kar " + faNum(want) + " vorudi mikhad, amma " + faNum(got) + " dade shod",
		en: "this takes " + strconv.Itoa(want) + " arguments but was given " + strconv.Itoa(got)}
}

func faNum(n int) string {
	switch n {
	case 0:
		return "hich"
	case 1:
		return "yek"
	case 2:
		return "do"
	case 3:
		return "se"
	case 4:
		return "chahar"
	}
	return strconv.Itoa(n)
}

func (m *Machine) call(argc int) (Value, bool, *fault, error) {
	if len(m.stack) < argc+1 {
		return Nil(), false, nil, m.bad(0,
			"in seda zadan chizi ke mikhast ro roye poshte peyda nakard",
			"this call did not find what it needed")
	}
	args := m.popN(argc)
	callee := m.pop()

	if callee.tag == TagBuiltin {
		if word, ok := builtins[callee.s]; ok {
			out, err := word(m, toRTs(args))
			if err != nil {
				return Nil(), false, nil, err
			}
			return FromRT(out), false, nil, nil
		}

		v, found, err := m.hostInterp().RunBuiltin(callee.s, toRTs(args), nil)
		if !found {
			return Nil(), false, nil, m.bad(0,
				"in kar ro kasi nanevashte",
				"nothing implements "+callee.s)
		}
		if err != nil {
			return Nil(), false, nil, err
		}
		return FromRT(v), false, nil, nil
	}
	switch c := callee.p.(type) {
	case *funcVal:
		if _, f, err := m.enterFunc(c, args); f != nil || err != nil {
			return Nil(), true, f, err
		}
		return Nil(), true, nil, nil
	case *rt.Class:
		v, entered, f, err := m.construct(c, args)
		if entered || err != nil {
			return v, entered, f, err
		}
		return v, false, f, err
	case *rt.ClassValue:
		v, entered, f, err := m.construct(c.Class, args)
		if entered || err != nil {
			return v, entered, f, err
		}
		return v, false, f, err
	}
	return Nil(), false, nil, m.bad(0,
		"in ro nemishe seda zad",
		"this is not something that can be called")
}

func (m *Machine) callMethod(name string, recv Value, args []Value) (Value, bool, *fault, error) {
	v, entered, fc, err, ok := m.tryMethod(name, recv, args)
	if ok {
		return v, entered, fc, err
	}
	return Nil(), false, nil, m.bad(0,
		"roye in chiz, method-i be esm-e '"+name+"' nist",
		"there is no method called "+name+" on this")
}

func (m *Machine) tryMethod(name string, recv Value, args []Value) (Value, bool, *fault, error, bool) {
	if obj, ok := recv.p.(*rt.Object); ok {
		if fn, f, found := m.methodOf(obj.Class, name); found {
			if f != nil {
				return Nil(), true, f, nil, true
			}

			full := append([]Value{recv}, args...)
			if _, f, err := m.enterFunc(fn, full); f != nil || err != nil {
				return Nil(), true, f, err, true
			}
			return Nil(), true, nil, nil, true
		}
	}
	if cl, ok := recv.p.(*rt.ClassValue); ok {
		if fn, f, found := m.methodOf(cl.Class, name); found {
			if f != nil {
				return Nil(), true, f, nil, true
			}
			full := append([]Value{recv}, args...)
			if _, f, err := m.enterFunc(fn, full); f != nil || err != nil {
				return Nil(), true, f, err, true
			}
			return Nil(), true, nil, nil, true
		}
	}
	fn, ok := methodsFor(recv)
	if ok {
		if word, found := fn[name]; found {

			full := append([]Value{recv}, args...)
			out, err := word(m, toRTs(full))
			if err != nil {
				return Nil(), false, nil, err, true
			}
			return FromRT(out), false, nil, nil, true
		}
	}

	v, found, err := m.hostInterp().RunMethod(recv.ToRT(), name, toRTs(args), nil)
	if err != nil {
		return Nil(), false, nil, err, true
	}
	if found {
		return FromRT(v), false, nil, nil, true
	}
	return Nil(), false, nil, nil, false
}

func methodsFor(v Value) (map[string]builtinFn, bool) {
	switch v.tag {
	case TagStr:
		return textMethods, true
	case TagList:
		return listMethods, true
	case TagDict, TagSet:
		return dictMethods, true
	}
	return nil, false
}

var (
	textMethods = map[string]builtinFn{}
	listMethods = map[string]builtinFn{}
	dictMethods = map[string]builtinFn{}
)

func init() {

	textMethods["tule"] = func(m *Machine, args []rt.Value) (rt.Value, error) {
		if len(args) != 1 {
			return rt.Nil, arity(1, len(args))
		}
		s, ok := args[0].(rt.Str)
		if !ok {
			return rt.Nil, m.bad(0, "in matn nist", "that is not a text")
		}
		return rt.Int(int64(rt.RuneCount(string(s)))), nil
	}
	listMethods["tule"] = func(m *Machine, args []rt.Value) (rt.Value, error) {
		if len(args) != 1 {
			return rt.Nil, arity(1, len(args))
		}
		l, ok := args[0].(*rt.List)
		if !ok {
			return rt.Nil, m.bad(0, "in list nist", "that is not a list")
		}
		return rt.Int(int64(len(l.Items))), nil
	}
}
