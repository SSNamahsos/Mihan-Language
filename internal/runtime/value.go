package runtime

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"mihan/internal/ast"
)

type Value any

type (
	Int int64

	Float float64

	Str string

	Bool bool

	Hich struct{}

	Range struct {
		Lo, Hi    Int
		Inclusive bool
	}

	StepRange struct {
		Lo, Hi, Step Int
	}
)

var Nil = Hich{}

type List struct{ Items []Value }

func NewList(items ...Value) *List { return &List{Items: items} }

type Dict struct {
	keys []Value
	vals []Value

	sIdx  map[string]int
	iIdx  map[int64]int
	index map[string]int
}

func NewDict() *Dict { return &Dict{} }

func (d *Dict) Len() int { return len(d.keys) }

func (d *Dict) Keys() []Value { return d.keys }

func dkey(k Value) (kind byte, s string, i int64) {
	switch t := k.(type) {
	case Str:
		return 's', string(t), 0
	case string:
		return 's', t, 0
	case Int:
		return 'i', "", int64(t)
	case Float:
		if float64(t) == math.Trunc(float64(t)) && math.Abs(float64(t)) < 1e18 {
			return 'i', "", int64(t)
		}
	}
	h, ok := HashKey(k)
	if !ok {
		return 0, "", 0
	}
	return 'h', h, 0
}

func (d *Dict) find(kind byte, s string, i int64) (int, bool) {
	var at int
	var ok bool
	switch kind {
	case 's':
		at, ok = d.sIdx[s]
	case 'i':
		at, ok = d.iIdx[i]
	case 'h':
		at, ok = d.index[s]
	}
	return at, ok
}

func (d *Dict) put(kind byte, s string, i int64, at int) {
	switch kind {
	case 's':
		if d.sIdx == nil {
			d.sIdx = map[string]int{}
		}
		d.sIdx[s] = at
	case 'i':
		if d.iIdx == nil {
			d.iIdx = map[int64]int{}
		}
		d.iIdx[i] = at
	case 'h':
		if d.index == nil {
			d.index = map[string]int{}
		}
		d.index[s] = at
	}
}

func (d *Dict) Get(k Value) (Value, bool) {
	kind, s, i := dkey(k)
	if kind == 0 {
		return nil, false
	}
	at, ok := d.find(kind, s, i)
	if !ok {
		return nil, false
	}
	return d.vals[at], true
}

func (d *Dict) Set(k, v Value) {
	kind, s, i := dkey(k)
	if kind == 0 {
		return
	}
	if at, found := d.find(kind, s, i); found {
		d.vals[at] = v
		return
	}
	d.put(kind, s, i, len(d.keys))
	d.keys = append(d.keys, k)
	d.vals = append(d.vals, v)
}

func (d *Dict) Delete(k Value) bool {
	kind, s, i := dkey(k)
	if kind == 0 {
		return false
	}
	at, found := d.find(kind, s, i)
	if !found {
		return false
	}
	d.keys = append(d.keys[:at], d.keys[at+1:]...)
	d.vals = append(d.vals[:at], d.vals[at+1:]...)
	switch kind {
	case 's':
		delete(d.sIdx, s)
	case 'i':
		delete(d.iIdx, i)
	case 'h':
		delete(d.index, s)
	}
	for j := at; j < len(d.keys); j++ {
		kk, ks, ki := dkey(d.keys[j])
		d.put(kk, ks, ki, j)
	}
	return true
}

type Set struct {
	order []Value
	index map[string]bool
}

func NewSet() *Set { return &Set{index: map[string]bool{}} }

func (s *Set) Len() int { return len(s.order) }

func (s *Set) Items() []Value { return s.order }

func (s *Set) Add(v Value) {
	h, ok := HashKey(v)
	if !ok {
		return
	}
	if s.index[h] {
		return
	}
	s.index[h] = true
	s.order = append(s.order, v)
}

func (s *Set) Has(v Value) bool {
	h, ok := HashKey(v)
	return ok && s.index[h]
}

func (s *Set) Delete(v Value) bool {
	h, ok := HashKey(v)
	if !ok || !s.index[h] {
		return false
	}
	delete(s.index, h)
	for i, item := range s.order {
		if ih, _ := HashKey(item); ih == h {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return true
}

type Class struct {
	Name    string
	Doc     string
	Parent  *Class
	Fields  []*ast.FieldDecl
	Methods map[string]*ast.FuncDecl

	Order []string

	Defaults map[string]Value

	Enumerated bool
}

func (c *Class) FindMethod(name string) (*ast.FuncDecl, *Class) {
	for k := c; k != nil; k = k.Parent {
		if m, ok := k.Methods[name]; ok {
			return m, k
		}
	}
	return nil, nil
}

func (c *Class) FindField(name string) (*ast.FieldDecl, bool) {
	for k := c; k != nil; k = k.Parent {
		for _, f := range k.Fields {
			if f.Name == name {
				return f, true
			}
		}
	}
	return nil, false
}

func (c *Class) IsA(d *Class) bool {
	for k := c; k != nil; k = k.Parent {
		if k == d {
			return true
		}
	}
	return false
}

type Object struct {
	Class  *Class
	Fields map[string]Value
}

func NewObject(c *Class) *Object {
	o := &Object{Class: c, Fields: make(map[string]Value, len(c.Fields))}

	var chain []*Class
	for k := c; k != nil; k = k.Parent {
		chain = append(chain, k)
	}
	for i := len(chain) - 1; i >= 0; i-- {
		k := chain[i]
		for _, f := range k.Fields {
			if v, ok := k.Defaults[HashKeyMust(f.Name)]; ok {
				o.Fields[f.Name] = freshCopy(v)
			}
		}
	}
	return o
}

type ClassValue struct {
	Class *Class
}

func (c *ClassValue) Value() Value {
	if c.Class.Enumerated {
		return NewObject(c.Class)
	}
	return Nil
}

func (o *Object) Get(name string) (Value, bool) {
	if v, ok := o.Fields[name]; ok {
		return v, true
	}
	for k := o.Class; k != nil; k = k.Parent {
		if _, ok := k.FindField(name); ok {
			return Nil, true
		}
	}
	return nil, false
}

func (o *Object) Set(name string, v Value) { o.Fields[name] = v }

type Function struct {
	Name   string
	Params []*ast.Param
	Body   *ast.Block

	Expr  ast.Expr
	Env   *Env
	Class *Class

	Native NativeFn

	Result string

	BuiltinName string

	Doc string

	Fast any
}

type NativeFn func(host any, args []Value, node ast.Node) (Value, error)

type BoundMethod struct {
	Recv Value
	Fn   *Function

	Owner *Class
}

type ErrorValue struct {
	Message Str
	Data    Value

	Name string

	Note string

	FA string
}

func (e *ErrorValue) Error() string { return string(e.Message) }

func NewError(format string, args ...any) *ErrorValue {
	return &ErrorValue{Message: Str(fmt.Sprintf(format, args...)), Data: Nil}
}

func TypeName(v Value) string {
	switch t := v.(type) {
	case Int:
		return "arz"
	case Float:
		return "kasr"
	case Str:
		return "matn"
	case Bool:
		return "bool"
	case Hich:
		return "hich"
	case *List:
		return "list"
	case *Dict:
		return "dict"
	case *Set:
		return "set"
	case *Object:
		return t.Class.Name
	case *Function:
		return "funct"
	case *BoundMethod:
		return "funct"
	case *Range, *StepRange:
		return "rang"
	case *ErrorValue:
		return "khata"
	case *ClassValue:
		return t.Class.Name
	case NamedArg:

		return TypeName(t.value)
	}
	return "hich"
}

func IsNumber(v Value) bool {
	switch v.(type) {
	case Int, Float, Bool:
		return true
	}
	return false
}

func Truthy(v Value) bool {
	switch t := v.(type) {
	case Bool:
		return bool(t)
	case Int:
		return t != 0
	case Float:
		return t != 0
	case Str:
		return t != ""
	case Hich:
		return false
	case *List:
		return len(t.Items) > 0
	case *Dict:
		return t.Len() > 0
	case *Set:
		return t.Len() > 0
	case *Range:
		return true
	}
	return true
}

func Show(v Value) string {
	var b strings.Builder
	write(&b, v, false)
	return b.String()
}

func Repr(v Value) string {
	var b strings.Builder
	write(&b, v, true)
	return b.String()
}

func write(b *strings.Builder, v Value, quoted bool) {
	switch t := v.(type) {
	case nil:
		b.WriteString("hich")
	case Hich:
		b.WriteString("hich")
	case Int:
		b.WriteString(strconv.FormatInt(int64(t), 10))
	case Bool:

		if t {
			b.WriteString("dorost")
		} else {
			b.WriteString("ghalat")
		}
	case Float:
		b.WriteString(formatFloat(float64(t)))
	case Str:
		if quoted {
			b.WriteString(Quote(string(t)))
		} else {
			b.WriteString(string(t))
		}
	case string:

		if quoted {
			b.WriteString(Quote(t))
		} else {
			b.WriteString(t)
		}
	case *List:

		b.WriteByte('[')
		for i, item := range t.Items {
			if i > 0 {
				b.WriteString(", ")
			}
			write(b, item, true)
		}
		b.WriteByte(']')
	case *Dict:
		b.WriteByte('{')
		for i, k := range t.Keys() {
			if i > 0 {
				b.WriteString(", ")
			}
			write(b, k, true)
			b.WriteString(": ")
			v, _ := t.Get(k)
			write(b, v, true)
		}
		b.WriteByte('}')
	case *Set:
		b.WriteByte('{')
		for i, item := range t.Items() {
			if i > 0 {
				b.WriteString(", ")
			}
			write(b, item, true)
		}
		b.WriteByte('}')
	case *Object:
		b.WriteString(t.Class.Name)
		b.WriteByte('{')
		names := t.FieldNames()
		for i, n := range names {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(n)
			b.WriteString(": ")
			write(b, t.Fields[n], true)
		}
		b.WriteByte('}')
	case *Range:
		b.WriteString(showInt(t.Lo))
		if t.Inclusive {
			b.WriteString("..+")
		} else {
			b.WriteString("..")
		}
		b.WriteString(showInt(t.Hi))
	case *StepRange:

		b.WriteString(showInt(t.Lo))
		b.WriteString("..")
		b.WriteString(showInt(t.Step))
		b.WriteString("..")
		b.WriteString(showInt(t.Hi))
	case *ErrorValue:
		b.WriteString("khata(")
		write(b, t.Message, true)
		if !isNil(t.Data) {
			b.WriteString(", ")
			write(b, t.Data, quoted)
		}
		b.WriteByte(')')
	case *Function:
		writeFunction(b, t, quoted)
	case *BoundMethod:
		b.WriteString("funct ")
		b.WriteString(t.Fn.Name)
		b.WriteByte('(')
		for i, p := range t.Fn.Params {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(p.Name)
		}
		b.WriteString(")")
	case NamedArg:
		write(b, t.value, quoted)
	}
}

func (o *Object) FieldNames() []string {
	seen := map[string]bool{}
	var out []string
	for k := o.Class; k != nil; k = k.Parent {
		for _, f := range k.Fields {
			if !seen[f.Name] {
				seen[f.Name] = true
				out = append(out, f.Name)
			}
		}
	}
	return out
}

func writeFunction(b *strings.Builder, f *Function, quoted bool) {
	b.WriteString("funct ")
	if f.Name == "" {
		b.WriteString("<bigoft>")
	} else {
		b.WriteString(f.Name)
	}
	b.WriteByte('(')
	for i, p := range f.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(p.Name)
	}
	b.WriteByte(')')
	if f.Result != "" && f.Result != "batri" {
		b.WriteString(" -> ")
		b.WriteString(f.Result)
	}
}

func showInt(i Int) string { return strconv.FormatInt(int64(i), 10) }

func formatFloat(f float64) string {
	switch {
	case math.IsInf(f, 1):
		return "batarlagh"
	case math.IsInf(f, -1):
		return "batarlagh-manfi"
	case math.IsNaN(f):
		return "bimalad"
	}
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eE") {
		s += ".0"
	}
	return s
}

func Quote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func isNil(v Value) bool {
	if v == nil {
		return true
	}
	_, ok := v.(Hich)
	return ok
}

func Equal(a, b Value) bool {
	if an, aok := numeric(a); aok {
		if bn, bok := numeric(b); bok {
			return an == bn
		}
		return false
	}
	switch x := a.(type) {
	case Str:
		y, ok := b.(Str)
		return ok && x == y
	case Bool:
		y, ok := b.(Bool)
		return ok && x == y
	case Hich:
		_, ok := b.(Hich)
		return ok || b == nil
	case nil:
		_, ok := b.(Hich)
		return ok || b == nil
	case *List:
		y, ok := b.(*List)
		if !ok || len(x.Items) != len(y.Items) {
			return false
		}
		for i := range x.Items {
			if !Equal(x.Items[i], y.Items[i]) {
				return false
			}
		}
		return true
	case *Dict:
		y, ok := b.(*Dict)
		if !ok || x.Len() != y.Len() {
			return false
		}
		for _, k := range x.Keys() {
			xv, _ := x.Get(k)
			yv, found := y.Get(k)
			if !found || !Equal(xv, yv) {
				return false
			}
		}
		return true
	case *Set:
		y, ok := b.(*Set)
		if !ok || x.Len() != y.Len() {
			return false
		}
		for _, item := range x.Items() {
			if !y.Has(item) {
				return false
			}
		}
		return true
	case *Object:
		y, ok := b.(*Object)
		if !ok || x.Class != y.Class {
			return false
		}
		for _, n := range x.FieldNames() {
			if !Equal(x.Fields[n], y.Fields[n]) {
				return false
			}
		}
		return true
	case *Function:
		y, ok := b.(*Function)
		return ok && x == y
	case *BoundMethod:
		y, ok := b.(*BoundMethod)
		return ok && x.Fn == y.Fn && Equal(x.Recv, y.Recv)
	}
	return a == b
}

func numeric(v Value) (float64, bool) {
	switch t := v.(type) {
	case Int:
		return float64(t), true
	case Float:
		return float64(t), true
	case Bool:
		if t {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

func Compare(a, b Value) (int, bool) {
	if an, aok := numeric(a); aok {
		if bn, bok := numeric(b); bok {
			return cmpFloat(an, bn), true
		}
		return 0, false
	}
	switch x := a.(type) {
	case Str:
		if y, ok := b.(Str); ok {
			return strings.Compare(string(x), string(y)), true
		}
	case *List:
		if y, ok := b.(*List); ok {
			return compareLists(x, y), true
		}
	}
	return 0, false
}

func compareLists(a, b *List) int {
	for i := 0; i < len(a.Items) && i < len(b.Items); i++ {
		if c, ok := Compare(a.Items[i], b.Items[i]); ok && c != 0 {
			return c
		}
	}
	return cmpInt(len(a.Items), len(b.Items))
}

func cmpFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func HashKey(v Value) (string, bool) {
	switch t := v.(type) {
	case Int:
		return "i" + strconv.FormatInt(int64(t), 10), true
	case Float:

		if float64(t) == math.Trunc(float64(t)) && math.Abs(float64(t)) < 1e18 {
			return "i" + strconv.FormatInt(int64(t), 10), true
		}
		return "f" + strconv.FormatFloat(float64(t), 'g', -1, 64), true
	case Bool:
		if t {
			return "b1", true
		}
		return "b0", true
	case Str:
		return "s" + string(t), true
	case string:

		return "s" + t, true
	case Hich:
		return "nil", true
	case nil:
		return "nil", true
	case *Object:
		return "o?" + t.Class.Name, true
	}
	return "", false
}

func HashKeyMust(v Value) string {
	h, _ := HashKey(v)
	return h
}

func RuneCount(s string) int { return utf8.RuneCountInString(s) }

func RuneAt(s string, i int) (rune, bool) {
	n := RuneCount(s)
	if i < 0 {
		i += n
	}
	if i < 0 || i >= n {
		return 0, false
	}

	if isASCII(s) {
		return rune(s[i]), true
	}
	off := 0
	for j := 0; j < i; j++ {
		_, w := utf8.DecodeRuneInString(s[off:])
		off += w
	}
	r, _ := utf8.DecodeRuneInString(s[off:])
	return r, true
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

func Slice(s string, lo, hi int, inclusive bool) string {
	n := RuneCount(s)
	if lo < 0 {
		lo += n
	}
	if hi < 0 {
		hi += n
	}
	if inclusive {
		hi++
	}
	if lo < 0 {
		lo = 0
	}
	if hi > n {
		hi = n
	}
	if hi <= lo {
		return ""
	}

	byteLo := byteIndex(s, lo)
	byteHi := byteIndex(s, hi)
	if byteLo >= 0 && byteHi >= byteLo {
		return s[byteLo:byteHi]
	}
	if byteLo >= 0 {

		s = s[byteLo:]
		lo = 0
		hi -= lo
	}
	var b strings.Builder
	for i := lo; i < hi; i++ {
		if r, ok := RuneAt(s, i); ok {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func byteIndex(s string, i int) int {
	if i <= 0 {
		return 0
	}
	off := 0
	for j := 0; j < i; j++ {
		if off >= len(s) {
			return -1
		}
		_, w := utf8.DecodeRuneInString(s[off:])
		off += w
	}
	return off
}

func SortValues(items []Value) {

	allInt, allStr := true, true
	for _, v := range items {
		switch v.(type) {
		case Int:
			allStr = false
		case Str:
			allInt = false
		default:
			allInt, allStr = false, false
		}
		if !allInt && !allStr {
			break
		}
	}
	if allInt && len(items) > 0 {

		type pair struct {
			k int64
			v Value
		}
		ps := make([]pair, len(items))
		for i, v := range items {
			ps[i] = pair{int64(v.(Int)), v}
		}
		slices.SortFunc(ps, func(a, b pair) int { return cmp.Compare(a.k, b.k) })
		for i := range ps {
			items[i] = ps[i].v
		}
		return
	}
	if allStr && len(items) > 0 {
		slices.SortFunc(items, func(a, b Value) int { return strings.Compare(string(a.(Str)), string(b.(Str))) })
		return
	}
	sort.SliceStable(items, func(i, j int) bool {
		c, ok := Compare(items[i], items[j])
		return ok && c < 0
	})
}

func freshCopy(v Value) Value {
	switch t := v.(type) {
	case *List:
		items := make([]Value, len(t.Items))
		for i, x := range t.Items {
			items[i] = freshCopy(x)
		}
		return NewList(items...)
	case *Dict:
		d := NewDict()
		for i, k := range t.keys {
			d.Set(k, freshCopy(t.vals[i]))
		}
		return d
	case *Set:
		n := NewSet()
		for _, x := range t.order {
			n.Add(x)
		}
		return n
	}
	return v
}
