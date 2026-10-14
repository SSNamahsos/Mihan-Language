package eval

import (
	"strings"
	"time"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

func fnNumberOrNil(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	switch args[0].(type) {
	case rt.Int, rt.Float:
		return args[0], nil
	case rt.Str:
	default:
		return rt.Nil, nil
	}
	v, err := fnNumber(in, args, node)
	if err != nil {
		return rt.Nil, nil
	}
	return v, nil
}

func fnTime(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	return rt.Float(float64(time.Now().UnixNano()) / 1e9), nil
}

var dateLayout = strings.NewReplacer(
	"YYYY", "2006", "MM", "01", "DD", "02",
	"HH", "15", "mm", "04", "ss", "05",
)

func fnDate(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	pattern := "YYYY-MM-DD HH:mm:ss"
	if len(args) > 0 {
		s, ok := args[0].(rt.Str)
		if !ok {
			return nil, in.faultAt(node, "E4127",
				"olgoo-ye tarikh bayad matn bashe, mesl-e \"YYYY-MM-DD\"",
				"the date pattern must be text such as \"YYYY-MM-DD\", not a %s", rt.TypeName(args[0]))
		}
		pattern = string(s)
	}
	at := time.Now()
	if len(args) > 1 {
		f, ok := toFloat(args[1])
		if !ok {
			return nil, in.faultAt(node, "E4127",
				"zaman bayad adad bashe (sanie az 1970, hamoon ke zaman mide)",
				"the moment must be a number of seconds since 1970 (what zaman gives), not a %s", rt.TypeName(args[1]))
		}
		at = time.Unix(0, int64(f*1e9))
	}
	return rt.Str(at.Format(dateLayout.Replace(pattern))), nil
}

func fnSleep(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	f, ok := toFloat(args[0])
	if !ok || f < 0 {
		return nil, in.faultAt(node, "E4127",
			"sabr yek adad-e sanie mikhad",
			"sabr needs a number of seconds, not %s", rt.Show(args[0]))
	}
	if f > 0 {
		whileWaiting(func() { time.Sleep(time.Duration(f * float64(time.Second))) })
	}
	return rt.Nil, nil
}
