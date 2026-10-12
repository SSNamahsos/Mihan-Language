package eval

import (
	"fmt"
	"strings"

	rt "mihan/internal/runtime"
)

func (in *Interp) strMethod(s rt.Str, name string, args []rt.Value) (rt.Value, error) {
	switch name {
	case "gooshband", "upper":
		return rt.Str(strings.ToUpper(string(s))), nil
	case "kiawtar", "lower":
		return rt.Str(strings.ToLower(string(s))), nil
	case "jabeja", "trim", "azadi", "strip":
		return rt.Str(strings.TrimSpace(string(s))), nil
	case "barmigefth", "startsWith":
		return rt.Bool(strings.HasPrefix(string(s), argText(args, 0))), nil
	case "bapakhsh", "endsWith":
		return rt.Bool(strings.HasSuffix(string(s), argText(args, 0))), nil
	case "shamil", "includes":
		return rt.Bool(strings.Contains(string(s), argText(args, 0))), nil
	case "jigarre", "replace":
		return rt.Str(strings.ReplaceAll(string(s), argText(args, 0), argText(args, 1))), nil
	case "tab", "split":
		sep := argText(args, 0)
		parts := strings.Split(string(s), sep)
		out := make([]rt.Value, len(parts))
		for i, p := range parts {
			out[i] = rt.Str(p)
		}
		return rt.NewList(out...), nil
	case "por", "join":

		t := string(s)
		n := rt.RuneCount(t)
		out := make([]rt.Value, n)
		for i := 0; i < n; i++ {
			r, _ := rt.RuneAt(t, i)
			out[i] = rt.Str(string(r))
		}
		return fnJoin(in, []rt.Value{rt.NewList(out...), args[0]}, nil)
	case "tule", "len":
		return rt.Int(rt.RuneCount(string(s))), nil
	case "banafsh", "hash":
		return fnHash(in, []rt.Value{s}, nil)
	}
	return nil, fmt.Errorf("matn has no method %q", name)
}

func (in *Interp) listMethod(l *rt.List, name string, args []rt.Value) (rt.Value, error) {
	switch name {
	case "bast", "push":
		l.Items = append(l.Items, args[0])
		return l, nil
	case "bardavar", "pop":
		if len(l.Items) == 0 {
			return rt.Nil, fmt.Errorf("this list is empty")
		}
		last := l.Items[len(l.Items)-1]
		l.Items = l.Items[:len(l.Items)-1]
		return last, nil
	case "biya", "insert":
		return fnInsert(in, []rt.Value{l, args[0], args[1]}, nil)
	case "hazar", "index":
		return fnIndexOf(in, []rt.Value{l, args[0]}, nil)
	case "barmadar", "contains":
		return fnContains(in, []rt.Value{l, args[0]}, nil)
	case "taghib", "sort":
		return fnSort(in, []rt.Value{l}, nil)
	case "forogh", "reverse":
		return fnReverse(in, []rt.Value{l}, nil)
	case "por", "join":
		return fnJoin(in, []rt.Value{l, args[0]}, nil)
	case "tule", "len", "shomar", "count":
		return rt.Int(len(l.Items)), nil
	case "filter":
		return fnFilter(in, []rt.Value{l, args[0]}, nil)
	case "tabdil_kol", "map":
		return fnMap(in, []rt.Value{l, args[0]}, nil)
	case "kamil", "reduce":
		all := append([]rt.Value{l}, args...)
		return fnReduce(in, all, nil)
	case "kamej", "any":
		return fnAny(in, []rt.Value{l, args[0]}, nil)
	case "hamej", "all":
		return fnAll(in, []rt.Value{l, args[0]}, nil)
	case "setare", "setof":
		return fnSetOf(in, []rt.Value{l}, nil)
	}
	return nil, fmt.Errorf("list has no method %q", name)
}

func (in *Interp) dictMethod(d *rt.Dict, name string, args []rt.Value) (rt.Value, error) {
	if n, ok := dictMethodArgs[name]; ok && len(args) < n {
		return nil, fmt.Errorf("the dict method %s needs %d value(s), and was given %d", name, n, len(args))
	}
	switch name {
	case "begir", "get":

		if v, ok := d.Get(args[0]); ok {
			return v, nil
		}
		if len(args) > 1 {
			return args[1], nil
		}
		return rt.Nil, nil
	case "dast", "keys":
		out := make([]rt.Value, 0, d.Len())
		out = append(out, d.Keys()...)
		return rt.NewList(out...), nil
	case "meghdar", "values":
		return fnValues(in, []rt.Value{d}, nil)
	case "tule", "len":
		return rt.Int(d.Len()), nil
	case "barmadar", "contains", "has_key":
		_, ok := d.Get(args[0])
		return rt.Bool(ok), nil
	case "hazar", "index":
		for i, k := range d.Keys() {
			if rt.Equal(k, args[0]) {
				return rt.Int(i), nil
			}
		}
		return rt.Int(-1), nil
	case "bast", "put":
		d.Set(args[0], args[1])
		return d, nil
	case "bardavar", "pop":
		v, ok := d.Get(args[0])
		if !ok {
			return rt.Nil, fmt.Errorf("there is no %s in this dict", rt.Show(args[0]))
		}
		d.Delete(args[0])
		return v, nil
	case "hazf", "delete":
		if len(args) < 1 {
			return rt.Nil, fmt.Errorf("hazf needs the key to remove")
		}
		return rt.Bool(d.Delete(args[0])), nil
	}
	return nil, fmt.Errorf("dict has no method %q", name)
}

func argText(args []rt.Value, i int) string {
	if i < len(args) {
		if s, ok := args[i].(rt.Str); ok {
			return string(s)
		}
		return rt.Show(args[i])
	}
	return ""
}

func stepRangeItems(r rt.StepRange) []rt.Value {
	var out []rt.Value
	if r.Step == 0 {
		return out
	}
	step := r.Step
	if step < 0 {
		step = -step
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
		_ = step
	}
	return out
}

var dictMethodArgs = map[string]int{"barmadar": 1, "contains": 1, "has_key": 1, "hazar": 1, "index": 1,
	"bast": 2, "put": 2, "bardavar": 1, "pop": 1, "begir": 1, "get": 1}
