package eval

import (
	"html"
	"net/url"
	"sort"
	"strings"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

var structureKeys = map[string]bool{
	"gone": true, "matn": true, "darun": true, "bache": true,
	"nam": true, "be": true, "ravesh": true, "noe": true, "jaye": true,
	"kar": true, "kelas": true, "shenase": true,
}

var goneTag = map[string]string{
	"jabe": "div", "band": "p", "onvan": "h1", "onvan2": "h2", "onvan3": "h3",
	"dokme": "button", "peyvand": "a", "fehrest": "ul", "radif": "li",
	"form": "form", "vorudi": "input", "tasvir": "img", "khat": "hr",
	"jadval": "table", "satr": "tr", "sarnevesht": "th", "khane_jadval": "td",
	"navar": "nav",
}

func fnPage(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	title := "Mihan"
	body := args[0]
	if len(args) >= 2 {
		s, ok := args[0].(rt.Str)
		if !ok {
			return nil, in.faultAt(node, "E4351",
				"onvan-e safhe bayad matn bashe",
				"the page title has to be text, not %s", rt.TypeName(args[0]))
		}
		title = string(s)
		body = args[1]
	}
	var b strings.Builder
	b.WriteString("<!doctype html>\n<html lang=\"fa\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString("<title>")
	b.WriteString(html.EscapeString(title))
	b.WriteString("</title>\n<style>\n")
	b.WriteString("body{margin:0;font-family:system-ui,sans-serif;line-height:1.5}\n")
	b.WriteString("button,input{font:inherit}\n")
	if len(args) == 3 {
		css, err := styleSheet(args[2])
		if err != nil {
			return nil, in.faultAt(node, "E4355", err.fa, "%s", err.en)
		}
		b.WriteString(css)
	}
	b.WriteString("</style>\n</head>\n<body>\n")
	if err := in.writePageBody(&b, body, node); err != nil {
		return nil, err
	}

	b.WriteString(pageScript)
	b.WriteString("</body>\n</html>\n")
	return rt.Str(b.String()), nil
}

func (in *Interp) writePageBody(b *strings.Builder, body rt.Value, node ast.Node) error {
	switch body.(type) {
	case *rt.Dict, *rt.List, rt.Str:
		return in.writePiece(b, body, node)
	default:
		return in.faultAt(node, "E4351",
			"safhe_kon yek jabe ya yek list az jabe-ha mikhad",
			"safhe_kon needs a box or a list of boxes, not %s", rt.TypeName(body))
	}
}

func (in *Interp) writePiece(b *strings.Builder, v rt.Value, node ast.Node) error {
	switch t := v.(type) {
	case *rt.List:
		for _, item := range t.Items {
			if err := in.writePiece(b, item, node); err != nil {
				return err
			}
		}
		return nil
	case rt.Str:
		b.WriteString(html.EscapeString(string(t)))
		return nil
	case rt.Int, rt.Float, rt.Bool, rt.Hich:
		b.WriteString(html.EscapeString(rt.Show(t)))
		return nil
	case *rt.Dict:
		return in.writeBox(b, t, node)
	default:
		return in.faultAt(node, "E4351",
			"in ro nemishe too safhe gozasht",
			"a %s cannot be put on a page", rt.TypeName(v))
	}
}

func (in *Interp) writeBox(b *strings.Builder, d *rt.Dict, node ast.Node) error {
	if err := in.unknownStyle(d, node); err != nil {
		return err
	}
	kind := dictStr(d, "gone")
	if kind == "" {
		kind = "jabe"
	}
	tag, ok := goneTag[kind]
	if !ok {
		return in.faultAt(node, "E4352",
			"in no-e ghete ro nemishnasam: "+kind,
			"this kind of piece is not one I know: %s", kind)
	}
	style, err := cssOf(d)
	if err != nil {
		return in.faultAt(node, "E4350", err.fa, "%s", err.en)
	}
	b.WriteByte('<')
	b.WriteString(tag)
	writeAttr(b, "style", style)
	writeAttr(b, "class", dictStr(d, "kelas"))
	writeAttr(b, "id", dictStr(d, "shenase"))
	switch kind {
	case "peyvand":
		writeAttr(b, "href", dictStr(d, "be"))
	case "tasvir":
		writeAttr(b, "src", dictStr(d, "be"))
		writeAttr(b, "alt", dictStr(d, "matn"))
	case "form":
		action := dictStr(d, "be")
		if action == "" {
			action = "/"
		}
		writeAttr(b, "action", action)
		writeAttr(b, "method", formMethod(dictStr(d, "ravesh")))
	case "vorudi":
		writeAttr(b, "name", dictStr(d, "nam"))
		writeAttr(b, "type", inputType(dictStr(d, "noe")))
		writeAttr(b, "placeholder", dictStr(d, "jaye"))
		writeAttr(b, "value", dictStr(d, "matn"))
	case "dokme":
		if kar := dictStr(d, "kar"); kar != "" {
			writeAttr(b, "type", "button")
			writeAttr(b, "data-kar", kar)
		} else {
			writeAttr(b, "type", "submit")
		}
	}
	if kind == "vorudi" || kind == "tasvir" || kind == "khat" {
		b.WriteString(">\n")
		return nil
	}
	b.WriteByte('>')
	if matn := dictStr(d, "matn"); matn != "" && kind != "tasvir" {
		b.WriteString(html.EscapeString(matn))
	}
	if inner, ok := childrenOf(d); ok {
		if err := in.writePiece(b, inner, node); err != nil {
			return err
		}
	}
	b.WriteString("</")
	b.WriteString(tag)
	b.WriteString(">\n")
	return nil
}

func (in *Interp) unknownStyle(d *rt.Dict, node ast.Node) error {
	for _, k := range d.Keys() {
		name, ok := jsonKey(k)
		if !ok {
			continue
		}
		if structureKeys[name] {
			continue
		}
		if _, known := styleByName(name); known {
			continue
		}
		msg := unknownStyleMsg(name)
		return in.faultAt(node, "E4350", msg, "%s", msg)
	}
	return nil
}

type pageFail struct {
	fa string
	en string
}

func childrenOf(d *rt.Dict) (rt.Value, bool) {
	if v, ok := d.Get(rt.Str("darun")); ok {
		return v, true
	}
	if v, ok := d.Get(rt.Str("bache")); ok {
		return v, true
	}
	return nil, false
}

func dictStr(d *rt.Dict, key string) string {
	v, ok := d.Get(rt.Str(key))
	if !ok {
		return ""
	}
	s, ok := v.(rt.Str)
	if !ok {
		return ""
	}
	return string(s)
}

func writeAttr(b *strings.Builder, name, value string) {
	if value == "" {
		return
	}
	b.WriteByte(' ')
	b.WriteString(name)
	b.WriteString("=\"")
	b.WriteString(html.EscapeString(value))
	b.WriteByte('"')
}

const pageScript = `<script>
document.addEventListener("click", function (e) {
  var el = e.target;
  while (el && !el.getAttribute("data-kar")) el = el.parentElement;
  if (!el) return;
  e.preventDefault();
  fetch(location.pathname + location.search, {
    method: "POST",
    headers: {"Content-Type": "application/x-www-form-urlencoded"},
    body: "kar=" + encodeURIComponent(el.getAttribute("data-kar"))
  }).then(function (r) { return r.text(); }).then(function (t) {
    document.open(); document.write(t); document.close();
  });
});
document.addEventListener("submit", function (e) {
  var f = e.target;
  if (!f || f.tagName !== "FORM") return;
  e.preventDefault();
  var data = new URLSearchParams(new FormData(f));
  fetch(f.getAttribute("action") || location.pathname, {
    method: (f.getAttribute("method") || "post").toUpperCase(),
    body: data
  }).then(function (r) { return r.text(); }).then(function (t) {
    document.open(); document.write(t); document.close();
  });
});
</script>
`

func formMethod(s string) string {
	switch s {
	case "", "ferest", "post":
		return "post"
	case "begir", "get":
		return "get"
	default:
		return s
	}
}

func inputType(s string) string {
	switch s {
	case "", "matn", "text":
		return "text"
	case "adad", "number":
		return "number"
	case "ramz", "password":
		return "password"
	default:
		return s
	}
}

func fnFormRead(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	s, ok := args[0].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4353",
			"form_bekhan matn-e form ro mikhad",
			"form_bekhan needs the text of a form, not %s", rt.TypeName(args[0]))
	}
	d := rt.NewDict()
	raw := string(s)
	if raw == "" {
		return d, nil
	}

	q, err := url.ParseQuery(raw)
	if err != nil {
		return nil, in.faultAt(node, "E4353",
			"in matn-e yek form nist",
			"this text is not a form: %v", err)
	}
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := q[k]
		if len(vs) > 0 {
			d.Set(rt.Str(k), rt.Str(vs[0]))
		}
	}
	return d, nil
}
