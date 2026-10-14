package eval

import (
	"mihan/internal/ast"
	rt "mihan/internal/runtime"
	"strings"
)

func fnJabe(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	rest, style := takeStyle(args)
	return styled("jabe", "", piecesOf(rest), nil, style), nil
}

func fnBand(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	return textPiece(in, "band", args, node)
}

func fnOnvan(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	return textPiece(in, "onvan", args, node)
}

func fnOnvan2(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	return textPiece(in, "onvan2", args, node)
}

func fnDokme(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	rest, style := takeStyle(args)
	if len(rest) < 1 {
		return nil, in.faultAt(node, "E4354",
			"dokme yek neveshte mikhad",
			"dokme needs the text on the button")
	}
	label, ok := rest[0].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4354",
			"neveshte-ye dokme bayad matn bashe",
			"the text on a dokme has to be text, not %s", rt.TypeName(rest[0]))
	}
	d := styled("dokme", string(label), nil, nil, style)
	if len(rest) >= 2 {
		kar, ok := rest[1].(rt.Str)
		if !ok {
			return nil, in.faultAt(node, "E4354",
				"kar bayad matn bashe",
				"the kar of a dokme has to be text, not %s", rt.TypeName(rest[1]))
		}
		d.Set(rt.Str("kar"), kar)
	}
	return d, nil
}

func fnPeyvand(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	rest, style := takeStyle(args)
	if len(rest) < 2 {
		return nil, in.faultAt(node, "E4354",
			"peyvand yek neveshte va yek neshani mikhad",
			"peyvand needs the text and the address it goes to")
	}
	label, ok1 := rest[0].(rt.Str)
	href, ok2 := rest[1].(rt.Str)
	if !ok1 || !ok2 {
		return nil, in.faultAt(node, "E4354",
			"neveshte va neshani bayad matn bashan",
			"the text and the address of a peyvand have to be text")
	}
	d := styled("peyvand", string(label), nil, map[string]rt.Value{
		"rang": rt.Str("abi"),
	}, style)
	d.Set(rt.Str("be"), href)
	return d, nil
}

func fnVorudi(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	rest, style := takeStyle(args)
	if len(rest) < 1 {
		return nil, in.faultAt(node, "E4354",
			"vorudi yek esm mikhad",
			"vorudi needs the name of the field")
	}
	name, ok := rest[0].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4354",
			"esm-e vorudi bayad matn bashe",
			"the name of a vorudi has to be text")
	}
	d := styled("vorudi", "", nil, map[string]rt.Value{
		"khat_dor": rt.Int(8),
		"fasele":   rt.Int(8),
	}, style)
	d.Set(rt.Str("nam"), name)
	return d, nil
}

func fnForm(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	rest, style := takeStyle(args)
	if len(rest) < 2 {
		return nil, in.faultAt(node, "E4354",
			"form yek neshani va ghete-haye dakhelesh ro mikhad",
			"form needs the address it sends to, and the pieces inside it")
	}
	action, ok := rest[0].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4354",
			"neshani-ye form bayad matn bashe",
			"the address of a form has to be text")
	}
	d := styled("form", "", piecesOf(rest[1:]), map[string]rt.Value{
		"nemayesh":     rt.Str("radif"),
		"fasele_khane": rt.Int(8),
	}, style)
	d.Set(rt.Str("be"), action)
	d.Set(rt.Str("ravesh"), rt.Str("ferest"))
	return d, nil
}

func fnKarte(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	rest, style := takeStyle(args)
	return styled("jabe", "", piecesOf(rest), map[string]rt.Value{
		"khat_dor":     rt.Int(16),
		"fasele":       rt.Int(20),
		"pas_zamine":   rt.Str("#1a1e25"),
		"rang":         rt.Str("sefid"),
		"fasele_birun": rt.Int(12),
	}, style), nil
}

func fnNavar(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	rest, style := takeStyle(args)
	return styled("navar", "", piecesOf(rest), map[string]rt.Value{
		"nemayesh":     rt.Str("radif"),
		"fasele_khane": rt.Int(16),
		"fasele":       rt.Int(16),
		"pas_zamine":   rt.Str("#1a1e25"),
	}, style), nil
}

func fnOfoghi(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	rest, style := takeStyle(args)
	return styled("jabe", "", piecesOf(rest), map[string]rt.Value{
		"nemayesh":     rt.Str("radif"),
		"fasele_khane": rt.Int(12),
	}, style), nil
}

func fnAmudi(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	rest, style := takeStyle(args)
	return styled("jabe", "", piecesOf(rest), map[string]rt.Value{
		"nemayesh":     rt.Str("sotun"),
		"fasele_khane": rt.Int(12),
	}, style), nil
}

func fnListe(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	rest, style := takeStyle(args)
	if len(rest) < 1 {
		return nil, in.faultAt(node, "E4354",
			"liste yek list mikhad",
			"liste needs a list")
	}
	items, ok := rest[0].(*rt.List)
	if !ok {
		return nil, in.faultAt(node, "E4354",
			"liste yek list mikhad",
			"liste needs a list, not %s", rt.TypeName(rest[0]))
	}
	rows := make([]rt.Value, len(items.Items))
	for i, item := range items.Items {
		if d, ok := item.(*rt.Dict); ok {
			rows[i] = d
			continue
		}
		rows[i] = styled("radif", rt.Show(item), nil, nil, nil)
	}
	return styled("fehrest", "", rt.NewList(rows...), nil, style), nil
}

func fnJadval(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	rest, style := takeStyle(args)
	if len(rest) < 2 {
		return nil, in.faultAt(node, "E4354",
			"jadval sarsotoon-ha va radif-ha ro mikhad",
			"jadval needs the headings and the rows")
	}
	heads, ok1 := rest[0].(*rt.List)
	rows, ok2 := rest[1].(*rt.List)
	if !ok1 || !ok2 {
		return nil, in.faultAt(node, "E4354",
			"sarsotoon va radif bayad list bashan",
			"the headings and the rows of a jadval have to be lists")
	}
	headCells := make([]rt.Value, len(heads.Items))
	for i, h := range heads.Items {
		headCells[i] = styled("sarnevesht", rt.Show(h), nil, nil, nil)
	}
	body := []rt.Value{styled("satr", "", rt.NewList(headCells...), nil, nil)}
	for _, row := range rows.Items {
		cells, ok := row.(*rt.List)
		if !ok {
			return nil, in.faultAt(node, "E4354",
				"har radif-e jadval yek list ast",
				"each row of a jadval is a list")
		}
		one := make([]rt.Value, len(cells.Items))
		for i, c := range cells.Items {
			one[i] = styled("khane_jadval", rt.Show(c), nil, nil, nil)
		}
		body = append(body, styled("satr", "", rt.NewList(one...), nil, nil))
	}
	return styled("jadval", "", rt.NewList(body...), nil, style), nil
}

func fnRah(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	req, ok := args[0].(*rt.Dict)
	routes, ok2 := args[1].(*rt.List)
	if !ok || !ok2 {
		return nil, in.faultAt(node, "E4354",
			"rah darkhast va list-e rah-ha ro mikhad",
			"rah needs the request and the list of paths")
	}
	path := dictStr(req, "path")
	var chosen rt.Value
	title := "Mihan"
	wrongMethod := false
	for _, item := range routes.Items {
		row, ok := item.(*rt.Dict)
		if !ok {
			return nil, in.faultAt(node, "E4354",
				"har rah yek dict ast",
				"each path given to rah has to be a dict")
		}
		params := rt.NewDict()
		if !matchRoute(dictStr(row, "path"), path, params) {
			continue
		}
		if m := dictStr(row, "method"); m != "" && !strings.EqualFold(m, dictStr(req, "method")) {
			wrongMethod = true
			continue
		}
		if len(params.Keys()) > 0 {
			req.Set(rt.Str("params"), params)
		}
		v, found := row.Get(rt.Str("tabe"))
		if !found {
			return nil, in.faultAt(node, "E4354",
				"in rah funct nadare",
				"the path %s has no funct", path)
		}
		chosen = v
		if t := dictStr(row, "onvan"); t != "" {
			title = t
		}
		break
	}
	if chosen == nil && wrongMethod {
		return pageAnswer(405, "this path does not take "+dictStr(req, "method"), "text/plain; charset=utf-8"), nil
	}
	if chosen == nil {
		missing := styled("band", "nist", nil, nil, nil)
		page, err := fnPage(in, []rt.Value{rt.Str("nist"), missing}, node)
		if err != nil {
			return nil, err
		}
		return htmlAnswer(404, page), nil
	}
	result, err := in.callValue(node, chosen, nil, []rt.Value{req})
	if err != nil {
		return nil, err
	}
	if d, ok := result.(*rt.Dict); ok {
		if _, has := d.Get(rt.Str("vaziat")); has {
			return d, nil
		}
	}
	if s, ok := result.(rt.Str); ok {
		return htmlAnswer(200, s), nil
	}
	page, err := fnPage(in, []rt.Value{rt.Str(title), result}, node)
	if err != nil {
		return nil, err
	}
	return htmlAnswer(200, page), nil
}

func textPiece(in *Interp, kind string, args []rt.Value, node ast.Node) (rt.Value, error) {
	rest, style := takeStyle(args)
	if len(rest) < 1 {
		return nil, in.faultAt(node, "E4354",
			kind+" yek neveshte mikhad",
			"%s needs text", kind)
	}
	s, ok := rest[0].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4354",
			"neveshte bayad matn bashe",
			"the text of %s has to be text, not %s", kind, rt.TypeName(rest[0]))
	}
	return styled(kind, string(s), nil, nil, style), nil
}

func styled(kind, text string, inner rt.Value, defaults map[string]rt.Value, style *rt.Dict) *rt.Dict {
	d := rt.NewDict()
	d.Set(rt.Str("gone"), rt.Str(kind))
	if text != "" {
		d.Set(rt.Str("matn"), rt.Str(text))
	}
	if inner != nil {
		d.Set(rt.Str("darun"), inner)
	}
	for k, v := range defaults {
		d.Set(rt.Str(k), v)
	}
	if style != nil {
		for _, k := range style.Keys() {
			v, _ := style.Get(k)
			d.Set(k, v)
		}
	}
	return d
}

func piecesOf(args []rt.Value) rt.Value {
	if len(args) == 1 {
		if l, ok := args[0].(*rt.List); ok {
			return l
		}
	}
	return rt.NewList(args...)
}

func takeStyle(args []rt.Value) ([]rt.Value, *rt.Dict) {
	if len(args) == 0 {
		return args, nil
	}
	d, ok := args[len(args)-1].(*rt.Dict)
	if !ok || !isStyleDict(d) {
		return args, nil
	}
	return args[:len(args)-1], d
}

func isStyleDict(d *rt.Dict) bool {
	if dictStr(d, "gone") != "" {
		return false
	}
	if _, ok := childrenOf(d); ok {
		return false
	}
	if d.Len() == 0 {
		return true
	}
	for _, k := range d.Keys() {
		name, ok := jsonKey(k)
		if !ok {
			return false
		}
		if name == "matn" {
			return false
		}
		if _, known := styleByName(name); known {
			continue
		}
		if structureKeys[name] {
			continue
		}
		return false
	}
	return true
}

func htmlAnswer(status int, body rt.Value) rt.Value {
	h := rt.NewDict()
	h.Set(rt.Str("Content-Type"), rt.Str("text/html; charset=utf-8"))
	d := rt.NewDict()
	d.Set(rt.Str("vaziat"), rt.Int(int64(status)))
	d.Set(rt.Str("matn"), body)
	d.Set(rt.Str("sarlawah"), h)
	return d
}
