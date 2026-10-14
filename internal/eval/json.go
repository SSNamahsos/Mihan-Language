package eval

import (
	"encoding/json"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

func fnJSONRead(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	s, ok := args[0].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4320",
			"json_bekhan matn mikhad",
			"json_bekhan needs text, not %s", rt.TypeName(args[0]))
	}
	text := strings.TrimPrefix(string(s), "\uFEFF")
	text = strings.TrimSpace(text)

	if v, ok := fastJSON(text); ok {
		return v, nil
	}
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()
	v, err := readJSON(dec)
	if err != nil {
		return nil, in.faultAt(node, "E4321",
			"in matn JSON nist",
			"this text is not JSON: %v", err)
	}
	if dec.More() {
		return nil, in.faultAt(node, "E4321",
			"in matn bish az yek meghdar-e JSON dare",
			"this text has more than one JSON value")
	}
	return v, nil
}

func fnJSONWrite(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	var b strings.Builder
	w := &jsonWriter{}
	if err := w.write(&b, args[0]); err != nil {
		return nil, in.faultAt(node, "E4322",
			err.fa,
			"%s", err.en)
	}
	return rt.Str(b.String()), nil
}

type jsonFail struct {
	fa string
	en string
}

func (e *jsonFail) Error() string { return e.en }

func readJSON(dec *json.Decoder) (rt.Value, error) {
	tok, err := dec.Token()
	if err != nil {
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			d := rt.NewDict()
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyTok.(string)
				if !ok {
					return nil, io.ErrUnexpectedEOF
				}
				val, err := readJSON(dec)
				if err != nil {
					return nil, err
				}
				d.Set(rt.Str(key), val)
			}
			if _, err := dec.Token(); err != nil {
				return nil, err
			}
			return d, nil
		case '[':
			items := []rt.Value{}
			for dec.More() {
				val, err := readJSON(dec)
				if err != nil {
					return nil, err
				}
				items = append(items, val)
			}
			if _, err := dec.Token(); err != nil {
				return nil, err
			}
			return rt.NewList(items...), nil
		default:
			return nil, io.ErrUnexpectedEOF
		}
	case string:
		return rt.Str(t), nil
	case bool:
		return rt.Bool(t), nil
	case nil:
		return rt.Nil, nil
	case json.Number:
		return jsonNumber(t)
	default:
		return nil, io.ErrUnexpectedEOF
	}
}

func jsonNumber(n json.Number) (rt.Value, error) {
	s := string(n)
	if strings.ContainsAny(s, ".eE") {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		return rt.Float(f), nil
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return rt.Int(i), nil
	}
	f, ferr := strconv.ParseFloat(s, 64)
	if ferr != nil {
		return nil, err
	}
	return rt.Float(f), nil
}

type jsonWriter struct {
	lists map[*rt.List]bool
	dicts map[*rt.Dict]bool
}

func (w *jsonWriter) write(b *strings.Builder, v rt.Value) *jsonFail {
	switch t := v.(type) {
	case nil:
		b.WriteString("null")
	case rt.Hich:
		b.WriteString("null")
	case rt.Bool:
		if t {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case rt.Int:
		b.WriteString(strconv.FormatInt(int64(t), 10))
	case rt.Float:
		f := float64(t)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return &jsonFail{
				fa: "JSON adad-e bi-payan ya bi-mani nadare",
				en: "JSON has no infinity and no not-a-number",
			}
		}
		s := strconv.FormatFloat(f, 'g', -1, 64)
		if !strings.ContainsAny(s, ".eE") {
			s += ".0"
		}
		b.WriteString(s)
	case rt.Str:
		raw, err := json.Marshal(string(t))
		if err != nil {
			return &jsonFail{fa: "in matn JSON nemishe", en: "this text cannot be JSON"}
		}
		b.Write(raw)
	case *rt.List:
		if w.lists == nil {
			w.lists = map[*rt.List]bool{}
		}
		if w.lists[t] {
			return &jsonFail{
				fa: "in meghdar khodesh ro too khodesh dare va JSON tamoom nemishe",
				en: "this value contains itself, so it cannot be JSON",
			}
		}
		w.lists[t] = true
		b.WriteByte('[')
		for i, item := range t.Items {
			if i > 0 {
				b.WriteByte(',')
			}
			if err := w.write(b, item); err != nil {
				return err
			}
		}
		b.WriteByte(']')
		delete(w.lists, t)
	case *rt.Dict:
		if w.dicts == nil {
			w.dicts = map[*rt.Dict]bool{}
		}
		if w.dicts[t] {
			return &jsonFail{
				fa: "in meghdar khodesh ro too khodesh dare va JSON tamoom nemishe",
				en: "this value contains itself, so it cannot be JSON",
			}
		}
		w.dicts[t] = true
		b.WriteByte('{')
		for i, k := range t.Keys() {
			key, ok := jsonKey(k)
			if !ok {
				return &jsonFail{
					fa: "kelid-e JSON faghat mitoone matn bashe",
					en: "a JSON object's keys have to be text, not " + rt.TypeName(k),
				}
			}
			if i > 0 {
				b.WriteByte(',')
			}
			raw, err := json.Marshal(key)
			if err != nil {
				return &jsonFail{fa: "in matn JSON nemishe", en: "this text cannot be JSON"}
			}
			b.Write(raw)
			b.WriteByte(':')
			val, _ := t.Get(k)
			if err := w.write(b, val); err != nil {
				return err
			}
		}
		b.WriteByte('}')
		delete(w.dicts, t)
	default:
		return &jsonFail{
			fa: "in meghdar ro nemishe JSON kard",
			en: "a " + rt.TypeName(v) + " cannot be written as JSON",
		}
	}
	return nil
}

func jsonKey(k rt.Value) (string, bool) {
	switch t := k.(type) {
	case rt.Str:
		return string(t), true
	case string:
		return t, true
	default:
		return "", false
	}
}

func fastJSON(text string) (rt.Value, bool) {
	p := &jsonReader{s: text}
	v, ok := p.value(0)
	if !ok {
		return nil, false
	}
	p.space()
	return v, p.i == len(p.s)
}

type jsonReader struct {
	s string
	i int
}

func (p *jsonReader) space() {
	for p.i < len(p.s) {
		switch p.s[p.i] {
		case ' ', '\t', '\n', '\r':
			p.i++
		default:
			return
		}
	}
}

func (p *jsonReader) value(depth int) (rt.Value, bool) {
	if depth > 10000 {
		return nil, false
	}
	p.space()
	if p.i >= len(p.s) {
		return nil, false
	}
	switch c := p.s[p.i]; {
	case c == '{':
		p.i++
		d := rt.NewDict()
		p.space()
		if p.i < len(p.s) && p.s[p.i] == '}' {
			p.i++
			return d, true
		}
		for {
			p.space()
			if p.i >= len(p.s) || p.s[p.i] != '"' {
				return nil, false
			}
			k, ok := p.str()
			if !ok {
				return nil, false
			}
			p.space()
			if p.i >= len(p.s) || p.s[p.i] != ':' {
				return nil, false
			}
			p.i++
			v, ok := p.value(depth + 1)
			if !ok {
				return nil, false
			}
			d.Set(rt.Str(k), v)
			p.space()
			if p.i >= len(p.s) {
				return nil, false
			}
			if p.s[p.i] == ',' {
				p.i++
				continue
			}
			if p.s[p.i] == '}' {
				p.i++
				return d, true
			}
			return nil, false
		}
	case c == '[':
		p.i++
		items := []rt.Value{}
		p.space()
		if p.i < len(p.s) && p.s[p.i] == ']' {
			p.i++
			return rt.NewList(items...), true
		}
		for {
			v, ok := p.value(depth + 1)
			if !ok {
				return nil, false
			}
			items = append(items, v)
			p.space()
			if p.i >= len(p.s) {
				return nil, false
			}
			if p.s[p.i] == ',' {
				p.i++
				continue
			}
			if p.s[p.i] == ']' {
				p.i++
				return rt.NewList(items...), true
			}
			return nil, false
		}
	case c == '"':
		s, ok := p.str()
		if !ok {
			return nil, false
		}
		return rt.Str(s), true
	case c == 't':
		if strings.HasPrefix(p.s[p.i:], "true") {
			p.i += 4
			return rt.Bool(true), true
		}
	case c == 'f':
		if strings.HasPrefix(p.s[p.i:], "false") {
			p.i += 5
			return rt.Bool(false), true
		}
	case c == 'n':
		if strings.HasPrefix(p.s[p.i:], "null") {
			p.i += 4
			return rt.Nil, true
		}
	case c == '-' || (c >= '0' && c <= '9'):
		return p.num()
	}
	return nil, false
}

func (p *jsonReader) num() (rt.Value, bool) {
	start := p.i
	if p.s[p.i] == '-' {
		p.i++
	}
	digits := func() int {
		n := 0
		for p.i < len(p.s) && p.s[p.i] >= '0' && p.s[p.i] <= '9' {
			p.i++
			n++
		}
		return n
	}
	if p.i < len(p.s) && p.s[p.i] == '0' {
		p.i++
	} else if digits() == 0 {
		return nil, false
	}
	if p.i < len(p.s) && p.s[p.i] == '.' {
		p.i++
		if digits() == 0 {
			return nil, false
		}
	}
	if p.i < len(p.s) && (p.s[p.i] == 'e' || p.s[p.i] == 'E') {
		p.i++
		if p.i < len(p.s) && (p.s[p.i] == '+' || p.s[p.i] == '-') {
			p.i++
		}
		if digits() == 0 {
			return nil, false
		}
	}
	v, err := jsonNumber(json.Number(p.s[start:p.i]))
	if err != nil {
		return nil, false
	}
	return v, true
}

func (p *jsonReader) str() (string, bool) {
	p.i++
	start := p.i
	for p.i < len(p.s) {
		c := p.s[p.i]
		if c == '"' {
			s := p.s[start:p.i]
			p.i++
			if !utf8.ValidString(s) {
				return "", false
			}
			return s, true
		}
		if c == '\\' || c < 0x20 {
			break
		}
		p.i++
	}

	p.i = start - 1
	end := p.i + 1
	for end < len(p.s) {
		c := p.s[end]
		if c == '\\' {
			end += 2
			continue
		}
		if c == '"' {
			break
		}
		if c < 0x20 {
			return "", false
		}
		end++
	}
	if end >= len(p.s) {
		return "", false
	}
	var out string
	if err := json.Unmarshal([]byte(p.s[p.i:end+1]), &out); err != nil {
		return "", false
	}
	if !utf8.ValidString(p.s[p.i : end+1]) {
		return "", false
	}
	p.i = end + 1
	return out, true
}
