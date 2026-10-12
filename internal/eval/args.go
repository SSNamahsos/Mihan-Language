package eval

import (
	"math"
	"strconv"
	"strings"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

func fnArgs(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	all := make([]rt.Value, 0, len(in.Args))
	for _, a := range in.Args {
		all = append(all, rt.Str(a))
	}
	if len(args) == 0 {
		return rt.NewList(all...), nil
	}
	if len(args) > 1 {
		return rt.Nil, in.faultAt(node, "E4121",
			"`arghoman` یک عدد می‌خواهد یا هیچ",
			"arghoman takes one number or none")
	}
	i, ok := toInt(args[0])
	if !ok {
		return rt.Nil, in.faultAt(node, "E4122",
			"شماره‌ی یک آرگومان باید عدد باشد",
			"the position of an argument must be a number")
	}

	if i < 0 || int(i) >= len(all) {
		return rt.Nil, in.faultFA(node, "E4123",
			"فقط %d آرگومان داده شده، و شماره‌ی %d خواسته شد",
			[]any{len(all), i},
			"there are only %d arguments, and %d was asked for", len(all), i)
	}
	return all[i], nil
}

func fnNumber(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	s, ok := args[0].(rt.Str)
	if !ok {
		return rt.Nil, in.faultAt(node, "E4124",
			"عدد باید متن باشد",
			"a number has to be written as text first")
	}
	word := strings.TrimSpace(string(s))
	n, err := strconv.ParseFloat(word, 64)
	if err != nil {
		return rt.Nil, in.faultFA(node, "E4125",
			"«%s» عدد نیست",
			[]any{word},
			"%s is not a number", word)
	}

	if n == math.Trunc(n) && math.Abs(n) < 1e18 {
		return rt.Int(int64(n)), nil
	}
	return rt.Float(n), nil
}
