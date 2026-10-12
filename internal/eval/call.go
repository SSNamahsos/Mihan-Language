package eval

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"mihan/internal/ast"
	"mihan/internal/diag"
	rt "mihan/internal/runtime"
)

const maxCallDepth = 800

func (in *Interp) makeLambda(x *ast.Lambda, env *rt.Env) rt.Value {
	fn := &rt.Function{Params: x.Params, Body: x.Body, Expr: x.Expr, Env: env}
	if in.fastEnabled() {
		if cf := compiledLambdaFor(x); cf != nil {
			fn.Fast = cf
		}
	}
	return fn
}

func (in *Interp) methodFunc(m *ast.FuncDecl, owner *rt.Class) *rt.Function {
	fn := &rt.Function{Name: m.Name, Params: m.Params, Body: m.Body, Env: in.Globals, Class: owner}
	if in.fastEnabled() {
		if cf := compiledMethodFor(m); cf != nil {
			fn.Fast = cf
		}
	}
	return fn
}

func (in *Interp) evalCall(x *ast.Call, env *rt.Env) (rt.Value, error) {
	if in.Hooks != nil && in.Hooks.OnCall != nil {
		in.Hooks.OnCall(x)
	}
	if in.Depth > maxCallDepth {
		return nil, in.faultAt(x, "E4150",
			"تابع‌ها بیش از حد همدیگر را صدا زدند و من ایستادم",
			"the calls went on deeper than %d and I stopped, so this is probably a function that never ends", maxCallDepth)
	}

	if id, ok := x.Fn.(*ast.Ident); ok {
		args, err := in.evalArgs(x.Args, env)
		if err != nil {
			return nil, err
		}
		if b := lookupIdent(id, env); b != nil {
			return in.callValue(x, b.Value, nil, args)
		}
		if c, ok := in.classes[id.Name]; ok {
			return in.constructNode(x, c, args)
		}
		if fn, ok := in.builtins[id.Name]; ok {
			return in.callNative(x, fn, args)
		}
		return nil, in.faultAt(x, "E4151",
			fmt.Sprintf("«%s» را هیچ‌جا نساخته‌ام", id.Name),
			"%s is not a name I know; it is not a variable, a funct, a darghan or a builtin", id.Name)
	}

	recv, err := in.eval(x.Fn, env)
	if err != nil {
		return nil, err
	}
	args, err := in.evalArgs(x.Args, env)
	if err != nil {
		return nil, err
	}
	return in.callValue(x, recv, nil, args)
}

func (in *Interp) callWith(node ast.Node, fn rt.Value, args []rt.Value) (rt.Value, error) {
	return in.callValue(node, fn, nil, args)
}

func (in *Interp) callValue(node ast.Node, callee rt.Value, recv *rt.Object, args []rt.Value) (rt.Value, error) {
	switch t := callee.(type) {
	case *rt.Function:
		return in.callFunction(node, t, recv, args)
	case *rt.BoundMethod:
		return in.callFunction(node, t.Fn, t.Recv.(*rt.Object), args)
	case *rt.ClassValue:
		return in.constructNode(node, t.Class, args)
	case *rt.ErrorValue:

		return nil, in.faultAt(node, "E4152",
			"روی یک khata تابع صدا زدی", "you tried to call a khata as if it were a funct")
	}
	return nil, in.faultAt(node, "E4153",
		fmt.Sprintf("%s را نمی‌شود صدا زد", rt.TypeName(callee)),
		"a %s cannot be called; you need a funct, a darghan or a builtin name", rt.TypeName(callee))
}

func (in *Interp) callNative(node ast.Node, fn *rt.Function, args []rt.Value) (rt.Value, error) {

	plain := make([]rt.Value, len(args))
	for i, a := range args {
		if na, ok := a.(rt.NamedArg); ok {
			plain[i] = na.Value()
			continue
		}
		plain[i] = a
	}
	v, err := fn.Native(in, plain, node)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (in *Interp) callFunction(node ast.Node, fn *rt.Function, recv *rt.Object, args []rt.Value) (rt.Value, error) {
	if fn.Native != nil {
		return in.callNative(node, fn, args)
	}
	if in.Depth > maxCallDepth {
		return nil, in.faultAt(node, "E4150",
			"تابع‌ها بیش از حد همدیگر را صدا زدند و من ایستادم",
			"the calls went on deeper than %d and I stopped, so this is probably a function that never ends", maxCallDepth)
	}
	if fn.Fast != nil {
		if v, done, err := in.tryFastCall(node, fn, recv, args); done {
			return v, err
		}
	}

	scope := rt.NewEnv(fn.Env)
	if recv != nil {
		scope.Declare("__khod__", recv)
	}

	if err := in.bindParams(node, fn, scope, args); err != nil {
		return nil, err
	}

	if fn.Expr != nil {
		return in.eval(fn.Expr, scope)
	}

	in.Depth++
	in.Trace = append(in.Trace, rt.Frame{
		Name: fn.Name, File: in.File,
		Pos: posOf(node), Callee: fn.Name,
	})
	flow, err := in.execBlockIn(fn.Body, scope)
	in.Trace = in.Trace[:len(in.Trace)-1]
	in.Depth--
	if err != nil {
		return nil, err
	}
	if in.Hooks != nil && in.Hooks.OnReturn != nil {
		in.Hooks.OnReturn(node)
	}
	if flow.Kind == rt.CtrlReturn {
		return flow.Val, nil
	}

	return rt.Nil, nil
}

func posOf(node ast.Node) diag.Pos {
	if node == nil {
		return diag.Pos{}
	}
	return node.Span().Start
}

func (in *Interp) bindParams(node ast.Node, fn *rt.Function, scope *rt.Env, args []rt.Value) error {

	plain := len(args) <= len(fn.Params)
	for _, a := range args {
		if _, ok := a.(rt.NamedArg); ok {
			plain = false
			break
		}
	}
	if plain {
		for i, p := range fn.Params {
			if i < len(args) {
				scope.Declare(p.Name, args[i])
				continue
			}
			if p.Default != nil {
				v, err := in.eval(p.Default, fn.Env)
				if err != nil {
					return err
				}
				scope.Declare(p.Name, v)
				continue
			}
			return faultf(node, "E4155",
				"به تابع «%s» مقدار %s داده نشد", "%s was not given a value for %s", fn.Name, p.Name)
		}
		return nil
	}
	named := map[string]rt.Value{}
	var positional []rt.Value
	for _, a := range args {
		na, ok := a.(rt.NamedArg)
		if ok && na.IsNamed() {
			named[na.Name()] = na.Value()
			continue
		}
		if ok {
			positional = append(positional, na.Value())
			continue
		}
		positional = append(positional, a)
	}

	if fn.Params == nil && len(args) == 0 {
		return nil
	}

	for name := range named {
		found := false
		for _, p := range fn.Params {
			if p.Name == name {
				found = true
				break
			}
		}
		if !found {
			return faultf(node, "E4154",
				"تابع «%s» پارامتری به نام «%s» ندارد", "the funct %s has no parameter called %s", fn.Name, name)
		}
	}

	pi := 0
	for _, p := range fn.Params {
		if v, ok := named[p.Name]; ok {
			scope.Declare(p.Name, v)
			continue
		}
		if pi < len(positional) {
			scope.Declare(p.Name, positional[pi])
			pi++
			continue
		}
		if p.Default != nil {

			v, err := in.eval(p.Default, fn.Env)
			if err != nil {
				return err
			}
			scope.Declare(p.Name, v)
			continue
		}
		return faultf(node, "E4155",
			"به تابع «%s» مقدار %s داده نشد", "%s was not given a value for %s", fn.Name, p.Name)
	}
	if pi < len(positional) {
		return faultf(node, "E4156",
			"به تابع «%s» %d مقدار داده شد ولی %d پارامتر دارد",
			"%s was given %d values but takes %d", fn.Name, len(positional), len(fn.Params))
	}
	return nil
}

func (in *Interp) evalDefault(e ast.Expr, env *rt.Env, node ast.Node) (rt.Value, error) {
	v, err := in.eval(e, env)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (in *Interp) constructNode(node ast.Node, c *rt.Class, args []rt.Value) (rt.Value, error) {
	if c.Enumerated {

		return rt.NewObject(c), nil
	}
	obj := rt.NewObject(c)

	if ctor, _ := c.FindMethod("__sakht__"); ctor != nil {
		fn := in.methodFunc(ctor, c)
		fn.Name = "__sakht__"
		if _, err := in.callFunction(node, fn, obj, args); err != nil {
			return nil, err
		}
		return obj, nil
	}
	if len(args) > 0 {
		return nil, faultf(node, "E4157",
			fmt.Sprintf("«%s» ساخت‌دهنده ندارد، پس %d مقدار اضافه آمد", c.Name, len(args)),
			"%s has no constructor, so %d values had nowhere to go", c.Name, len(args))
	}
	return obj, nil
}

func (in *Interp) evalMember(x *ast.Member, env *rt.Env) (rt.Value, error) {
	recv, err := in.eval(x.X, env)
	if err != nil {
		return nil, err
	}

	target, err := in.readMember(x, recv)
	if err != nil {
		return nil, err
	}
	if !x.Call {
		return target, nil
	}
	args, err := in.evalArgs(x.Args, env)
	if err != nil {
		return nil, err
	}
	switch target.(type) {
	case *rt.Function, *rt.BoundMethod, *rt.ClassValue:
	default:

		return nil, in.faultAt(x, "E4154",
			fmt.Sprintf("«%s» یک %s است، نه تابعی که صدا زدنی", x.Name, rt.TypeName(target)),
			"%s is a %s, not something to call; write %s without brackets", x.Name, rt.TypeName(target), x.Name)
	}
	return in.callValue(x, target, nil, args)
}

func (in *Interp) readMember(x *ast.Member, recv rt.Value) (rt.Value, error) {
	switch t := recv.(type) {
	case *rt.Object:
		if v, ok := t.Get(x.Name); ok {
			return v, nil
		}
		if m, owner := t.Class.FindMethod(x.Name); m != nil {
			return &rt.BoundMethod{Recv: t, Fn: in.methodFunc(m, owner), Owner: owner}, nil
		}
		return nil, in.faultAt(x, "E4160",
			fmt.Sprintf("«%s» در «%s» نیست", x.Name, t.Class.Name),
			"%s is not a field or a method of %s", x.Name, t.Class.Name)
	case *rt.ClassValue:

		if t.Class.Enumerated {
			if _, ok := t.Class.FindField(x.Name); ok {
				if v, found := t.Class.Defaults[rt.HashKeyMust(x.Name)]; found {
					return v, nil
				}
			}
		}
		if m, owner := t.Class.FindMethod(x.Name); m != nil {
			return &rt.Function{Name: m.Name, Params: m.Params, Body: m.Body, Env: in.Globals, Class: owner}, nil
		}
		return nil, in.faultAt(x, "E4161",
			fmt.Sprintf("«%s» روی «%s» نیست", x.Name, t.Class.Name),
			"%s is not a field or a method of %s", x.Name, t.Class.Name)
	case *rt.ErrorValue:
		if x.Name == "message" {
			return t.Message, nil
		}
		if x.Name == "data" {
			return t.Data, nil
		}
		if x.Name == "name" {
			return rt.Str(t.Name), nil
		}
		return nil, in.faultAt(x, "E4162",
			fmt.Sprintf("روی khata چیزی به نام «%s» نیست", x.Name),
			"a khata has no %s; it carries message, data and name", x.Name)
	}

	if v, ok := in.collectionProperty(recv, x.Name); ok {
		return v, nil
	}
	if v, ok, handled := in.collectionMember(recv, x.Name); handled {
		if !ok {
			return nil, in.faultAt(x, "E4163",
				fmt.Sprintf("روی %s چیزی به نام «%s» نیست", rt.TypeName(recv), x.Name),
				"a %s has nothing called %s", rt.TypeName(recv), x.Name)
		}
		return v, nil
	}

	switch recv.(type) {
	case rt.Str, *rt.List, *rt.Dict, *rt.Set, *rt.Range:
		return nil, in.faultAt(x, "E4163",
			fmt.Sprintf("روی %s چیزی به نام «%s» نیست", rt.TypeName(recv), x.Name),
			"a %s has nothing called %s", rt.TypeName(recv), x.Name)
	}
	return nil, in.faultAt(x, "E4164",
		fmt.Sprintf("نمی‌شود از %s عضو خواند", rt.TypeName(recv)),
		"you cannot read a field from a %s; only a darghan, a sakht or a khata has fields", rt.TypeName(recv))
}

func (in *Interp) collectionProperty(recv rt.Value, name string) (rt.Value, bool) {
	switch name {
	case "tule", "len":
		switch t := recv.(type) {
		case rt.Str:
			return rt.Int(rt.RuneCount(string(t))), true
		case *rt.List:
			return rt.Int(len(t.Items)), true
		case *rt.Set:
			return rt.Int(t.Len()), true
		case *rt.Dict:
			return rt.Int(t.Len()), true
		case *rt.Range:
			return rt.Int(len(rangeItems(*t))), true
		}
	}
	return nil, false
}

func (in *Interp) collectionMember(recv rt.Value, name string) (rt.Value, bool, bool) {
	switch t := recv.(type) {
	case rt.Str:
		if in.nativeMethod("matn", name) {
			return in.bindMethod(name, func(ip *Interp, args []rt.Value) (rt.Value, error) {
				return ip.strMethod(t, name, args)
			}), true, true
		}
	case *rt.List:
		if in.nativeMethod("list", name) {
			return in.bindMethod(name, func(ip *Interp, args []rt.Value) (rt.Value, error) {
				return ip.listMethod(t, name, args)
			}), true, true
		}
	case *rt.Dict:
		if in.nativeMethod("dict", name) {
			return in.bindMethod(name, func(ip *Interp, args []rt.Value) (rt.Value, error) {
				return ip.dictMethod(t, name, args)
			}), true, true
		}
	}
	return nil, false, false
}

func (in *Interp) bindMethod(name string, fn func(*Interp, []rt.Value) (rt.Value, error)) rt.Value {
	return &rt.Function{
		Name: name,
		Native: native(func(ip *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
			return fn(ip, args)
		}),
	}
}

type argValue = rt.NamedArg

func (in *Interp) evalArgs(argNodes []ast.Arg, env *rt.Env) ([]rt.Value, error) {
	out := make([]rt.Value, 0, len(argNodes))
	for _, a := range argNodes {
		v, err := in.eval(a.Expr, env)
		if err != nil {
			return nil, err
		}
		if a.Name == "" {

			out = append(out, v)
			continue
		}
		out = append(out, rt.NewNamedArg(a.Name, v))
	}
	return out, nil
}

func (in *Interp) getIndex(x ast.Node, recv, at rt.Value) (rt.Value, error) {
	switch t := recv.(type) {
	case rt.Str:
		i, ok := toInt(at)
		if !ok {
			return nil, in.faultAt(x, "E4170",
				"شمارهِ حرف باید عدد باشد", "a matn index must be a whole number")
		}

		r, found := rt.RuneAt(string(t), int(i))
		if !found {
			n := rt.RuneCount(string(t))
			return nil, in.faultAt(x, "E4171",
				fmt.Sprintf("حرفِ %d در matnی به بلندی %d نیست", i, n),
				"position %d is outside a matn of length %d", i, n)
		}
		return rt.Str(string(r)), nil
	case *rt.List:
		i, ok := toInt(at)
		if !ok {
			return nil, in.faultAt(x, "E4172",
				"شماره عضو باید عدد باشد", "a list position must be a whole number")
		}
		idx := int(i)
		if idx < 0 {
			idx += len(t.Items)
		}
		if idx < 0 || idx >= len(t.Items) {
			return nil, in.faultAt(x, "E4173",
				fmt.Sprintf("عضو %d در listی به بلندی %d نیست", i, len(t.Items)),
				"position %d is outside a list of length %d", i, len(t.Items))
		}
		return t.Items[idx], nil
	case *rt.Dict:
		v, ok := t.Get(at)
		if !ok {
			return nil, in.faultAt(x, "E4174",
				fmt.Sprintf("کلید «%s» در dict نیست", rt.Show(at)),
				"there is no %s in this dict", rt.Show(at))
		}
		return v, nil
	case *rt.Set:
		items := t.Items()
		i, ok := toInt(at)
		if !ok {
			return nil, in.faultAt(x, "E4175",
				"شماره عضو set باید عدد باشد", "a set position must be a whole number")
		}
		idx := int(i)
		if idx < 0 {
			idx += len(items)
		}
		if idx < 0 || idx >= len(items) {
			return nil, in.faultAt(x, "E4176",
				fmt.Sprintf("عضو %d در setی به بلندی %d نیست", i, len(items)),
				"position %d is outside a set of length %d", i, len(items))
		}
		return items[idx], nil
	case *rt.Range:
		items := rangeItems(*t)
		i, ok := toInt(at)
		if !ok {
			return nil, in.faultAt(x, "E4177",
				"شماره عضو بازه باید عدد باشد", "a range position must be a whole number")
		}
		idx := int(i)
		if idx < 0 {
			idx += len(items)
		}
		if idx < 0 || idx >= len(items) {
			return nil, in.faultAt(x, "E4178",
				fmt.Sprintf("عضو %d در بازه‌ای به بلندی %d نیست", i, len(items)),
				"position %d is outside a range of length %d", i, len(items))
		}
		return items[idx], nil
	}
	return nil, in.faultAt(x, "E4179",
		fmt.Sprintf("%s را نمی‌شود با [] خواند", rt.TypeName(recv)),
		"you cannot use [] with a %s", rt.TypeName(recv))
}

func (in *Interp) setIndex(node ast.Node, recv, at, v rt.Value) error {
	switch t := recv.(type) {
	case *rt.List:
		i, ok := toInt(at)
		if !ok {
			return in.faultAt(node, "E4180", "شماره عضو باید عدد باشد", "a list position must be a whole number")
		}
		idx := int(i)
		if idx < 0 {
			idx += len(t.Items)
		}
		if idx < 0 || idx >= len(t.Items) {
			return in.faultAt(node, "E4181",
				fmt.Sprintf("عضو %d در listی به بلندی %d نیست", i, len(t.Items)),
				"position %d is outside a list of length %d", i, len(t.Items))
		}
		t.Items[idx] = v
		return nil
	case *rt.Dict:
		if _, ok := rt.HashKey(at); !ok {
			return in.faultAt(node, "E4182",
				"کلید dict باید عدد، matn، bool یا hich باشد",
				"a dict key has to be a number, matn, bool or hich, not %s", rt.TypeName(at))
		}
		t.Set(at, v)
		return nil
	}
	return in.faultAt(node, "E4183",
		fmt.Sprintf("روی %s با [] نمی‌شود مقدار گذاشت", rt.TypeName(recv)),
		"you cannot store into a %s with []", rt.TypeName(recv))
}

func (in *Interp) setMember(node ast.Node, recv rt.Value, name string, v rt.Value) error {
	obj, ok := recv.(*rt.Object)
	if !ok {
		return in.faultAt(node, "E4184",
			fmt.Sprintf("روی %s نمی‌شود عضو گذاشت", rt.TypeName(recv)),
			"you cannot set a field on a %s; only a darghan or a sakht has fields", rt.TypeName(recv))
	}
	if _, exists := obj.Class.FindField(name); !exists {
		return in.faultAt(node, "E4185",
			fmt.Sprintf("«%s» در «%s» نیست", name, obj.Class.Name),
			"%s is not a field of %s, so there is nowhere to put a value", name, obj.Class.Name)
	}
	obj.Set(name, v)
	return nil
}

func (in *Interp) evalSlice(x *ast.Slice, env *rt.Env) (rt.Value, error) {
	recv, err := in.eval(x.X, env)
	if err != nil {
		return nil, err
	}
	var lo, hi rt.Value
	if x.Lo != nil {
		if lo, err = in.eval(x.Lo, env); err != nil {
			return nil, err
		}
	}
	if x.Hi != nil {
		if hi, err = in.eval(x.Hi, env); err != nil {
			return nil, err
		}
	}
	return in.sliceOf(x, recv, lo, hi, x.Inclusive)
}

func (in *Interp) sliceOf(x ast.Node, recv, lo, hi rt.Value, inclusive bool) (rt.Value, error) {
	bounds := func(n int) (int, int, bool) {
		start, end := 0, n
		if lo != nil {
			if i, ok := toInt(lo); ok {
				start = int(i)
			}
		}
		if hi != nil {
			if i, ok := toInt(hi); ok {
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
		return start, end, true
	}
	switch t := recv.(type) {
	case *rt.List:
		s, e, _ := bounds(len(t.Items))
		out := make([]rt.Value, e-s)
		copy(out, t.Items[s:e])
		return rt.NewList(out...), nil
	case rt.Str:

		loI, hiI := 0, rt.RuneCount(string(t))
		if lo != nil {
			if n, ok := toInt(lo); ok {
				loI = int(n)
			}
		}
		if hi != nil {
			if n, ok := toInt(hi); ok {
				hiI = int(n)
			}
		}
		return rt.Str(rt.Slice(string(t), loI, hiI, inclusive)), nil
	case *rt.Set:
		items := t.Items()
		s, e, _ := bounds(len(items))
		out := make([]rt.Value, e-s)
		copy(out, items[s:e])
		return rt.NewList(out...), nil
	case *rt.Range:
		items := rangeItems(*t)
		s, e, _ := bounds(len(items))
		out := make([]rt.Value, e-s)
		copy(out, items[s:e])
		return rt.NewList(out...), nil
	}
	return nil, in.faultAt(x, "E4186",
		fmt.Sprintf("از %s نمی‌شوت برش گرفت", rt.TypeName(recv)),
		"you cannot take a slice of a %s", rt.TypeName(recv))
}

func (in *Interp) convertTo(node ast.Node, v rt.Value, typeName string) (rt.Value, error) {
	base := typeName
	if i := strings.IndexByte(base, '<'); i > 0 {
		base = base[:i]
	}

	switch base {
	case "int":
		base = "arz"
	case "float", "double":
		base = "kasr"
	case "string", "str":
		base = "matn"
	case "boolean":
		base = "bool"
	}
	switch base {
	case "arz":
		switch t := v.(type) {
		case rt.Int:
			return t, nil
		case rt.Bool:
			if t {
				return rt.Int(1), nil
			}
			return rt.Int(0), nil
		case rt.Float:
			return rt.Int(math.Trunc(float64(t))), nil
		case rt.Str:
			if i, ok := parseWhole(string(t)); ok {
				return rt.Int(i), nil
			}
		}
	case "kasr":
		if s, ok := v.(rt.Str); ok {
			if f, err := strconv.ParseFloat(strings.TrimSpace(string(s)), 64); err == nil {
				return rt.Float(f), nil
			}
			break
		}
		if f, ok := toFloat(v); ok {
			return rt.Float(f), nil
		}
	case "matn":
		if _, ok := v.(rt.Str); ok {
			return v, nil
		}

		return rt.Str(rt.Show(v)), nil
	case "bool":
		return rt.Bool(rt.Truthy(v)), nil
	}
	return nil, in.faultAt(node, "E4190",
		fmt.Sprintf("نمی‌شود %s را «%s» کرد", rt.TypeName(v), typeName),
		"a %s cannot become %s", rt.TypeName(v), typeName)
}

func parseWhole(s string) (rt.Int, bool) {
	i, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, false
	}
	return rt.Int(i), true
}

func faultf(node ast.Node, code, fa, enFormat string, args ...any) error {
	return &rt.Panic{
		Value: &rt.ErrorValue{Message: rt.Str(fmt.Sprintf(enFormat, args...)), Data: rt.Nil, Name: code, FA: persian(fa, args...)},
		Trace: []rt.Frame{{Pos: node.Span().Start}},
	}
}
