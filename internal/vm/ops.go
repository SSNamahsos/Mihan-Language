package vm

import (
	"strconv"
	"strings"

	"mihan/internal/ir"
	rt "mihan/internal/runtime"
)

func makeIter(hi, lo Value, inclusive bool) Value {
	a, aok := lo.AsInt()
	b, bok := hi.AsInt()
	if !aok || !bok {
		return BadIter(wrongType(hi, lo, numbers))
	}
	step := int64(1)
	if a > b {
		step = -1
	}
	last := b
	if inclusive {
		last = b + step
	}
	return Iter(&iter{cur: a, step: step, last: last})
}

func (m *Machine) makeStepIter(v Value) Value {

	if v.tag == TagStr {
		return Iter(&iter{str: v.s})
	}
	switch x := v.p.(type) {
	case *rt.List:
		return Iter(&iter{list: x})
	case *rt.Dict:
		l := rt.NewList()
		for _, k := range x.Keys() {
			l.Items = append(l.Items, k)
		}
		return Iter(&iter{list: l})
	}
	return BadIter(wrongType(v, Nil(), kind{verb: "walk", what: "these", wantT: "list, matn ya dict"}))
}

func next(it *iter) (Value, bool) {
	switch {
	case it.list != nil:
		if it.at >= len(it.list.Items) {
			return Nil(), false
		}
		out := FromRT(it.list.Items[it.at])
		it.at++
		return out, true
	case it.str != "":
		r := []rune(it.str)
		if it.sat >= len(r) {
			return Nil(), false
		}
		out := Str(string(r[it.sat]))
		it.sat++
		return out, true
	case it.done, (it.step > 0 && it.cur >= it.last) || (it.step < 0 && it.cur <= it.last):
		return Nil(), false
	}
	out := Int(it.cur)
	it.cur += it.step
	return out, true
}

func (m *Machine) jumpIfFalse(f *frame, target int32) (bool, error) {
	if n := len(m.stack); n > 0 {
		if it, ok := m.stack[n-1].p.(*iter); ok {
			out, more := next(it)
			if !more {

				m.stack = m.stack[:n-1]
				m.pc = target
				return false, nil
			}
			m.push(out)
			return false, nil
		}
	}
	if !m.pop().Truthy() {
		m.pc = target
	}
	return false, nil
}

func compareOp(op ir.Op, a, b Value) (Value, *fault) {
	if op == ir.OpEq {
		return Bool(equal(a, b)), nil
	}
	if op == ir.OpNe {
		return Bool(!equal(a, b)), nil
	}
	c, ok := compare(a, b)
	if !ok {
		return Nil(), wrongType(b, a, kind{verb: "put in order", what: "these", wantT: "values that can be ordered"})
	}
	switch op {
	case ir.OpLt:
		return Bool(c < 0), nil
	case ir.OpLe:
		return Bool(c <= 0), nil
	case ir.OpGt:
		return Bool(c > 0), nil
	case ir.OpGe:
		return Bool(c >= 0), nil
	}
	return Nil(), &fault{code: "E4309",
		fa: "این مقایسه را نمی‌دانم",
		en: "I do not know this comparison"}
}

func unaryMinus(v Value) (Value, *fault) {
	switch v.tag {
	case TagInt:
		return Int(-v.n), nil
	case TagFloat:
		return Float(-v.Float()), nil
	case TagBool:
		return Int(-v.n), nil
	}
	return Nil(), wrongType(v, Nil(), numbers)
}

func contains(coll, item Value) (Value, *fault) {
	if l, ok := coll.p.(*rt.List); ok {
		for _, it := range l.Items {
			if equal(item, FromRT(it)) {
				return Bool(true), nil
			}
		}
		return Bool(false), nil
	}
	if s, ok := coll.AsStr(); ok {
		return Bool(containsText(s, item.Text())), nil
	}
	if d, ok := coll.p.(*rt.Dict); ok {
		_, found := d.Get(item.ToRT())
		return Bool(found), nil
	}
	if st, ok := coll.p.(*rt.Set); ok {
		return Bool(st.Has(item.ToRT())), nil
	}
	if it, ok := coll.p.(*iter); ok {

		for _, n := range iterValues(*it) {
			if equal(item, FromRT(n)) {
				return Bool(true), nil
			}
		}
		return Bool(false), nil
	}
	return Nil(), &fault{code: "E4140",
		fa: "نمی‌شود در " + coll.TypeName() + " گشت",
		en: "`dar` needs a list, set, dict, matn or range, not a " + coll.TypeName()}
}

func containsText(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func appendTo(coll, val Value) *fault {
	switch x := coll.p.(type) {
	case *rt.List:
		x.Items = append(x.Items, val.ToRT())
		return nil
	case *rt.Set:
		x.Add(val.ToRT())
		return nil
	case *rt.Dict:
		if d, ok := val.p.(*rt.Dict); ok {
			for _, k := range d.Keys() {
				v, _ := d.Get(k)
				x.Set(k, v)
			}
			return nil
		}
	}
	return wrongType(val, coll, kind{verb: "add to", what: "this", wantT: "list, set ya dict"})
}

func index(coll, key Value) (Value, *fault) {
	switch x := coll.p.(type) {
	case *rt.List:
		i, ok := key.AsInt()
		if !ok {
			return Nil(), wrongType(key, coll, numbers)
		}
		items := x.Items
		n := int64(len(items))
		if i < 0 {
			i += n
		}
		if i < 0 || i >= n {
			return Nil(), outOfRange(i, n)
		}
		return FromRT(items[i]), nil
	case *rt.Dict:
		v, ok := x.Get(key.ToRT())
		if !ok {
			return Nil(), missingKey(key, coll)
		}
		return FromRT(v), nil
	case *rt.Set:
		if x.Has(key.ToRT()) {
			return coll, nil
		}
		return Nil(), missingKey(key, coll)
	}
	if s, ok := coll.AsStr(); ok {
		i, ok := key.AsInt()
		if !ok {
			return Nil(), wrongType(key, coll, numbers)
		}
		r := []rune(s)
		n := int64(len(r))
		if i < 0 {
			i += n
		}
		if i < 0 || i >= n {
			return Nil(), outOfRange(i, n)
		}
		return Str(string(r[i])), nil
	}
	return Nil(), wrongType(key, coll, kind{verb: "read from", what: "this", wantT: "list, matn ya dict"})
}

func setIndex(coll, key, val Value) *fault {
	switch x := coll.p.(type) {
	case *rt.List:
		i, ok := key.AsInt()
		if !ok {
			return wrongType(key, coll, numbers)
		}
		items := x.Items
		n := int64(len(items))
		if i < 0 {
			i += n
		}
		if i < 0 || i >= n {
			return outOfRange(i, n)
		}
		items[i] = val.ToRT()
		return nil
	case *rt.Dict:
		x.Set(key.ToRT(), val.ToRT())
		return nil
	}
	return wrongType(key, coll, kind{verb: "write to", what: "this", wantT: "list ya dict"})
}

func (m *Machine) makeDict(n int) Value {

	items := m.popN(n * 2)
	d := rt.NewDict()

	for i := 0; i+1 < len(items); i += 2 {
		d.Set(items[i].ToRT(), items[i+1].ToRT())
	}
	return Dict(d)
}

func (m *Machine) interpolate(n int) Value {
	parts := m.popN(n)
	if len(parts) == 1 {
		return Str(parts[0].Text())
	}
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p.Text())
	}
	return Str(b.String())
}

func sliceOf(coll, lo, hi Value, inclusive bool) (Value, *fault) {

	if coll.tag == TagStr {
		loI, hiI := 0, rt.RuneCount(coll.s)
		if !lo.IsNil() {
			if n, ok := lo.AsInt(); ok {
				loI = int(n)
			}
		}
		if !hi.IsNil() {
			if n, ok := hi.AsInt(); ok {
				hiI = int(n)
			}
		}
		return Str(rt.Slice(coll.s, loI, hiI, inclusive)), nil
	}
	switch x := coll.p.(type) {
	case *rt.List:
		s, e := sliceBounds(len(x.Items), lo, hi, inclusive)
		out := make([]rt.Value, e-s)
		copy(out, x.Items[s:e])
		return List(rt.NewList(out...)), nil
	case *rt.Set:
		items := x.Items()
		s, e := sliceBounds(len(items), lo, hi, inclusive)
		out := make([]rt.Value, e-s)
		copy(out, items[s:e])
		return List(rt.NewList(out...)), nil
	case *iter:

		items := iterValues(*x)
		s, e := sliceBounds(len(items), lo, hi, inclusive)
		out := make([]rt.Value, e-s)
		copy(out, items[s:e])
		return List(rt.NewList(out...)), nil
	case *rt.Range:

		items := rangeValues(*x)
		s, e := sliceBounds(len(items), lo, hi, inclusive)
		out := make([]rt.Value, e-s)
		copy(out, items[s:e])
		return List(rt.NewList(out...)), nil
	}
	return Nil(), &fault{code: "E4186",
		fa: "از " + coll.TypeName() + " نمی‌شود برش گرفت",
		en: "you cannot take a slice of a " + coll.TypeName()}
}

func iterValues(it iter) []rt.Value {
	var out []rt.Value
	c := it
	for {
		v, more := next(&c)
		if !more {
			return out
		}
		out = append(out, v.ToRT())
	}
}

func sliceBounds(n int, lo, hi Value, inclusive bool) (int, int) {
	start, end := 0, n
	if !lo.IsNil() {
		if i, ok := lo.AsInt(); ok {
			start = int(i)
		}
	}
	if !hi.IsNil() {
		if i, ok := hi.AsInt(); ok {
			end = int(i)
			if inclusive {
				end++
			}
		}
	}
	if start < 0 {
		start += n
	}
	if end < 0 {
		end += n
	}
	if start < 0 {
		start = 0
	}

	if start > n {
		start = n
	}
	if end > n {
		end = n
	}
	if end < start {
		end = start
	}
	return start, end
}

func rangeValues(r rt.Range) []rt.Value {
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

func stepRangeValues(r rt.StepRange) []rt.Value {
	var out []rt.Value
	if r.Step == 0 {
		return out
	}
	for i := r.Lo; ; i += r.Step {
		if r.Step > 0 && i > r.Hi {
			break
		}
		if r.Step < 0 && i < r.Hi {
			break
		}
		out = append(out, i)
		if len(out) > 10_000_000 {
			break
		}
	}
	return out
}

func length(v Value) Value {
	if v.tag == TagStr {
		return Int(int64(len([]rune(v.s))))
	}
	switch x := v.p.(type) {
	case *rt.List:
		return Int(int64(len(x.Items)))
	case *rt.Dict:
		return Int(int64(x.Len()))
	case *rt.Set:
		return Int(int64(x.Len()))
	}
	return Int(0)
}

func outOfRange(i, n int64) *fault {
	return &fault{code: "E4231",
		fa: "این جایگاه (" + strconv.FormatInt(i, 10) + ") بیرون از فهرستی است که " +
			strconv.FormatInt(n, 10) + " عضو دارد",
		en: "that index is outside a collection of " + strconv.FormatInt(n, 10) + " items"}
}

func missingKey(key, coll Value) *fault {
	return &fault{code: "E4230",
		fa: "کلیدی به نام «" + key.Text() + "» در این دیکشنری نیست",
		en: "there is no key " + key.Text() + " in this dict"}
}
