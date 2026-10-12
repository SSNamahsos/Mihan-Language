package vm

import (
	"math"
	"strconv"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

type Tag uint8

const (
	TagNil Tag = iota
	TagInt
	TagFloat
	TagStr
	TagBool
	TagList
	TagDict
	TagSet
	TagClass
	TagObject
	TagFunc
	TagBoundMethod
	TagBuiltin
	TagError

	TagIter
	TagBadIter

	TagNone
)

type Value struct {
	tag Tag

	n int64

	s string

	p any
}

func Int(v int64) Value { return Value{tag: TagInt, n: v} }

func Float(v float64) Value { return Value{tag: TagFloat, n: int64(math.Float64bits(v))} }

func Str(v string) Value { return Value{tag: TagStr, s: v} }

func Bool(v bool) Value {
	n := int64(0)
	if v {
		n = 1
	}
	return Value{tag: TagBool, n: n}
}

func Nil() Value { return Value{tag: TagNil} }

func List(l *rt.List) Value { return Value{tag: TagList, p: l} }

func Dict(d *rt.Dict) Value { return Value{tag: TagDict, p: d} }

func Iter(i *iter) Value { return Value{tag: TagIter, p: i} }

func BadIter(f *fault) Value { return Value{tag: TagBadIter, p: f} }
func Set(s *rt.Set) Value    { return Value{tag: TagSet, p: s} }

func Class(c *rt.Class) Value { return Value{tag: TagClass, p: c} }

func ClassValue(c *rt.ClassValue) Value {
	return Value{tag: TagClass, p: c}
}

func Object(o *rt.Object) Value { return Value{tag: TagObject, p: o} }

func Err(e *rt.ErrorValue) Value { return Value{tag: TagError, p: e} }

func Func(f *rt.Function) Value { return Value{tag: TagFunc, p: f} }

func BoundMethod(m *rt.BoundMethod) Value { return Value{tag: TagBoundMethod, p: m} }

func Builtin(name string, f *rt.Function) Value {
	return Value{tag: TagBuiltin, s: name, p: f}
}

func Range(lo, hi int64, inclusive bool) Value {
	return Value{tag: TagInt, n: 0, p: rt.Range{Lo: rt.Int(lo), Hi: rt.Int(hi), Inclusive: inclusive}}
}

func StepRange(lo, step, hi int64) Value {
	return Value{tag: TagInt, n: 0, p: rt.StepRange{Lo: rt.Int(lo), Step: rt.Int(step), Hi: rt.Int(hi)}}
}

func (v Value) TagOf() Tag { return v.tag }

func (v Value) IsNil() bool { return v.tag == TagNil }

func (v Value) IsNumber() bool {
	switch v.tag {
	case TagInt, TagFloat:
		return true
	}
	return false
}

func (v Value) AsInt() (int64, bool) {
	switch v.tag {
	case TagInt, TagBool:
		return v.n, true
	case TagFloat:
		f := v.Float()
		if f == float64(int64(f)) && math.Abs(f) < 1e18 {
			return int64(f), true
		}
	}
	return 0, false
}

func (v Value) Float() float64 { return math.Float64frombits(uint64(v.n)) }

func (v Value) AsFloat() (float64, bool) {
	switch v.tag {
	case TagFloat:
		return v.Float(), true
	case TagInt:
		return float64(v.n), true
	}
	return 0, false
}

func (v Value) AsBool() (bool, bool) {
	switch v.tag {
	case TagBool, TagInt:
		return v.n != 0, true
	case TagFloat:
		return v.Float() != 0, true
	}
	return false, false
}

func (v Value) AsStr() (string, bool) {
	switch v.tag {
	case TagStr:
		return v.s, true
	}
	return "", false
}

func (v Value) Text() string {
	switch v.tag {
	case TagStr:
		return v.s
	case TagInt:
		return strconv.FormatInt(v.n, 10)
	case TagFloat:

		return rt.Show(rt.Float(v.Float()))
	case TagBool:
		if v.n != 0 {
			return "dorost"
		}
		return "ghalat"
	case TagNil:
		return "hich"
	}
	return rt.Show(v.ToRT())
}

func (v Value) Truthy() bool {
	switch v.tag {
	case TagNil:
		return false
	case TagBool, TagInt:
		return v.n != 0
	case TagFloat:
		return v.Float() != 0
	}
	return true
}

func (v Value) TypeName() string {
	switch v.tag {
	case TagNil:
		return "hich"
	case TagInt:
		return "adad"
	case TagFloat:
		return "kesr"
	case TagStr:
		return "matn"
	case TagBool:
		return "bool"
	case TagList:
		return "list"
	case TagDict:
		return "dict"
	case TagSet:
		return "set"
	case TagClass:
		return "darghan"
	case TagObject:
		return "sheishe"
	case TagFunc, TagBoundMethod, TagBuiltin:
		return "funct"
	case TagError:
		return "khata"
	}
	return "?"
}

func (v Value) ToRT() rt.Value {
	switch v.tag {
	case TagNil:
		return rt.Nil
	case TagInt:
		return rt.Int(v.n)
	case TagFloat:
		return rt.Float(v.Float())
	case TagStr:
		return rt.Str(v.s)
	case TagBool:
		return rt.Bool(v.n != 0)

	}
	if fv, ok := v.p.(*funcVal); ok {
		return &rt.Function{
			Name: fv.name,
			Native: func(_ any, args []rt.Value, _ ast.Node) (rt.Value, error) {
				m := fv.home
				if m == nil {
					return nil, &fault{code: "E5002",
						fa: "این تابع دیگر به ماشینی که ساخته بود وصل نیست",
						en: "this function is no longer connected to the machine it was made on"}
				}
				vs := make([]Value, len(args))
				for i, a := range args {
					vs[i] = FromRT(a)
				}
				out, f, err := m.callGo(fv, vs)
				if err != nil {
					return nil, err
				}
				if f != nil {
					return nil, f
				}
				return out.ToRT(), nil
			},
		}
	}
	return v.p
}

func FromRT(v rt.Value) Value {
	switch t := v.(type) {
	case nil:
		return Nil()
	case rt.Hich:
		return Nil()
	case rt.Int:
		return Int(int64(t))
	case rt.Float:
		return Float(float64(t))
	case rt.Str:
		return Str(string(t))
	case rt.Bool:
		return Bool(bool(t))
	case rt.Range:
		return Range(int64(t.Lo), int64(t.Hi), t.Inclusive)
	case rt.StepRange:
		return StepRange(int64(t.Lo), int64(t.Step), int64(t.Hi))
	case *rt.List:
		return List(t)
	case *rt.Dict:
		return Dict(t)
	case *rt.Set:
		return Set(t)
	case *rt.Class:
		return Class(t)
	case *rt.ClassValue:
		return ClassValue(t)
	case *rt.Object:
		return Object(t)
	case *rt.ErrorValue:
		return Err(t)
	case *rt.Function:
		return Func(t)
	case *rt.BoundMethod:
		return BoundMethod(t)
	}
	return Value{tag: TagNil}
}

func toRTList(vs []Value) []rt.Value {
	out := make([]rt.Value, len(vs))
	for i, v := range vs {
		out[i] = v.ToRT()
	}
	return out
}
