package eval

import (
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
	fa     string
}

var styleWords = []styleWord{
	{"rang", "color", false, "rang-e matn"},
	{"zamine", "background", false, "zamine"},
	{"rang_zamine", "background-color", false, "rang-e zamine"},
	{"pas_zamine", "background", false, "zamine"},
	{"tasvir_zamine", "background-image", false, "aks-e zamine"},
	{"andaze_zamine", "background-size", false, "andaze-ye aks-e zamine"},
	{"jaye_zamine", "background-position", false, "jaye aks-e zamine"},
	{"hashie", "border", true, "hashie"},
	{"khat", "border", true, "hashie"},
	{"rang_hashie", "border-color", false, "rang-e hashie"},
	{"khat_rang", "border-color", false, "rang-e hashie"},
	{"pahnaye_hashie", "border-width", true, "koloftie hashie"},
	{"sabk_hashie", "border-style", false, "sabk-e hashie"},
	{"hashie_bala", "border-top", true, "hashie-ye bala"},
	{"hashie_payin", "border-bottom", true, "hashie-ye payin"},
	{"hashie_chap", "border-left", true, "hashie-ye chap"},
	{"hashie_rast", "border-right", true, "hashie-ye rast"},
	{"gooshe_gerd", "border-radius", true, "gerdie gooshe-ha"},
	{"khat_dor", "border-radius", true, "gerdie gooshe-ha"},
	{"khat_birooni", "outline", true, "khat-e birooni"},
	{"fasele_dakheli", "padding", true, "fasele-ye dakheli"},
	{"fasele", "padding", true, "fasele-ye dakheli"},
	{"fasele_dakheli_bala", "padding-top", true, "fasele-ye dakheli az bala"},
	{"fasele_dakheli_payin", "padding-bottom", true, "fasele-ye dakheli az payin"},
	{"fasele_dakheli_chap", "padding-left", true, "fasele-ye dakheli az chap"},
	{"fasele_dakheli_rast", "padding-right", true, "fasele-ye dakheli az rast"},
	{"fasele_birooni", "margin", true, "fasele-ye birooni"},
	{"fasele_birun", "margin", true, "fasele-ye birooni"},
	{"fasele_birooni_bala", "margin-top", true, "fasele-ye birooni az bala"},
	{"fasele_birooni_payin", "margin-bottom", true, "fasele-ye birooni az payin"},
	{"fasele_birooni_chap", "margin-left", true, "fasele-ye birooni az chap"},
	{"fasele_birooni_rast", "margin-right", true, "fasele-ye birooni az rast"},
	{"fasele_beyn", "gap", true, "fasele-ye beyn-e bache-ha"},
	{"fasele_khane", "gap", true, "fasele-ye beyn-e bache-ha"},
	{"fasele_radif", "row-gap", true, "fasele-ye beyn-e radif-ha"},
	{"fasele_sotun", "column-gap", true, "fasele-ye beyn-e sotun-ha"},
	{"pahna", "width", true, "pahna"},
	{"ertefa", "height", true, "ertefa"},
	{"bala", "height", true, "ertefa"},
	{"kamtarin_pahna", "min-width", true, "kamtarin pahna"},
	{"bishtarin_pahna", "max-width", true, "bishtarin pahna"},
	{"kamtarin_ertefa", "min-height", true, "kamtarin ertefa"},
	{"bishtarin_ertefa", "max-height", true, "bishtarin ertefa"},
	{"andaze_jabe", "box-sizing", false, "hesab-e andaze-ye jabe"},
	{"ghalam", "font-family", false, "ghalam (font)"},
	{"khane", "font-family", false, "ghalam (font)"},
	{"andaze_matn", "font-size", true, "andaze-ye matn"},
	{"andaze", "font-size", true, "andaze-ye matn"},
	{"vazn_matn", "font-weight", false, "vazn (koloftie) matn"},
	{"vazn", "font-weight", false, "vazn (koloftie) matn"},
	{"sabk_matn", "font-style", false, "kaj ya saf"},
	{"tarazbandi", "text-align", false, "tarazbandi-ye matn"},
	{"chid", "text-align", false, "tarazbandi-ye matn"},
	{"ertefa_khat", "line-height", false, "ertefa-ye har khat"},
	{"fasele_harf", "letter-spacing", true, "fasele-ye harf-ha"},
	{"tazin_matn", "text-decoration", false, "zir khat ya roo khat"},
	{"halat_harf", "text-transform", false, "bozorg ya kochak-e harf-ha"},
	{"saye_matn", "text-shadow", false, "saye-ye matn"},
	{"shekastan_kalame", "word-break", false, "shekastan-e kalame-ha"},
	{"faza_khali", "white-space", false, "raftar ba faza-ye khali"},
	{"jahat_matn", "direction", false, "rast be chap ya chap be rast"},
	{"nemayesh", "display", false, "no-e nemayesh"},
	{"jahat", "flex-direction", false, "jahat-e chidan-e bache-ha"},
	{"pichesh", "flex-wrap", false, "raftan be khat-e baad"},
	{"chidman", "justify-content", false, "chidan dar jahat-e asli"},
	{"hamtarazi", "align-items", false, "hamtarazi-ye bache-ha"},
	{"hamtarazi_khod", "align-self", false, "hamtarazi-ye khod-e jabe"},
	{"hamtarazi_mohtava", "align-content", false, "hamtarazi-ye khat-ha"},
	{"felex", "flex", false, "sahm az faza"},
	{"roshd", "flex-grow", false, "cheghadr bozorg beshe"},
	{"kahesh", "flex-shrink", false, "cheghadr kochak beshe"},
	{"paye", "flex-basis", true, "andaze-ye avalie"},
	{"tartib", "order", false, "tartib"},
	{"sotun_ha", "grid-template-columns", false, "sotun-haye shabake"},
	{"radif_ha", "grid-template-rows", false, "radif-haye shabake"},
	{"nahie_ha", "grid-template-areas", false, "nahie-haye shabake"},
	{"nahie", "grid-area", false, "nahie dar shabake"},
	{"sotun_shabake", "grid-column", false, "sotun dar shabake"},
	{"radif_shabake", "grid-row", false, "radif dar shabake"},
	{"jaygozari", "place-items", false, "jaygozari-ye bache-ha"},
	{"mogheiyat", "position", false, "no-e jaygozari"},
	{"az_bala", "top", true, "fasele az bala"},
	{"az_payin", "bottom", true, "fasele az payin"},
	{"az_chap", "left", true, "fasele az chap"},
	{"az_rast", "right", true, "fasele az rast"},
	{"laye", "z-index", false, "laye (che chizi roo ast)"},
	{"sarriz", "overflow", false, "ba chizi ke jash nemishe"},
	{"sarriz_ofoghi", "overflow-x", false, "sarriz-e ofoghi"},
	{"sarriz_amudi", "overflow-y", false, "sarriz-e amudi"},
	{"shenavar", "float", false, "shenavar"},
	{"saye", "box-shadow", false, "saye-ye jabe"},
	{"shafafiat", "opacity", false, "shafafiat (0 ta 1)"},
	{"shafaf", "opacity", false, "shafafiat (0 ta 1)"},
	{"didani", "visibility", false, "dide beshe ya na"},
	{"safi", "filter", false, "safi (tari, roshanai, ...)"},
	{"safi_posht", "backdrop-filter", false, "safi-ye posht-e jabe"},
	{"enteghal", "transition", false, "enteghal-e narm"},
	{"tabdil", "transform", false, "tabdil (charkhesh, bozorgi, ...)"},
	{"mabda_tabdil", "transform-origin", false, "markaz-e tabdil"},
	{"harekat", "animation", false, "harekat (animation)"},
	{"nam_harekat", "animation-name", false, "esm-e harekat"},
	{"modat_harekat", "animation-duration", false, "modat-e harekat"},
	{"takhir_harekat", "animation-delay", false, "takhir-e harekat"},
	{"tekrar_harekat", "animation-iteration-count", false, "chand bar tekrar"},
	{"nashangar", "cursor", false, "shekl-e nashangar-e mous"},
	{"lams", "pointer-events", false, "javab be mous"},
	{"entekhab", "user-select", false, "mishe matn ro entekhab kard?"},
	{"jaygiri_tasvir", "object-fit", false, "aks chetor por kone"},
	{"nesbat", "aspect-ratio", false, "nesbat-e pahna be ertefa"},
	{"neshane_fehrest", "list-style", false, "neshane-ye fehrest"},
	{"mohtava", "content", false, "mohtava (baraye ::ghabl va ::baad)"},
}

var cssLength = map[string]bool{
	"width": true, "height": true, "min-width": true, "max-width": true, "min-height": true, "max-height": true,
	"padding": true, "padding-top": true, "padding-bottom": true, "padding-left": true, "padding-right": true,
	"margin": true, "margin-top": true, "margin-bottom": true, "margin-left": true, "margin-right": true,
	"border": true, "border-top": true, "border-bottom": true, "border-left": true, "border-right": true,
	"border-width": true, "border-radius": true, "outline": true, "outline-offset": true, "font-size": true,
	"gap": true, "row-gap": true, "column-gap": true, "top": true, "bottom": true, "left": true, "right": true,
	"letter-spacing": true, "word-spacing": true, "text-indent": true, "flex-basis": true, "inset": true,
}

var cssNames = map[string]bool{
	"accent-color": true, "all": true, "animation-direction": true, "animation-fill-mode": true,
	"animation-play-state": true, "animation-timing-function": true, "backface-visibility": true,
	"background-attachment": true, "background-clip": true, "background-repeat": true,
	"border-collapse": true, "border-spacing": true, "border-top-left-radius": true,
	"border-top-right-radius": true, "border-bottom-left-radius": true, "border-bottom-right-radius": true,
	"caret-color": true, "clip-path": true, "color-scheme": true, "columns": true, "counter-increment": true,
	"counter-reset": true, "empty-cells": true, "fill": true, "flex-flow": true, "font": true,
	"font-variant": true, "grid": true, "grid-auto-flow": true, "grid-auto-columns": true, "grid-auto-rows": true,
	"grid-template": true, "hyphens": true, "isolation": true, "justify-items": true, "justify-self": true,
	"list-style-type": true, "list-style-position": true, "mask": true, "mix-blend-mode": true,
	"object-position": true, "outline-color": true, "outline-style": true, "outline-width": true,
	"overflow-wrap": true, "perspective": true, "place-content": true, "place-self": true, "quotes": true,
	"resize": true, "scroll-behavior": true, "scroll-snap-type": true, "scroll-snap-align": true,
	"stroke": true, "stroke-width": true, "table-layout": true, "text-overflow": true,
	"text-decoration-color": true, "text-underline-offset": true, "transition-delay": true,
	"transition-duration": true, "transition-property": true, "transition-timing-function": true,
	"vertical-align": true, "will-change": true, "word-wrap": true, "writing-mode": true,
	"text-wrap": true, "inset": true, "word-spacing": true, "text-indent": true, "outline-offset": true,
}

var colorWords = map[string]string{
	"siah": "black", "meshki": "black", "sefid": "white", "ghermez": "#dc2626", "abi": "#2563eb",
	"sabz": "#16a34a", "zard": "#eab308", "narenji": "#f97316", "banafsh": "#7c3aed", "banafshe": "#7c3aed",
	"soorati": "#ec4899", "khakestari": "#6b7280", "ghahvei": "#92400e", "firoozei": "#06b6d4",
	"sormei": "#1e3a8a", "talayi": "#d4af37", "noghrei": "#c0c0c0", "zereshki": "#9f1239",
	"abi_aseman": "#38bdf8", "sabz_tire": "#14532d", "ghermez_tire": "#7f1d1d", "kerem": "#fef3c7",
	"shafaf": "transparent", "hamrang": "currentColor",
}

var valueWords = map[string]string{
	"markaz": "center", "vasat": "center", "chap": "left", "rast": "right", "bala": "top", "payin": "bottom",
	"por": "100%", "hich": "none", "makhfi": "hidden", "peida": "visible", "khodkar": "auto", "adi": "normal",
	"porrang": "bold", "zakhim": "bold", "nazok": "300", "kaj": "italic", "zir_khat": "underline",
	"roo_khat": "line-through", "boluk": "block", "khati": "inline", "boluk_khati": "inline-block",
	"felex": "flex", "shabake": "grid", "radif": "row", "sotun": "column", "radif_baraks": "row-reverse",
	"sotun_baraks": "column-reverse", "pichide": "wrap", "bi_pich": "nowrap", "shoru": "flex-start",
	"payan": "flex-end", "faseledar": "space-between", "dor_ta_dor": "space-around", "mosavi": "space-evenly",
	"keshide": "stretch", "motlagh": "absolute", "nesbi": "relative", "sabet": "fixed", "chasban": "sticky",
	"istade": "static", "dast": "pointer", "pishfarz": "default", "mamnoo": "not-allowed", "entezar": "wait",
	"neshangar_matn": "text", "jabeja": "move", "narm": "ease", "narm_vorud": "ease-in", "narm_khoruj": "ease-out",
	"narm_dotarafe": "ease-in-out", "yeknavakht": "linear", "bi_payan": "infinite", "raft_bargasht": "alternate",
	"baraks": "reverse", "har_do": "both", "jolo": "forwards", "por_kon": "cover", "jash_bede": "contain",
	"bozorg_harf": "uppercase", "kochak_harf": "lowercase", "rast_be_chap": "rtl", "chap_be_rast": "ltr",
	"saf": "solid", "khat_chin": "dashed", "noghte_chin": "dotted", "dotayi": "double", "lagzan": "scroll",
	"hame": "all", "dayere": "circle", "noghte": "disc", "mojood": "border-box", "ers": "inherit",
}

var funcWords = map[string]string{
	"charkhesh": "rotate", "bozorgi": "scale", "jabejayi": "translate", "jabejayi_x": "translateX",
	"jabejayi_y": "translateY", "kaji": "skew", "tari": "blur", "roshanai": "brightness",
	"khakestari": "grayscale", "shib": "linear-gradient", "shib_dayere": "radial-gradient",
	"tekrar": "repeat", "mohasebe": "calc", "kamtarin": "min", "bishtarin": "max", "mahdood": "clamp",
	"motaghayer": "var", "aks": "url", "saye": "drop-shadow", "ashba": "saturate",
}

var pseudoWords = map[string]string{
	"rooye": "hover", "fokus": "focus", "faal": "active", "aval": "first-child", "akhar": "last-child",
	"ghabl": "before", "baad": "after", "dide_shode": "visited", "gheyr_faal": "disabled", "zoj": "nth-child(even)",
	"fard": "nth-child(odd)",
}

var selectorTags = map[string]string{"badane": "body", "safhe": "html", "hame": "*"}

var styleByNameMap = func() map[string]styleWord {
	m := map[string]styleWord{}
	for _, w := range styleWords {
		m[w.name] = w
	}
	return m
}()

func styleByName(name string) (styleWord, bool) {
	if w, ok := styleByNameMap[name]; ok {
		return w, true
	}
	if strings.HasPrefix(name, "--") {
		return styleWord{name: name, css: name}, true
	}
	css := strings.ReplaceAll(name, "_", "-")
	if cssNames[css] || cssLength[css] {
		return styleWord{name: name, css: css, length: cssLength[css]}, true
	}
	for _, w := range styleWords {
		if w.css == css {
			return styleWord{name: name, css: css, length: w.length}, true
		}
	}
	return styleWord{}, false
}

func StyleWordList() [][4]string {
	out := make([][4]string, 0, len(styleWords))
	for _, w := range styleWords {
		out = append(out, [4]string{w.name, w.css, w.fa, ""})
	}
	return out
}

func StyleValueList() [][2]string {
	var out [][2]string
	for k, v := range colorWords {
		out = append(out, [2]string{k, v})
	}
	for k, v := range valueWords {
		out = append(out, [2]string{k, v})
	}
	for k, v := range funcWords {
		out = append(out, [2]string{k + "()", v + "()"})
	}
	sort.Slice(out, func(i, j int) bool { return out[i][0] < out[j][0] })
	return out
}

func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentByte(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9') || c == '-'
}

func translateCSSValue(s string, props bool) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		c := s[i]
		if c == '"' || c == '\'' {
			j := strings.IndexByte(s[i+1:], c)
			if j < 0 {
				b.WriteString(s[i:])
				break
			}
			b.WriteString(s[i : i+j+2])
			i += j + 2
			continue
		}
		if !isIdentStart(c) {
			b.WriteByte(c)
			i++
			continue
		}
		j := i
		for j < len(s) && isIdentByte(s[j]) {
			j++
		}
		word := s[i:j]
		prev := byte(' ')
		if i > 0 {
			prev = s[i-1]
		}
		if prev == '#' || prev == '.' || prev == '-' || (prev >= '0' && prev <= '9') {
			b.WriteString(word)
			i = j
			continue
		}
		if j < len(s) && s[j] == '(' {
			name := word
			if f, ok := funcWords[word]; ok {
				name = f
			}
			b.WriteString(name)
			if name == "url" {
				k := strings.IndexByte(s[j:], ')')
				if k >= 0 {
					b.WriteString(s[j : j+k+1])
					i = j + k + 1
					continue
				}
			}
			i = j
			continue
		}
		if props {
			if w, ok := styleByNameMap[word]; ok {
				b.WriteString(w.css)
				i = j
				continue
			}
		}
		if v, ok := colorWords[word]; ok {
			b.WriteString(v)
		} else if v, ok := valueWords[word]; ok {
			b.WriteString(v)
		} else {
			b.WriteString(word)
		}
		i = j
	}
	return b.String()
}

func cssValue(w styleWord, v rt.Value) (string, *pageFail) {
	switch t := v.(type) {
	case rt.Int:
		s := strconv.FormatInt(int64(t), 10)
		if w.length && t != 0 {
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
		props := w.css == "transition" || w.css == "transition-property" || w.css == "will-change"
		return translateCSSValue(string(t), props), nil
	case *rt.List:
		parts := make([]string, 0, len(t.Items))
		sep := " "
		if w.css == "transition" || w.css == "animation" || w.css == "font-family" || w.css == "box-shadow" || w.css == "background-image" || w.css == "text-shadow" {
			sep = ", "
		}
		for _, item := range t.Items {
			p, err := cssValue(w, item)
			if err != nil {
				return "", err
			}
			parts = append(parts, p)
		}
		return strings.Join(parts, sep), nil
	default:
		msg := "meghdar-e '" + w.name + "' bayad matn, adad ya list bashe, na " + rt.TypeName(v)
		return "", &pageFail{fa: msg, en: msg}
	}
}

func cssDecl(name string, v rt.Value) (string, *pageFail) {
	w, ok := styleByName(name)
	if !ok {
		msg := unknownStyleMsg(name)
		return "", &pageFail{fa: msg, en: msg}
	}
	if _, none := v.(rt.Hich); none || v == nil {
		return "", nil
	}
	if w.css == "display" {
		if s, ok := v.(rt.Str); ok && (string(s) == "radif" || string(s) == "sotun") {
			dir := "row"
			if string(s) == "sotun" {
				dir = "column"
			}
			return "display:flex;flex-direction:" + dir + ";", nil
		}
	}
	text, err := cssValue(w, v)
	if err != nil {
		return "", err
	}
	return w.css + ":" + text + ";", nil
}

func unknownStyleMsg(name string) string {
	msg := "in kalame-ye sabk ro nemishnasam: " + name
	best, bestD := "", 3
	for _, w := range styleWords {
		if d := editDistance(name, w.name); d < bestD {
			best, bestD = w.name, d
		}
	}
	if best != "" {
		msg += "; shayad manzooret '" + best + "' bood?"
	}
	return msg
}

func editDistance(a, b string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}

func cssOf(d *rt.Dict) (string, *pageFail) {
	var b strings.Builder
	for _, k := range d.Keys() {
		name, ok := jsonKey(k)
		if !ok || structureKeys[name] {
			continue
		}
		v, _ := d.Get(k)
		text, err := cssDecl(name, v)
		if err != nil {
			return "", err
		}
		b.WriteString(text)
	}
	return b.String(), nil
}

func translateSelector(sel string) string {
	var b strings.Builder
	i := 0
	for i < len(sel) {
		c := sel[i]
		if c == '[' {
			j := strings.IndexByte(sel[i:], ']')
			if j < 0 {
				b.WriteString(sel[i:])
				break
			}
			b.WriteString(sel[i : i+j+1])
			i += j + 1
			continue
		}
		if !isIdentStart(c) {
			b.WriteByte(c)
			i++
			continue
		}
		j := i
		for j < len(sel) && isIdentByte(sel[j]) {
			j++
		}
		word := sel[i:j]
		prev := byte(' ')
		if i > 0 {
			prev = sel[i-1]
		}
		switch {
		case prev == ':':
			if p, ok := pseudoWords[word]; ok {
				word = p
			}
		case prev == '.' || prev == '#' || prev == '-':
		default:
			if t, ok := goneTag[word]; ok {
				word = t
			} else if t, ok := selectorTags[word]; ok {
				word = t
			}
		}
		b.WriteString(word)
		i = j
	}
	return b.String()
}

func atRule(key string) (string, bool) {
	f := strings.Fields(key)
	if len(f) == 0 {
		return key, false
	}
	switch f[0] {
	case "@harekat", "@keyframes":
		if len(f) > 1 {
			return "@keyframes " + f[1], true
		}
	case "@kochak_tar", "@bozorg_tar":
		if len(f) > 1 {
			n := f[1]
			if _, err := strconv.Atoi(n); err == nil {
				n += "px"
			}
			if f[0] == "@kochak_tar" {
				return "@media (max-width:" + n + ")", false
			}
			return "@media (min-width:" + n + ")", false
		}
	case "@kam_harekat":
		return "@media (prefers-reduced-motion:reduce)", false
	case "@tarik":
		return "@media (prefers-color-scheme:dark)", false
	case "@roshan":
		return "@media (prefers-color-scheme:light)", false
	}
	return key, false
}

var frameWords = map[string]string{"az": "from", "ta": "to", "shoru": "from", "payan": "to"}

func cssRules(b *strings.Builder, d *rt.Dict, frames bool) *pageFail {
	for _, k := range d.Keys() {
		key, ok := jsonKey(k)
		if !ok {
			continue
		}
		v, _ := d.Get(k)
		if s, isStr := v.(rt.Str); isStr && strings.HasPrefix(key, "@") {
			b.WriteString(key + " " + string(s) + ";\n")
			continue
		}
		inner, ok := v.(*rt.Dict)
		if !ok {
			msg := "dar sabk, '" + key + "' bayad yek dict az kalame-haye sabk dashte bashe"
			if _, known := styleByName(key); known {
				msg = "'" + key + "' yek kalame-ye sabk ast; bayad dakhel-e yek entekhabgar bashe, mesl-e {\"onvan\": {\"" + key + "\": ...}}"
			}
			return &pageFail{fa: msg, en: msg}
		}
		if strings.HasPrefix(key, "@") {
			head, isFrames := atRule(key)
			b.WriteString(head + " {\n")
			if err := cssRules(b, inner, isFrames); err != nil {
				return err
			}
			b.WriteString("}\n")
			continue
		}
		sel := key
		if frames {
			parts := strings.Split(key, ",")
			for i, p := range parts {
				p = strings.TrimSpace(p)
				if f, ok := frameWords[p]; ok {
					p = f
				}
				parts[i] = p
			}
			sel = strings.Join(parts, ",")
		} else {
			sel = translateSelector(key)
		}
		body, err := cssOf(inner)
		if err != nil {
			return err
		}
		b.WriteString(sel + " {" + body + "}\n")
	}
	return nil
}

func styleSheet(v rt.Value) (string, *pageFail) {
	switch t := v.(type) {
	case rt.Str:
		return string(t), nil
	case *rt.Dict:
		var b strings.Builder
		if err := cssRules(&b, t, false); err != nil {
			return "", err
		}
		return b.String(), nil
	case *rt.List:
		var b strings.Builder
		for _, item := range t.Items {
			s, err := styleSheet(item)
			if err != nil {
				return "", err
			}
			b.WriteString(s)
		}
		return b.String(), nil
	default:
		msg := "sabk yek dict az entekhabgar-ha mikhad, na " + rt.TypeName(v)
		return "", &pageFail{fa: msg, en: msg}
	}
}

func fnStyle(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	s, err := styleSheet(args[0])
	if err != nil {
		return nil, in.faultAt(node, "E4355", err.fa, "%s", err.en)
	}
	return rt.Str(s), nil
}

func fnStyleInline(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	d, ok := args[0].(*rt.Dict)
	if !ok {
		return nil, in.faultAt(node, "E4355", "sabk_khat yek dict az kalame-haye sabk mikhad", "sabk_khat yek dict az kalame-haye sabk mikhad")
	}
	s, err := cssOf(d)
	if err != nil {
		return nil, in.faultAt(node, "E4350", err.fa, "%s", err.en)
	}
	return rt.Str(s), nil
}
