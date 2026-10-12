package eval

import (
	"html"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

type styleWord struct {
	name   string
	css    string
	length bool
}

var styleWords = []styleWord{
	{"rang", "color", false},
	{"pas_zamine", "background", false},
	{"khat", "border", false},
	{"khat_rang", "border-color", false},
	{"khat_dor", "border-radius", true},
	{"andaze", "font-size", true},
	{"khane", "font-family", false},
	{"vazn", "font-weight", false},
	{"chid", "text-align", false},
	{"pahna", "width", true},
	{"bala", "height", true},
	{"fasele", "padding", true},
	{"fasele_birun", "margin", true},
	{"fasele_khane", "gap", true},
	{"saye", "box-shadow", false},
	{"shafaf", "opacity", false},
	{"nemayesh", "display", false},
}

var structureKeys = map[string]bool{
	"gone": true, "matn": true, "darun": true, "bache": true,
	"nam": true, "be": true, "ravesh": true, "noe": true, "jaye": true,
	"kar": true,
}

var goneTag = map[string]string{
	"jabe": "div", "band": "p", "onvan": "h1", "onvan2": "h2", "onvan3": "h3",
	"dokme": "button", "peyvand": "a", "fehrest": "ul", "radif": "li",
	"form": "form", "vorudi": "input", "tasvir": "img", "khat": "hr",
	"jadval": "table", "satr": "tr", "sarnevesht": "th", "khane_jadval": "td",
	"navar": "nav",
}

var colorWords = map[string]string{
	"meshki": "black", "sefid": "white", "sabz": "#16a34a", "ghermez": "#dc2626",
	"abi": "#2563eb", "zard": "#eab308", "narenji": "#f97316", "soorati": "#ec4899",
	"khakestari": "#6b7280", "banafshe": "#7c3aed",
}

var alignWords = map[string]string{
	"rast": "right", "chap": "left", "vasat": "center",
}

func styleByName(name string) (styleWord, bool) {
	for _, w := range styleWords {
		if w.name == name {
			return w, true
		}
	}
	return styleWord{}, false
}

func fnPage(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	title := "Mihan"
	body := args[0]
	if len(args) == 2 {
		s, ok := args[0].(rt.Str)
		if !ok {
			return nil, in.faultAt(node, "E4351",
				"عنوان صفحه باید متن باشد",
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
			"safhe_kon یک جعبه یا یک فهرست از جعبه‌ها می‌خواهد",
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
			"این را نمی‌شود در صفحه گذاشت",
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
			"این گونه از قطعه را نمی‌شناسم: "+kind,
			"this kind of piece is not one I know: %s", kind)
	}
	style, err := cssOf(d)
	if err != nil {
		return in.faultAt(node, "E4350", err.fa, "%s", err.en)
	}
	b.WriteByte('<')
	b.WriteString(tag)
	writeAttr(b, "style", style)
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
		return in.faultAt(node, "E4350",
			"این واژهٔ سبک را نمی‌شناسم: "+name,
			"this style word is not one I know: %s", name)
	}
	return nil
}

type pageFail struct {
	fa string
	en string
}

func cssOf(d *rt.Dict) (string, *pageFail) {
	var b strings.Builder
	for _, w := range styleWords {
		v, ok := d.Get(rt.Str(w.name))
		if !ok || v == nil {
			continue
		}
		if _, none := v.(rt.Hich); none {
			continue
		}
		if w.name == "nemayesh" {
			if s, ok := v.(rt.Str); ok && (string(s) == "radif" || string(s) == "sotun") {
				b.WriteString("display:flex;flex-direction:")
				if string(s) == "sotun" {
					b.WriteString("column")
				} else {
					b.WriteString("row")
				}
				b.WriteByte(';')
				continue
			}
		}
		text, err := cssValue(w, v)
		if err != nil {
			return "", err
		}
		b.WriteString(w.css)
		b.WriteByte(':')
		b.WriteString(text)
		b.WriteByte(';')
	}
	return b.String(), nil
}

func cssValue(w styleWord, v rt.Value) (string, *pageFail) {
	switch t := v.(type) {
	case rt.Int:
		s := strconv.FormatInt(int64(t), 10)
		if w.length {
			s += "px"
		}
		return s, nil
	case rt.Float:
		s := strconv.FormatFloat(float64(t), 'f', -1, 64)
		if w.length {
			s += "px"
		}
		return s, nil
	case rt.Str:
		s := string(t)
		switch w.name {
		case "rang", "pas_zamine", "khat_rang":
			if c, ok := colorWords[s]; ok {
				return c, nil
			}
		case "chid":
			if a, ok := alignWords[s]; ok {
				return a, nil
			}
		}
		return s, nil
	default:
		return "", &pageFail{
			fa: "مقدار این سبک باید متن یا عدد باشد",
			en: "the value of " + w.name + " has to be text or a number, not " + rt.TypeName(v),
		}
	}
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
			"form_bekhan متنِ فرم را می‌خواهد",
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
			"این متنِ یک فرم نیست",
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
