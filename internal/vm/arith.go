package vm

import (
	"math"

	"mihan/internal/ir"
	rt "mihan/internal/runtime"
)

type fault struct {
	code string

	fa, en string

	line int32

	threw *Value
}

func (e *fault) Error() string { return e.en }

type kind struct {
	verb  string
	what  string
	wantT string
}

var (
	numbers = kind{verb: "do", what: "numbers", wantT: "adad or kesr"}
	texts   = kind{verb: "do", what: "text", wantT: "matn"}
)

func wrongType(a Value, b Value, k kind) *fault {
	return &fault{
		code: "E4300",
		fa:   "این کار را نمی‌شود کرد، چون یکی از این دو " + k.what + " نیست",
		en:   "you cannot " + k.verb + " " + k.what + ": one of these is a " + a.TypeName() + " and the other is a " + b.TypeName(),
	}
}

func arith(op ir.Op, a, b Value) (Value, *fault) {

	if op == ir.OpAdd {
		switch {
		case a.tag == TagStr && b.tag == TagStr:
			return Str(a.s + b.s), nil
		case a.tag == TagStr:
			return Str(a.s + b.Text()), nil
		case b.tag == TagStr:
			return Str(a.Text() + b.s), nil
		case a.tag == TagList && b.tag == TagList:
			return listConcat(a, b)
		case a.tag == TagDict && b.tag == TagDict:
			return dictMerge(a, b)
		}
	}
	af, aok := a.AsFloat()
	bf, bok := b.AsFloat()
	if !aok || !bok {
		return Nil(), wrongType(a, b, numbers)
	}

	_, aint := a.IsWhole()
	_, bint := b.IsWhole()
	whole := aint && bint

	switch op {
	case ir.OpAdd:

		if a.tag == TagStr {
			return Str(a.s + b.Text()), nil
		}
		if b.tag == TagStr {
			return Str(a.Text() + b.s), nil
		}
		if a.tag == TagList && b.tag == TagList {
			return listConcat(a, b)
		}
		if a.tag == TagDict && b.tag == TagDict {
			return dictMerge(a, b)
		}
		if whole {
			return Int(int64(af) + int64(bf)), nil
		}
		return Float(af + bf), nil
	case ir.OpSub:
		if whole {
			return Int(int64(af) - int64(bf)), nil
		}
		return Float(af - bf), nil
	case ir.OpMul:
		if whole {
			return Int(int64(af) * int64(bf)), nil
		}
		return Float(af * bf), nil
	case ir.OpDiv:

		if bf == 0 {
			return Nil(), &fault{
				code: "E4301",
				fa:   "نمی‌شود بر صفر تقسیم کرد",
				en:   "you cannot divide by zero",
			}
		}
		return Float(af / bf), nil
	case ir.OpIDiv:

		if bf == 0 {
			return Nil(), &fault{
				code: "E4302",
				fa:   "نمی‌شود بر صفر تقسیم کرد",
				en:   "you cannot divide by zero",
			}
		}
		return Int(int64(math.Floor(af / bf))), nil
	case ir.OpRem:
		if whole {
			if int64(bf) == 0 {
				return Nil(), &fault{
					code: "E4303",
					fa:   "نمی‌شود بر صفر باقی‌مانده گرفت",
					en:   "you cannot take the remainder of a division by zero",
				}
			}
			return Int(flooredMod(int64(af), int64(bf))), nil
		}
		return Float(math.Mod(af, bf)), nil
	case ir.OpPow:

		r := math.Pow(af, bf)
		if whole && r == math.Trunc(r) && math.Abs(r) < 1e18 {
			return Int(int64(r)), nil
		}
		return Float(r), nil
	}
	return Nil(), &fault{code: "E4309", fa: "این حساب را نمی‌کنم", en: "I do not know this arithmetic"}
}

func flooredMod(a, b int64) int64 {
	r := a % b
	if r != 0 && (r < 0) != (b < 0) {
		r += b
	}
	return r
}

func (v Value) IsWhole() (int64, bool) {
	switch v.tag {
	case TagInt, TagBool:
		return v.n, true
	case TagFloat:
		f := v.Float()
		if f == math.Trunc(f) && math.Abs(f) < 1e18 {
			return int64(f), true
		}
	}
	return 0, false
}

func listConcat(a, b Value) (Value, *fault) {
	l, r := a.p.(*rt.List), b.p.(*rt.List)
	out := make([]rt.Value, 0, len(l.Items)+len(r.Items))
	out = append(out, l.Items...)
	out = append(out, r.Items...)
	return List(rt.NewList(out...)), nil
}

func dictMerge(a, b Value) (Value, *fault) {
	l, r := a.p.(*rt.Dict), b.p.(*rt.Dict)
	out := rt.NewDict()
	for _, k := range l.Keys() {
		if v, ok := l.Get(k); ok {
			out.Set(k, v)
		}
	}
	for _, k := range r.Keys() {
		if v, ok := r.Get(k); ok {
			out.Set(k, v)
		}
	}
	return Dict(out), nil
}

func compare(a, b Value) (int, bool) {
	if a.tag == TagStr && b.tag == TagStr {
		switch {
		case a.s < b.s:
			return -1, true
		case a.s > b.s:
			return 1, true
		}
		return 0, true
	}
	af, aok := a.AsFloat()
	bf, bok := b.AsFloat()
	if !aok || !bok {
		return 0, false
	}
	switch {
	case af < bf:
		return -1, true
	case af > bf:
		return 1, true
	}
	return 0, true
}

func equal(a, b Value) bool {
	if a.tag == TagStr && b.tag == TagStr {
		return a.s == b.s
	}

	af, aok := a.AsFloat()
	bf, bok := b.AsFloat()
	if aok && bok {
		return af == bf
	}
	if a.tag != b.tag {
		return false
	}
	switch a.tag {
	case TagNil:
		return true
	case TagBool:
		return (a.n != 0) == (b.n != 0)
	case TagList:
		al, bl := a.p.(*rt.List), b.p.(*rt.List)
		if len(al.Items) != len(bl.Items) {
			return false
		}
		for i := range al.Items {
			if !equal(FromRT(al.Items[i]), FromRT(bl.Items[i])) {
				return false
			}
		}
		return true
	case TagDict:
		ad, bd := a.p.(*rt.Dict), b.p.(*rt.Dict)
		if ad.Len() != bd.Len() {
			return false
		}
		for _, k := range ad.Keys() {
			want, ok := bd.Get(k)
			if !ok {
				return false
			}
			got, _ := ad.Get(k)
			if !equal(FromRT(got), FromRT(want)) {
				return false
			}
		}
		return true
	}
	if a.p == nil && b.p == nil {
		return a.n == b.n && a.s == b.s
	}
	return a.p == b.p
}
