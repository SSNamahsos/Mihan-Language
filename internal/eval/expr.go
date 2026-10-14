package eval

import (
	"fmt"
	"math"
	"strings"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
	"mihan/internal/token"
)

func (in *Interp) eval(e ast.Expr, env *rt.Env) (rt.Value, error) {
	switch x := e.(type) {
	case *ast.IntLit:
		return rt.Int(x.Value), nil
	case *ast.FloatLit:
		return rt.Float(x.Value), nil
	case *ast.StrLit:
		return rt.Str(x.Value), nil
	case *ast.CharLit:
		return rt.Str(string(x.Value)), nil
	case *ast.BoolLit:
		return rt.Bool(x.Value), nil
	case *ast.HichLit:
		return rt.Nil, nil

	case *ast.Ident:
		return in.evalIdent(x, env)

	case *ast.InterpLit:
		return in.evalInterp(x, env)

	case *ast.ListComp:
		return in.evalListComp(x, env)

	case *ast.ListLit:
		items := make([]rt.Value, 0, len(x.Elems))
		for _, el := range x.Elems {
			v, err := in.eval(el, env)
			if err != nil {
				return nil, err
			}
			items = append(items, v)
		}
		return rt.NewList(items...), nil

	case *ast.DictLit:
		d := rt.NewDict()
		for _, entry := range x.Entries {
			k, err := in.eval(entry.Key, env)
			if err != nil {
				return nil, err
			}
			v, err := in.eval(entry.Value, env)
			if err != nil {
				return nil, err
			}
			if _, ok := rt.HashKey(k); !ok {
				return nil, in.faultAt(entry.Key, "E4101",
					"kelid-e dict bayad adad, matn, bool ya hich bashe",
					"a dict key has to be a number, matn, bool or hich, not %s", rt.TypeName(k))
			}
			d.Set(k, v)
		}
		return d, nil

	case *ast.SetLit:
		s := rt.NewSet()
		for _, el := range x.Elems {
			v, err := in.eval(el, env)
			if err != nil {
				return nil, err
			}
			if _, ok := rt.HashKey(v); !ok {
				return nil, in.faultAt(el, "E4102",
					"ozv-e set bayad adad, matn, bool ya hich bashe",
					"a set member has to be a number, matn, bool or hich, not %s", rt.TypeName(v))
			}
			s.Add(v)
		}
		return s, nil

	case *ast.Group:
		return in.eval(x.X, env)

	case *ast.Unary:
		return in.evalUnary(x, env)

	case *ast.Binary:
		return in.evalBinary(x, env)

	case *ast.Ternary:
		cond, err := in.eval(x.Cond, env)
		if err != nil {
			return nil, err
		}
		if rt.Truthy(cond) {
			return in.eval(x.Then, env)
		}
		return in.eval(x.Else, env)

	case *ast.Assign:
		if err := in.execAssign(x, env); err != nil {
			return nil, err
		}
		return rt.Nil, nil

	case *ast.RangeExpr:
		lo, err := in.eval(x.Lo, env)
		if err != nil {
			return nil, err
		}
		hi, err := in.eval(x.Hi, env)
		if err != nil {
			return nil, err
		}
		li, ok1 := toInt(lo)
		hiI, ok2 := toInt(hi)
		if !ok1 || !ok2 {
			return nil, in.faultAt(x, "E4103",
				"do sar-e baze bayad adad bashan", "both ends of a range must be whole numbers")
		}
		return &rt.Range{Lo: li, Hi: hiI, Inclusive: x.Inclusive}, nil

	case *ast.Index:
		recv, err := in.eval(x.X, env)
		if err != nil {
			return nil, err
		}
		at, err := in.eval(x.At, env)
		if err != nil {
			return nil, err
		}
		return in.getIndex(x, recv, at)

	case *ast.Slice:
		return in.evalSlice(x, env)

	case *ast.Member:
		return in.evalMember(x, env)

	case *ast.Call:
		return in.evalCall(x, env)

	case *ast.Lambda:
		return in.makeLambda(x, env), nil

	case *ast.ErrorExpr:

		if x.Message != "" {
			return rt.Nil, in.faultAt(x, "E4104", x.Message, "%s", x.English)
		}
		return rt.Nil, in.faultAt(x, "E4104", "in ebarat khoonde nashod", "this expression could not be read")
	}
	return rt.Nil, in.faultAt(e, "E4105",
		"in ebarat ro hanooz hesab nemikonam", "I do not know how to work out this expression yet (%s)", e.NodeKind())
}

func (in *Interp) evalIdent(x *ast.Ident, env *rt.Env) (rt.Value, error) {
	switch x.Name {
	case "khod":

		if b, _ := env.Lookup("__khod__"); b != nil {
			return b.Value, nil
		}
		return nil, in.faultAt(x, "E4106",
			"`khod` biroon az method kar nemikone",
			"khod only means anything inside a method; here there is no current object")
	}
	if b := lookupIdent(x, env); b != nil {
		if !b.Used && b.Value == nil {
			return rt.Nil, in.faultAt(x, "E4107",
				fmt.Sprintf("'%s' hanooz meghdar nagerefte", x.Name),
				"%s was declared but never given a value", x.Name)
		}
		return b.Value, nil
	}

	if c, ok := in.classes[x.Name]; ok {
		return &rt.ClassValue{Class: c}, nil
	}

	if fn, ok := in.builtins[x.Name]; ok {
		return fn, nil
	}

	if typeWords[x.Name] {
		return rt.Str(x.Name), nil
	}
	return nil, in.faultNote(x, "E4108",
		fmt.Sprintf("'%s' ro hich ja nasakhtam", x.Name),
		"%s is not a name I know; it is not a variable, a funct, a darghan or a builtin",
		"har esm bayad ghabl az estefade meghdar begire; esmi ke dakhel-e yek funct sakhte mishe faghat too hamoon funct zende ast, va esm-e halghe-ye har faghat too hamoon halghe",
		x.Name)
}

func (in *Interp) evalInterp(x *ast.InterpLit, env *rt.Env) (rt.Value, error) {
	var b strings.Builder
	for _, seg := range x.Segs {
		if seg.Expr == nil {
			b.WriteString(seg.Text)
			continue
		}
		v, err := in.eval(seg.Expr, env)
		if err != nil {
			return nil, err
		}

		b.WriteString(rt.Show(v))
	}
	return rt.Str(b.String()), nil
}

func (in *Interp) evalUnary(x *ast.Unary, env *rt.Env) (rt.Value, error) {
	v, err := in.eval(x.X, env)
	if err != nil {
		return nil, err
	}
	return in.unaryOp(x, v)
}

func (in *Interp) unaryOp(x *ast.Unary, v rt.Value) (rt.Value, error) {
	switch x.Op {
	case "na", "!":
		return rt.Bool(!rt.Truthy(v)), nil
	case "-":
		switch t := v.(type) {
		case rt.Int:
			return -t, nil
		case rt.Float:
			return -t, nil
		case rt.Bool:
			if t {
				return rt.Int(-1), nil
			}
			return rt.Int(0), nil
		}
	case "+":
		if rt.IsNumber(v) {
			return v, nil
		}
	case "~":
		if i, ok := toInt(v); ok {
			return rt.Int(^i), nil
		}
	}
	return nil, in.faultAt(x, "E4110",
		fmt.Sprintf("in amalgar roye %s kar nemikone", rt.TypeName(v)),
		"the operator %s does not work on %s", x.Op, rt.TypeName(v))
}

func (in *Interp) evalBinary(x *ast.Binary, env *rt.Env) (rt.Value, error) {

	switch x.Op {
	case "va", "&&":
		l, err := in.eval(x.L, env)
		if err != nil {
			return nil, err
		}
		if !rt.Truthy(l) {
			return rt.Bool(false), nil
		}
		r, err := in.eval(x.R, env)
		if err != nil {
			return nil, err
		}
		return rt.Bool(rt.Truthy(r)), nil
	case "ya", "||":
		l, err := in.eval(x.L, env)
		if err != nil {
			return nil, err
		}
		if rt.Truthy(l) {
			return rt.Bool(true), nil
		}
		r, err := in.eval(x.R, env)
		if err != nil {
			return nil, err
		}
		return rt.Bool(rt.Truthy(r)), nil
	case "dar":
		return in.evalIn(x, env)
	}

	l, err := in.eval(x.L, env)
	if err != nil {
		return nil, err
	}
	r, err := in.eval(x.R, env)
	if err != nil {
		return nil, err
	}
	return in.binaryOp(x, x.Op, l, r)
}

func textOp(op token.Kind) bool { return op == "+" }

func (in *Interp) binaryOp(node ast.Node, op token.Kind, l, r rt.Value) (rt.Value, error) {

	if li, ok := l.(rt.Int); ok {
		if ri, ok := r.(rt.Int); ok {
			switch op {
			case "+", "-", "*":
				return in.intArith(node, op, li, ri)
			case "<":
				return rt.Bool(li < ri), nil
			case "<=":
				return rt.Bool(li <= ri), nil
			case ">":
				return rt.Bool(li > ri), nil
			case ">=":
				return rt.Bool(li >= ri), nil
			case "==":
				return rt.Bool(li == ri), nil
			case "!=":
				return rt.Bool(li != ri), nil
			}
		}
	}

	if textOp(op) {
		if ls, ok := l.(rt.Str); ok {
			if rs, ok := r.(rt.Str); ok {
				return ls + rs, nil
			}
			return rt.Str(string(ls) + rt.Show(r)), nil
		}
		if rs, ok := r.(rt.Str); ok {
			return rt.Str(rt.Show(l) + string(rs)), nil
		}
	}

	if op == "*" {
		if s, ok := l.(rt.Str); ok {
			if n, ok := toInt(r); ok {
				return rt.Str(strings.Repeat(string(s), clampRepeat(n))), nil
			}
		}
		if s, ok := r.(rt.Str); ok {
			if n, ok := toInt(l); ok {
				return rt.Str(strings.Repeat(string(s), clampRepeat(n))), nil
			}
		}
	}

	switch op {
	case "+":
		return in.add(node, l, r)
	case "-":
		return in.arith(node, "-", l, r)
	case "*":
		return in.arith(node, "*", l, r)
	case "/":
		return in.divide(node, l, r)
	case "ghesmat":
		return in.floorDiv(node, l, r)
	case "%":
		return in.modulus(node, l, r)
	case "**":
		return in.power(node, l, r)
	case "==":
		return rt.Bool(rt.Equal(l, r)), nil
	case "!=":
		return rt.Bool(!rt.Equal(l, r)), nil
	case "<", "<=", ">", ">=":
		return in.compareOp(node, op, l, r)
	case "<=>":
		c, ok := rt.Compare(l, r)
		if !ok {
			return nil, in.faultAt(node, "E4120",
				fmt.Sprintf("nemishe %s ro ba %s moghayese kard", rt.TypeName(l), rt.TypeName(r)),
				"a %s cannot be measured against a %s", rt.TypeName(l), rt.TypeName(r))
		}
		return rt.Int(c), nil
	case "&", "|", "^", "<<", ">>":
		li, ok1 := toInt(l)
		ri, ok2 := toInt(r)
		if !ok1 || !ok2 {
			return nil, in.faultAt(node, "E4121",
				"in kar faghat ba adad-e kamel mishe", "this operation only works on whole numbers")
		}
		return in.bitOp(node, op, li, ri)
	}
	return nil, in.faultAt(node, "E4122",
		fmt.Sprintf("amalgar-e '%s' ro nemishnasam", op),
		"the operator %s is not one I know", op)
}

func (in *Interp) add(node ast.Node, l, r rt.Value) (rt.Value, error) {
	if ls, ok := l.(rt.Str); ok {
		if rs, ok := r.(rt.Str); ok {
			return ls + rs, nil
		}
		return rt.Str(string(ls) + rt.Show(r)), nil
	}

	if ll, ok := l.(*rt.List); ok {
		if rl, ok := r.(*rt.List); ok {
			out := make([]rt.Value, 0, len(ll.Items)+len(rl.Items))
			out = append(out, ll.Items...)
			out = append(out, rl.Items...)
			return rt.NewList(out...), nil
		}
		return rt.NewList(append(append([]rt.Value{}, ll.Items...), r)...), nil
	}
	if ls, ok := l.(rt.Str); ok {
		return rt.Str(string(ls) + rt.Show(r)), nil
	}
	if _, ok := l.(rt.Str); ok {
		return rt.Str(rt.Show(l)), nil
	}
	if rs, ok := r.(rt.Str); ok {
		return rt.Str(rt.Show(l) + string(rs)), nil
	}
	if li, ok := l.(*rt.List); ok {
		out := make([]rt.Value, 0, len(li.Items)+1)
		out = append(out, li.Items...)
		out = append(out, r)
		return rt.NewList(out...), nil
	}
	return in.arith(node, "+", l, r)
}

func (in *Interp) arith(node ast.Node, op token.Kind, l, r rt.Value) (rt.Value, error) {

	li, lok := l.(rt.Int)
	ri, rok := r.(rt.Int)
	if lok && rok {
		switch op {
		case "+", "-", "*":
			return in.intArith(node, op, li, ri)
		}
	}
	lf, ok1 := toFloat(l)
	rf, ok2 := toFloat(r)
	if !ok1 || !ok2 {
		return nil, in.faultAt(node, "E4130",
			fmt.Sprintf("nemishe %s ro %s kard", rt.TypeName(l), rt.TypeName(r)),
			"a %s and a %s cannot be combined this way", rt.TypeName(l), rt.TypeName(r))
	}
	switch op {
	case "+":
		return rt.Float(lf + rf), nil
	case "-":
		return rt.Float(lf - rf), nil
	case "*":
		return rt.Float(lf * rf), nil
	}
	return nil, in.faultAt(node, "E4131",
		fmt.Sprintf("amalgar-e '%s' roye in do meghdar kar nemikone", op),
		"the operator %s does not work on a %s and a %s", op, rt.TypeName(l), rt.TypeName(r))
}

func (in *Interp) intArith(node ast.Node, op token.Kind, a, b rt.Int) (rt.Value, error) {
	var c rt.Int
	bad := false
	switch op {
	case "+":
		c = a + b
		bad = (a >= 0) == (b >= 0) && (c >= 0) != (a >= 0)
	case "-":
		c = a - b
		bad = (a >= 0) != (b >= 0) && (c >= 0) != (a >= 0)
	case "*":
		c = a * b
		bad = a != 0 && (c/a != b || (a == -1 && b == math.MinInt64) || (b == -1 && a == math.MinInt64))
	}
	if bad {
		return nil, in.overflow(node)
	}
	return c, nil
}

func (in *Interp) overflow(node ast.Node) error {
	return in.faultAt(node, "E4126",
		"in adad az andaze-ye adad-e kamel (64 biti) bozorgtar shod",
		"the whole number got too big for 64 bits; use a float (e.g. 1.0 * x) if you need bigger numbers")
}

func (in *Interp) divide(node ast.Node, l, r rt.Value) (rt.Value, error) {
	rf, ok := toFloat(r)
	if !ok {
		return nil, in.faultAt(node, "E4132",
			"nemishe bar %s taghsim kard", "you cannot divide by %s", rt.TypeName(r))
	}
	if rf == 0 {
		return nil, in.faultAt(node, "E4133",
			"nemishe bar sefr taghsim kard", "division by zero has no answer")
	}

	lf, ok := toFloat(l)
	if !ok {
		return nil, in.faultAt(node, "E4132",
			"nemishe %s ro bar %s taghsim kard", "a %s cannot be divided by a %s", rt.TypeName(l), rt.TypeName(r))
	}
	return rt.Float(lf / rf), nil
}

func (in *Interp) floorDiv(node ast.Node, l, r rt.Value) (rt.Value, error) {
	li, ok1 := toInt(l)
	ri, ok2 := toInt(r)
	if !ok1 || !ok2 {

		lf, fok1 := toFloat(l)
		rf, fok2 := toFloat(r)
		if !fok1 || !fok2 {
			return nil, in.faultAt(node, "E4134",
				"nemishe %s ro %s kard", "a %s and a %s cannot be divided like this",
				rt.TypeName(l), rt.TypeName(r))
		}
		if rf == 0 {
			return nil, in.faultAt(node, "E4133",
				"nemishe bar sefr taghsim kard", "division by zero has no answer")
		}
		return rt.Int(math.Floor(lf / rf)), nil
	}
	if ri == 0 {
		return nil, in.faultAt(node, "E4133",
			"nemishe bar sefr taghsim kard", "division by zero has no answer")
	}
	q := li / ri

	if (li%ri != 0) && ((li < 0) != (ri < 0)) {
		q--
	}
	return q, nil
}

func (in *Interp) modulus(node ast.Node, l, r rt.Value) (rt.Value, error) {
	li, ok1 := l.(rt.Int)
	ri, ok2 := r.(rt.Int)
	if ok1 && ok2 {
		if ri == 0 {
			return nil, in.faultAt(node, "E4135",
				"nemishe bar sefr baghimande gereft", "the remainder of dividing by zero has no answer")
		}
		m := li % ri
		if m != 0 && ((m < 0) != (ri < 0)) {
			m += ri
		}
		return m, nil
	}
	lf, fok1 := toFloat(l)
	rf, fok2 := toFloat(r)
	if !fok1 || !fok2 {
		return nil, in.faultAt(node, "E4136",
			"baghimande faghat baraye adad ast", "the remainder is only defined for numbers")
	}
	if rf == 0 {
		return nil, in.faultAt(node, "E4135",
			"nemishe bar sefr baghimande gereft", "the remainder of dividing by zero has no answer")
	}
	return rt.Float(math.Mod(lf, rf)), nil
}

func (in *Interp) power(node ast.Node, l, r rt.Value) (rt.Value, error) {
	li, lok := l.(rt.Int)
	ri, rok := r.(rt.Int)
	if lok && rok {
		if ri >= 0 {
			result := rt.Int(1)
			for i := rt.Int(0); i < ri; i++ {
				v, err := in.intArith(node, "*", result, li)
				if err != nil {
					return nil, err
				}
				result = v.(rt.Int)
				if result == 0 || result == 1 && li == 1 {
					break
				}
			}
			return result, nil
		}
		if li == 1 {
			return rt.Int(1), nil
		}
		if li == -1 {
			if ri%2 == 0 {
				return rt.Int(1), nil
			}
			return rt.Int(-1), nil
		}
		if li == 0 {
			return nil, in.faultAt(node, "E4137",
				"sefr be tavan-e manfi javab nadare", "zero to a negative power has no answer")
		}
		return rt.Float(1 / math.Pow(float64(li), float64(ri))), nil
	}
	lf, ok1 := toFloat(l)
	rf, ok2 := toFloat(r)
	if !ok1 || !ok2 {
		return nil, in.faultAt(node, "E4138",
			"tavan faghat baraye adad ast", "a power is only defined for numbers")
	}
	return rt.Float(math.Pow(lf, rf)), nil
}

func (in *Interp) bitOp(node ast.Node, op token.Kind, l, r rt.Int) (rt.Value, error) {
	switch op {
	case "&":
		return l & r, nil
	case "|":
		return l | r, nil
	case "^":
		return l ^ r, nil
	case "<<":
		if r < 0 {
			return rt.Int(int64(l) >> uint(-r)), nil
		}
		if r > 62 {
			return nil, fmt.Errorf("bit shift by %d is too far", r)
		}
		return l << uint(r), nil
	case ">>":
		if r < 0 {
			return l << uint(-r), nil
		}
		return l >> uint(r), nil
	}
	return nil, fmt.Errorf("unknown bit operator %s", op)
}

func clampRepeat(n rt.Int) int {
	const limit = 1 << 24
	if n < 0 {
		return 0
	}
	if n > limit {
		return limit
	}
	return int(n)
}

func (in *Interp) compareOp(node ast.Node, op token.Kind, l, r rt.Value) (rt.Value, error) {
	c, ok := rt.Compare(l, r)
	if !ok {
		return nil, in.faultAt(node, "E4139",
			fmt.Sprintf("nemishe %s ro ba %s moghayese kard", rt.TypeName(l), rt.TypeName(r)),
			"a %s cannot be compared with a %s", rt.TypeName(l), rt.TypeName(r))
	}
	switch op {
	case "<":
		return rt.Bool(c < 0), nil
	case "<=":
		return rt.Bool(c <= 0), nil
	case ">":
		return rt.Bool(c > 0), nil
	case ">=":
		return rt.Bool(c >= 0), nil
	}
	return rt.Bool(false), nil
}

func (in *Interp) evalIn(x *ast.Binary, env *rt.Env) (rt.Value, error) {
	l, err := in.eval(x.L, env)
	if err != nil {
		return nil, err
	}
	r, err := in.eval(x.R, env)
	if err != nil {
		return nil, err
	}
	return in.inOp(x, l, r)
}

func (in *Interp) inOp(x *ast.Binary, l, r rt.Value) (rt.Value, error) {
	switch t := r.(type) {
	case *rt.List:
		for _, item := range t.Items {
			if rt.Equal(l, item) {
				return rt.Bool(true), nil
			}
		}
		return rt.Bool(false), nil
	case *rt.Set:
		return rt.Bool(t.Has(l)), nil
	case *rt.Dict:
		_, ok := t.Get(l)
		return rt.Bool(ok), nil
	case rt.Str:
		return rt.Bool(strings.Contains(string(t), rt.Show(l))), nil
	case *rt.Range:
		for _, n := range rangeItems(*t) {
			if rt.Equal(l, n) {
				return rt.Bool(true), nil
			}
		}
		return rt.Bool(false), nil
	}
	return nil, in.faultAt(x, "E4140",
		fmt.Sprintf("nemishe too %s gasht", rt.TypeName(r)),
		"`dar` needs a list, set, dict, matn or range, not a %s", rt.TypeName(r))
}

var typeWords = map[string]bool{
	"arz": true, "int": true, "kasr": true, "float": true, "double": true,
	"matn": true, "string": true, "str": true, "bool": true, "boolean": true,
	"harf": true, "char": true, "rune": true, "list": true, "dict": true,
	"map": true, "set": true,
}

func (in *Interp) evalListComp(x *ast.ListComp, env *rt.Env) (rt.Value, error) {
	coll, err := in.eval(x.Iter, env)
	if err != nil {
		return nil, err
	}
	items, err := in.iterate(x.Iter, coll)
	if err != nil {
		return nil, err
	}
	out := make([]rt.Value, 0, len(items))
	for _, item := range items {
		scope := rt.NewBlockEnv(env)
		scope.Declare(x.VarName, item)
		if x.Filter != nil {
			keep, err := in.eval(x.Filter, scope)
			if err != nil {
				return nil, err
			}
			if !rt.Truthy(keep) {
				continue
			}
		}
		v, err := in.eval(x.Elem, scope)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return rt.NewList(out...), nil
}
