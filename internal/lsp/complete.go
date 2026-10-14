package lsp

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"mihan/internal/analysis"
	"mihan/internal/ast"
	"mihan/internal/diag"
	"mihan/internal/eval"
	"mihan/internal/format"
	"mihan/internal/ir"
	"mihan/internal/parser"
	"mihan/internal/token"
)

const (
	tUnknown = ""
	tStr     = "matn"
	tList    = "list"
	tDict    = "dict"
	tNum     = "num"
	tClass   = "class"
	tStatic  = "static"
	tModule  = "module"
)

type vtype struct {
	kind string
	name string
}

type cctx struct {
	s     *Server
	uri   string
	text  string
	off   int
	facts *analysis.File
	scope *analysis.Scope
	lines []string
	line  int
	depth int
}

const triggerHints = "editor.action.triggerParameterHints"

func inCode(text string, off int) bool { return scanCode(text, off, nil) }

func scanCode(text string, off int, visit func(i int)) bool {
	if off > len(text) {
		off = len(text)
	}
	type frame struct {
		str    byte
		interp bool
		braces int
	}
	stack := []frame{{}}
	i := 0
	for i < off {
		top := &stack[len(stack)-1]
		c := text[i]
		if top.str != 0 {
			switch {
			case c == '\\' && top.str == '"':
				i += 2
				continue
			case c == top.str:
				stack = stack[:len(stack)-1]
			case c == '\n' && top.str == '"':
				stack = stack[:len(stack)-1]
			case c == '{' && top.interp:
				stack = append(stack, frame{braces: 1})
			}
			i++
			continue
		}
		switch {
		case c == '#' || (c == '/' && i+1 < len(text) && text[i+1] == '/'):
			nl := strings.IndexByte(text[i:], '\n')
			if nl < 0 || i+nl >= off {
				return false
			}
			i += nl
			continue
		case c == '/' && i+1 < len(text) && text[i+1] == '*':
			end := strings.Index(text[i+2:], "*/")
			if end < 0 || i+2+end+2 > off {
				return false
			}
			i += 2 + end + 2
			continue
		case c == '"' || c == '`':
			interp := i > 0 && text[i-1] == '$'
			stack = append(stack, frame{str: c, interp: interp})
		case c == '\'':
			j := i + 1
			for j < len(text) && j < i+8 && text[j] != '\'' && text[j] != '\n' {
				if text[j] == '\\' {
					j++
				}
				j++
			}
			if j >= off {
				return false
			}
			i = j + 1
			continue
		case c == '{' && len(stack) > 1:
			top.braces++
			if visit != nil && top.braces > 1 {
				visit(i)
			}
			i++
			continue
		case c == '}' && len(stack) > 1:
			top.braces--
			if top.braces == 0 {
				stack = stack[:len(stack)-1]
				i++
				continue
			}
		}
		if visit != nil {
			visit(i)
		}
		i++
	}
	return stack[len(stack)-1].str == 0
}

func (s *Server) analyze(uri, text string, pos Position) (*analysis.File, diag.Pos) {
	p := s.posOf(text, pos)
	if f := s.file(uri); f != nil && f.text == text && f.facts != nil && !f.bag.HasErrors() {
		return f.facts, p
	}
	lines := strings.Split(text, "\n")
	if pos.Line >= 0 && pos.Line < len(lines) {
		cur := lines[pos.Line]
		indent := cur[:len(cur)-len(strings.TrimLeft(cur, " \t"))]
		lines[pos.Line] = indent + "hich"
		fixed := strings.Join(lines, "\n")
		bag := &diag.Bag{}
		tree := parser.Parse(uri, []byte(fixed), bag)
		if tree != nil {
			ir.Lower(tree, bag)
			np := s.posOf(fixed, Position{Line: pos.Line, Character: len([]rune(indent)) + 4})
			return analysis.Build(tree), np
		}
	}
	if f := s.file(uri); f != nil && f.facts != nil {
		return f.facts, p
	}
	if f := s.lastGood[uri]; f != nil {
		return f, p
	}
	return nil, p
}

func (s *Server) Completion(uri string, pos Position) *CompletionList {
	f := s.file(uri)
	empty := &CompletionList{Items: []CompletionItem{}}
	if f == nil {
		return &CompletionList{Items: staticItems(true)}
	}
	text := f.text
	off := s.posOf(text, pos).Off
	if !inCode(text, off) {
		if ok, isValue, key := styleContext(text, off); ok {
			return &CompletionList{Items: styleItems(isValue, key)}
		}
		return empty
	}
	facts, apos := s.analyze(uri, text, pos)
	c := &cctx{s: s, uri: uri, text: text, off: off, facts: facts, lines: strings.Split(text, "\n"), line: pos.Line}
	if facts != nil {
		c.scope = facts.ScopeAt(apos)
	}
	lineText := []rune(lineOf(text, pos.Line))
	col := pos.Character
	if col > len(lineText) {
		col = len(lineText)
	}
	start := col
	for start > 0 && isNameRune(lineText[start-1]) {
		start--
	}
	if start > 0 && lineText[start-1] == '.' {
		recv := receiverBefore(lineText, start-1)
		if recv == "" || isNumber(recv) {
			return empty
		}
		t := c.typeOfExpr(recv)
		return &CompletionList{Items: c.membersOf(t)}
	}
	if start < col && isDigitRune(lineText[start]) {
		return empty
	}
	head := strings.TrimSpace(string(lineText[:start]))
	atStart := head == ""
	afterElse := head == "warna" || head == "else"
	if strings.HasPrefix(head, "biar") || strings.HasPrefix(head, "import") {
		return empty
	}
	if head != "" && (strings.HasSuffix(head, "funct") || strings.HasSuffix(head, "darghan") || strings.HasSuffix(head, "def") || strings.HasSuffix(head, "class") || strings.HasSuffix(head, "fn")) {
		return empty
	}

	items := make([]CompletionItem, 0, 256)
	seen := map[string]bool{}
	add := func(it CompletionItem) {
		key := it.Label + "\x00" + string(rune(it.Kind))
		if seen[it.Label] && it.Kind != KindSnippet {
			return
		}
		if seen[key] {
			return
		}
		seen[it.Label] = true
		seen[key] = true
		items = append(items, it)
	}
	for _, it := range c.scopeItems() {
		add(it)
	}
	for _, it := range c.importItems() {
		add(it)
	}
	if atStart {
		for _, it := range snippetItems() {
			add(it)
		}
	} else if afterElse {
		for _, it := range snippetItems() {
			if it.Label == "agar" || it.Label == "if" {
				add(it)
			}
		}
	}
	for _, it := range staticItems(atStart) {
		add(it)
	}
	return &CompletionList{IsIncomplete: false, Items: items}
}

func blockRange(text string, lo int) (int, int) {
	if lo < 0 || lo > len(text) {
		return 0, 0
	}
	ls := strings.LastIndexByte(text[:lo], '\n') + 1
	head := text[ls:]
	if nl := strings.IndexByte(head, '\n'); nl >= 0 {
		head = head[:nl]
	}
	base := len(head) - len(strings.TrimLeft(head, " \t"))
	end := ls + len(head)
	for end < len(text) {
		next := text[end+1:]
		line := next
		if nl := strings.IndexByte(next, '\n'); nl >= 0 {
			line = next[:nl]
		}
		trim := strings.TrimLeft(line, " \t")
		if trim != "" && len(line)-len(trim) <= base {
			break
		}
		end += 1 + len(line)
	}
	return ls, end
}

func isDigitRune(r rune) bool { return r >= '0' && r <= '9' }

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !isDigitRune(r) && r != '_' {
			return false
		}
	}
	return true
}

func receiverBefore(line []rune, dot int) string {
	i := dot
	depth := 0
	for i > 0 {
		r := line[i-1]
		if depth > 0 {
			switch r {
			case ')', ']', '}':
				depth++
			case '(', '[', '{':
				depth--
			case '"':
				j := i - 2
				for j >= 0 && line[j] != '"' {
					j--
				}
				if j < 0 {
					return ""
				}
				i = j + 1
			}
			i--
			continue
		}
		switch {
		case isNameRune(r) || r == '.':
			i--
		case r == ')' || r == ']' || r == '}':
			depth++
			i--
		case r == '"' || r == '`':
			j := i - 2
			for j >= 0 && line[j] != r {
				j--
			}
			if j < 0 {
				return ""
			}
			i = j
			if i > 0 && line[i-1] == '$' {
				i--
			}
		default:
			return strings.TrimSpace(string(line[i:dot]))
		}
	}
	return strings.TrimSpace(string(line[i:dot]))
}

func mapType(t string, c *cctx) vtype {
	switch strings.ToLower(t) {
	case "matn", "string", "str", "harf", "char", "rune":
		return vtype{kind: tStr}
	case "list", "araye", "array", "rang":
		return vtype{kind: tList}
	case "dict", "map":
		return vtype{kind: tDict}
	case "arz", "int", "kasr", "float", "double", "adad", "ashari":
		return vtype{kind: tNum}
	}
	if c != nil && c.classDecl(t) != nil {
		return vtype{kind: tClass, name: t}
	}
	return vtype{}
}

func (c *cctx) classDecl(name string) *classInfo {
	if c.facts == nil || c.facts.Tree == nil || name == "" {
		return nil
	}
	for _, d := range c.facts.Tree.Decls {
		switch x := d.(type) {
		case *ast.ClassDecl:
			if x.Name == name {
				return &classInfo{name: x.Name, bases: x.Bases, fields: x.Fields, methods: x.Methods, span: x.Span()}
			}
		case *ast.StructDecl:
			if x.Name == name {
				return &classInfo{name: x.Name, bases: x.Bases, fields: x.Fields, methods: x.Methods, span: x.Span()}
			}
		}
	}
	return nil
}

type classInfo struct {
	name    string
	bases   []string
	fields  []*ast.FieldDecl
	methods []*ast.FuncDecl
	span    diag.Span
}

func (c *cctx) funcDecl(name string) *ast.FuncDecl {
	if c.facts == nil || c.facts.Tree == nil {
		return nil
	}
	for _, d := range c.facts.Tree.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name == name {
			return fn
		}
	}
	return nil
}

func (c *cctx) enclosingClass() string {
	if c.facts == nil || c.facts.Tree == nil {
		return ""
	}
	best := ""
	for _, d := range c.facts.Tree.Decls {
		var name string
		var sp diag.Span
		switch x := d.(type) {
		case *ast.ClassDecl:
			name, sp = x.Name, x.Span()
		case *ast.StructDecl:
			name, sp = x.Name, x.Span()
		default:
			continue
		}
		lo, hi := blockRange(c.text, sp.Start.Off)
		if lo <= c.off && c.off <= hi+1 {
			best = name
		}
	}
	if best != "" {
		return best
	}
	for i := c.line; i >= 0 && i < len(c.lines); i-- {
		l := c.lines[i]
		t := strings.TrimLeft(l, " \t")
		if len(t) == len(l) && t != "" {
			for _, kw := range []string{"darghan ", "class ", "sakht ", "struct "} {
				if strings.HasPrefix(t, kw) {
					n := strings.TrimSpace(t[len(kw):])
					j := 0
					for _, r := range n {
						if !isNameRune(r) {
							break
						}
						j += len(string(r))
					}
					return n[:j]
				}
			}
			if i != c.line {
				return ""
			}
		}
	}
	return ""
}

var importLine = regexp.MustCompile(`^\s*(?:biar|import|use)\s+["` + "`" + `]([^"` + "`" + `]+)["` + "`" + `](?:\s+as\s+([\p{L}_][\p{L}\p{N}_]*))?`)

type importInfo struct {
	path, alias string
}

func (c *cctx) imports() []importInfo {
	var out []importInfo
	for _, l := range c.lines {
		if m := importLine.FindStringSubmatch(l); m != nil {
			out = append(out, importInfo{path: m[1], alias: m[2]})
		}
	}
	return out
}

func (c *cctx) typeOfName(name string) vtype {
	if name == "khod" || name == "this" || name == "self" {
		if cl := c.enclosingClass(); cl != "" {
			return vtype{kind: tClass, name: cl}
		}
		return vtype{}
	}
	for _, im := range c.imports() {
		if im.alias == name {
			return vtype{kind: tModule, name: im.path}
		}
	}
	if c.classDecl(name) != nil {
		return vtype{kind: tStatic, name: name}
	}
	if c.scope != nil {
		if d := c.scope.Lookup(name); d != nil {
			if d.TypeName != "" {
				if t := mapType(d.TypeName, c); t.kind != "" {
					return t
				}
			}
			if d.Kind == analysis.DeclParam {
				if t := c.paramDefault(name); t.kind != "" {
					return t
				}
			}
		}
	}
	return c.typeFromAssignments(name)
}

func (c *cctx) paramDefault(name string) vtype {
	if c.facts == nil || c.facts.Tree == nil {
		return vtype{}
	}
	var visit func(fns []*ast.FuncDecl) vtype
	visit = func(fns []*ast.FuncDecl) vtype {
		for _, fn := range fns {
			lo, hi := blockRange(c.text, fn.NamePos.Off)
			if !(lo <= c.off && c.off <= hi+1) {
				continue
			}
			for _, p := range fn.Params {
				if p.Name == name && p.Default != nil {
					return c.typeOfExpr(strings.TrimSpace(format.Expr(p.Default)))
				}
			}
		}
		return vtype{}
	}
	var fns []*ast.FuncDecl
	for _, d := range c.facts.Tree.Decls {
		switch x := d.(type) {
		case *ast.FuncDecl:
			fns = append(fns, x)
		case *ast.ClassDecl:
			fns = append(fns, x.Methods...)
		}
	}
	return visit(fns)
}

func assignRe(name string) *regexp.Regexp {
	return regexp.MustCompile(`^\s*(?:([\p{L}_][\p{L}\p{N}_]*)\s+)?` + regexp.QuoteMeta(name) + `\s*(?::\s*([\p{L}_][\p{L}\p{N}_]*)\s*)?=([^=].*)$`)
}

func (c *cctx) typeFromAssignments(name string) vtype {
	if c.depth > 6 {
		return vtype{}
	}
	re := assignRe(name)
	last := c.line
	if last >= len(c.lines) {
		last = len(c.lines) - 1
	}
	for i := last; i >= 0; i-- {
		l := c.lines[i]
		if i == c.line {
			continue
		}
		m := re.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		for _, tn := range []string{m[1], m[2]} {
			if tn != "" {
				if t := mapType(tn, c); t.kind != "" {
					return t
				}
			}
		}
		rhs := strings.TrimSpace(m[3])
		if t := c.sub().typeOfExpr(rhs); t.kind != "" {
			return t
		}
	}
	return vtype{}
}

func (c *cctx) sub() *cctx {
	n := *c
	n.depth++
	return &n
}

func (c *cctx) typeOfExpr(e string) vtype {
	if c.depth > 8 {
		return vtype{}
	}
	r := []rune(strings.TrimSpace(e))
	if len(r) == 0 {
		return vtype{}
	}
	i := 0
	var cur vtype
	lastName := ""
	switch {
	case r[0] == '"' || r[0] == '`' || (r[0] == '$' && len(r) > 1 && (r[1] == '"' || r[1] == '`')):
		q := r[0]
		j := 1
		if q == '$' {
			q = r[1]
			j = 2
		}
		for j < len(r) && r[j] != q {
			if r[j] == '\\' {
				j++
			}
			j++
		}
		i = j + 1
		cur = vtype{kind: tStr}
	case r[0] == '[':
		i = matchClose(r, 0)
		cur = vtype{kind: tList}
	case r[0] == '{':
		i = matchClose(r, 0)
		cur = vtype{kind: tDict}
	case isDigitRune(r[0]) || r[0] == '-':
		return vtype{kind: tNum}
	case isNameRune(r[0]):
		j := 0
		for j < len(r) && isNameRune(r[j]) {
			j++
		}
		lastName = string(r[:j])
		i = j
		if i < len(r) && r[i] == '(' {
			end := matchClose(r, i)
			cur = c.callResult(lastName)
			i = end
			lastName = ""
		} else {
			cur = c.sub().typeOfName(lastName)
		}
	default:
		return vtype{}
	}
	for i < len(r) {
		switch r[i] {
		case '.':
			j := i + 1
			for j < len(r) && isNameRune(r[j]) {
				j++
			}
			name := string(r[i+1 : j])
			i = j
			if i < len(r) && r[i] == '(' {
				i = matchClose(r, i)
				cur = c.methodResult(cur, name)
			} else {
				cur = c.fieldType(cur, name)
			}
		case '[':
			i = matchClose(r, i)
			cur = vtype{}
		default:
			return cur
		}
		if cur.kind == "" {
			return cur
		}
	}
	return cur
}

func matchClose(r []rune, open int) int {
	depth := 0
	for i := open; i < len(r); i++ {
		switch r[i] {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
			if depth == 0 {
				return i + 1
			}
		case '"':
			i++
			for i < len(r) && r[i] != '"' {
				if r[i] == '\\' {
					i++
				}
				i++
			}
		}
	}
	return len(r)
}

func (c *cctx) callResult(name string) vtype {
	if c.classDecl(name) != nil {
		return vtype{kind: tClass, name: name}
	}
	if fn := c.funcDecl(name); fn != nil {
		return c.funcResult(fn)
	}
	for _, b := range eval.Builtins() {
		if b.Name == name || b.Alias == name {
			return mapType(b.Result, c)
		}
	}
	return vtype{}
}

var returnRe = regexp.MustCompile(`^\s*(?:bargard|return)\s+(.+)$`)

func (c *cctx) funcResult(fn *ast.FuncDecl) vtype {
	if fn.Result != "" {
		if t := mapType(fn.Result, c); t.kind != "" {
			return t
		}
	}
	if c.depth > 4 {
		return vtype{}
	}
	lo, hi := blockRange(c.text, fn.NamePos.Off)
	if lo < 0 || hi > len(c.text) || lo >= hi {
		return vtype{}
	}
	for _, l := range strings.Split(c.text[lo:hi], "\n") {
		if m := returnRe.FindStringSubmatch(l); m != nil {
			n := c.sub()
			n.off = lo
			if t := n.typeOfExpr(m[1]); t.kind != "" && t.kind != tNum {
				return t
			}
		}
	}
	return vtype{}
}

func methodTable(kind string) []bmethod {
	switch kind {
	case tStr:
		return strMethods
	case tList:
		return listMethods
	case tDict:
		return dictMethods
	}
	return nil
}

func (c *cctx) methodResult(t vtype, name string) vtype {
	for _, m := range methodTable(t.kind) {
		if m.name == name || m.alias == name {
			return mapType(m.result, c)
		}
	}
	if t.kind == tClass || t.kind == tStatic {
		for _, cl := range c.classChain(t.name) {
			for _, m := range cl.methods {
				if m.Name == name {
					n := c.sub()
					return n.funcResult(m)
				}
			}
		}
	}
	if t.kind == tModule {
		if mod := c.s.module(c.uri, t.name); mod != nil {
			for _, fn := range mod.funcs {
				if fn.Name == name && fn.Result != "" {
					return mapType(fn.Result, nil)
				}
			}
			for _, cl := range mod.classes {
				if cl == name {
					return vtype{}
				}
			}
		}
	}
	return vtype{}
}

func (c *cctx) fieldType(t vtype, name string) vtype {
	if t.kind != tClass && t.kind != tStatic {
		return vtype{}
	}
	for _, cl := range c.classChain(t.name) {
		for _, f := range cl.fields {
			if f.Name != name {
				continue
			}
			if f.TypeName != "" {
				return mapType(f.TypeName, c)
			}
			if f.Default != nil {
				return c.sub().typeOfExpr(strings.TrimSpace(format.Expr(f.Default)))
			}
		}
		lo, hi := blockRange(c.text, cl.span.Start.Off)
		if lo >= 0 && hi <= len(c.text) && lo < hi {
			re := regexp.MustCompile(`khod\.` + regexp.QuoteMeta(name) + `\s*=([^=].*)$`)
			for _, l := range strings.Split(c.text[lo:hi], "\n") {
				if m := re.FindStringSubmatch(l); m != nil {
					rhs := strings.TrimSpace(m[1])
					if rhs == name {
						continue
					}
					if t := c.sub().typeOfExpr(rhs); t.kind != "" {
						return t
					}
				}
			}
		}
	}
	return vtype{}
}

func (c *cctx) classChain(name string) []*classInfo {
	var out []*classInfo
	seen := map[string]bool{}
	var walk func(n string)
	walk = func(n string) {
		n = strings.TrimSpace(n)
		if seen[n] || len(out) > 16 {
			return
		}
		seen[n] = true
		cl := c.classDecl(n)
		if cl == nil {
			return
		}
		out = append(out, cl)
		for _, b := range cl.bases {
			walk(b)
		}
	}
	walk(name)
	return out
}

var khodAssign = regexp.MustCompile(`khod\.([\p{L}_][\p{L}\p{N}_]*)\s*=[^=]`)

func methodItem(m bmethod, owner string, english bool) CompletionItem {
	label := m.name
	other := m.alias
	if english {
		label, other = m.alias, m.name
	}
	sig := label + "(" + m.params + ")"
	if m.result != "" && m.result != "any" {
		sig += " → " + m.result
	}
	doc := "**" + m.fa + "**\n\n" + m.en
	if other != "" {
		doc += "\n\n`" + label + "` = `" + other + "`"
	}
	doc += "\n\n```mihan\n" + owner + "." + sig + "\n```"
	it := CompletionItem{
		Label:            label,
		Kind:             KindMethod,
		Detail:           sig,
		Documentation:    &MarkupContent{Kind: Markup, Value: doc},
		InsertTextFormat: InsertSnippet,
		SortText:         "1" + label,
	}
	if english {
		it.SortText = "2" + label
	}
	if m.params == "" {
		it.InsertText = label + "()"
	} else {
		it.InsertText = label + "($1)$0"
		it.Command = &Command{Title: "signature", Command: triggerHints}
	}
	return it
}

func (c *cctx) membersOf(t vtype) []CompletionItem {
	items := []CompletionItem{}
	switch t.kind {
	case tStr, tList, tDict:
		for _, m := range methodTable(t.kind) {
			items = append(items, methodItem(m, t.kind, false))
		}
		for _, m := range methodTable(t.kind) {
			if m.alias != "" {
				items = append(items, methodItem(m, t.kind, true))
			}
		}
		return items
	case tClass, tStatic:
		return c.classMembers(t.name)
	case tModule:
		return c.moduleMembers(t.name, "0")
	case tNum:
		return items
	}
	seen := map[string]bool{}
	if c.facts != nil && c.facts.Tree != nil {
		for _, d := range c.facts.Tree.Decls {
			if cl, ok := d.(*ast.ClassDecl); ok {
				for _, it := range c.classMembers(cl.Name) {
					if !seen[it.Label] {
						seen[it.Label] = true
						it.SortText = "3" + it.Label
						items = append(items, it)
					}
				}
			}
		}
	}
	for _, kind := range []string{tStr, tList, tDict} {
		for _, m := range methodTable(kind) {
			if !seen[m.name] {
				seen[m.name] = true
				it := methodItem(m, kind, false)
				it.SortText = "4" + m.name
				items = append(items, it)
			}
		}
	}
	return items
}

func (c *cctx) classMembers(name string) []CompletionItem {
	var items []CompletionItem
	seen := map[string]bool{}
	for depth, cl := range c.classChain(name) {
		rank := string(rune('0' + depth%10))
		for _, f := range cl.fields {
			if seen[f.Name] {
				continue
			}
			seen[f.Name] = true
			detail := "field of " + cl.name
			if f.TypeName != "" {
				detail = f.TypeName + " · " + detail
			}
			items = append(items, CompletionItem{
				Label:            f.Name,
				Kind:             KindField,
				Detail:           detail,
				Documentation:    &MarkupContent{Kind: Markup, Value: "**ویژگی " + cl.name + "**\n\nfield of `" + cl.name + "`"},
				SortText:         "0" + rank + f.Name,
				CommitCharacters: []string{"."},
			})
		}
		lo, hi := blockRange(c.text, cl.span.Start.Off)
		if lo >= 0 && hi <= len(c.text) && lo < hi {
			for _, m := range khodAssign.FindAllStringSubmatch(c.text[lo:hi], -1) {
				if seen[m[1]] {
					continue
				}
				seen[m[1]] = true
				items = append(items, CompletionItem{
					Label:            m[1],
					Kind:             KindField,
					Detail:           "field of " + cl.name,
					Documentation:    &MarkupContent{Kind: Markup, Value: "**ویژگی " + cl.name + "**\n\nset with `khod." + m[1] + "`"},
					SortText:         "0" + rank + m[1],
					CommitCharacters: []string{"."},
				})
			}
		}
		for _, m := range cl.methods {
			if seen[m.Name] || strings.HasPrefix(m.Name, "__") {
				continue
			}
			seen[m.Name] = true
			label, params := funcSignature(m)
			label = strings.TrimPrefix(label, "funct ")
			it := CompletionItem{
				Label:            m.Name,
				Kind:             KindMethod,
				Detail:           label,
				Documentation:    &MarkupContent{Kind: Markup, Value: "**متد " + cl.name + "**\n\nmethod of `" + cl.name + "`\n\n```mihan\n" + label + "\n```" + docSuffix(m.Doc)},
				InsertTextFormat: InsertSnippet,
				SortText:         "1" + rank + m.Name,
			}
			if len(params) == 0 {
				it.InsertText = m.Name + "()"
			} else {
				it.InsertText = m.Name + "($1)$0"
				it.Command = &Command{Title: "signature", Command: triggerHints}
			}
			items = append(items, it)
		}
	}
	return items
}

func docSuffix(d string) string {
	if strings.TrimSpace(d) == "" {
		return ""
	}
	return "\n\n" + strings.TrimSpace(d)
}

func (c *cctx) scopeItems() []CompletionItem {
	var items []CompletionItem
	if c.scope == nil {
		return items
	}
	seen := map[string]bool{}
	for sc := c.scope; sc != nil; sc = sc.Parent {
		global := sc.Parent == nil
		for i := len(sc.Decls) - 1; i >= 0; i-- {
			d := sc.Decls[i]
			if seen[d.Name] || d.Name == "" {
				continue
			}
			if d.Pos.Line-1 == c.line && d.Kind != analysis.DeclFunc && d.Kind != analysis.DeclClass {
				continue
			}
			seen[d.Name] = true
			items = append(items, c.declItem(d, global))
		}
	}
	return items
}

func (c *cctx) declItem(d *analysis.Decl, global bool) CompletionItem {
	rank := "0"
	where := "محلی · local"
	if global {
		rank = "1"
		where = "سراسری · global"
	}
	it := CompletionItem{Label: d.Name, Kind: kindOf(d.Kind), SortText: rank + d.Name}
	switch d.Kind {
	case analysis.DeclParam:
		it.SortText = "0" + d.Name
		where = "پارامتر · parameter"
	case analysis.DeclFor, analysis.DeclForIn:
		where = "متغیر حلقه · loop variable"
	case analysis.DeclCatch:
		where = "خطای گرفته‌شده · caught error"
	}
	switch d.Kind {
	case analysis.DeclFunc:
		label, params, _ := declSignature(d)
		it.Detail = label
		fn, _ := d.Node.(*ast.FuncDecl)
		doc := ""
		if fn != nil {
			doc = docSuffix(fn.Doc)
		}
		it.Documentation = &MarkupContent{Kind: Markup, Value: "**تابع** · function\n\n```mihan\n" + label + "\n```\n\nline " + itoa(d.Pos.Line) + doc}
		it.InsertTextFormat = InsertSnippet
		if len(params) == 0 {
			it.InsertText = d.Name + "()"
		} else {
			it.InsertText = d.Name + "($1)$0"
			it.Command = &Command{Title: "signature", Command: triggerHints}
		}
		it.SortText = rank + "1" + d.Name
	case analysis.DeclClass:
		it.Detail = "darghan " + d.Name
		it.Documentation = &MarkupContent{Kind: Markup, Value: "**کلاس** · class `" + d.Name + "`\n\nline " + itoa(d.Pos.Line)}
		it.CommitCharacters = []string{"."}
		it.SortText = rank + "1" + d.Name
	case analysis.DeclEnum, analysis.DeclIface:
		it.Detail = d.Kind.String()
	default:
		n := c.sub()
		n.off = d.Pos.Off
		t := vtype{}
		if d.TypeName != "" {
			t = mapType(d.TypeName, c)
		}
		tn := d.TypeName
		if tn == "" && t.kind == "" {
			t = c.typeFromAssignments(d.Name)
		}
		if tn == "" {
			switch t.kind {
			case tClass:
				tn = t.name
			case tStr, tList, tDict:
				tn = t.kind
			case tNum:
				tn = "adad"
			}
		}
		it.Detail = where
		if tn != "" {
			it.Detail = tn + " · " + where
		}
		it.Documentation = &MarkupContent{Kind: Markup, Value: "```mihan\n" + d.Name + "\n```\n\n" + where + "\n\nline " + itoa(d.Pos.Line)}
		it.CommitCharacters = []string{"."}
	}
	return it
}

type moduleInfo struct {
	mtime   time.Time
	funcs   []*ast.FuncDecl
	classes []string
	vars    []string
}

var moduleCache = struct {
	sync.Mutex
	m map[string]*moduleInfo
}{m: map[string]*moduleInfo{}}

func uriPath(uri string) string {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" {
		return ""
	}
	p := u.Path
	if len(p) > 2 && p[0] == '/' && p[2] == ':' {
		p = p[1:]
	}
	return filepath.FromSlash(p)
}

func (s *Server) module(uri, path string) *moduleInfo {
	base := uriPath(uri)
	var cands []string
	if filepath.IsAbs(path) {
		cands = []string{path, path + ".mhn"}
	} else if base != "" {
		dir := filepath.Dir(base)
		cands = []string{filepath.Join(dir, path), filepath.Join(dir, path+".mhn")}
	}
	for _, p := range cands {
		st, err := os.Stat(p)
		if err != nil || st.IsDir() {
			continue
		}
		moduleCache.Lock()
		m := moduleCache.m[p]
		moduleCache.Unlock()
		if m != nil && m.mtime.Equal(st.ModTime()) {
			return m
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		bag := &diag.Bag{}
		tree := parser.Parse(p, src, bag)
		m = &moduleInfo{mtime: st.ModTime()}
		if tree != nil {
			seen := map[string]bool{}
			for _, d := range tree.Decls {
				switch x := d.(type) {
				case *ast.FuncDecl:
					m.funcs = append(m.funcs, x)
				case *ast.ClassDecl:
					m.classes = append(m.classes, x.Name)
				case *ast.StructDecl:
					m.classes = append(m.classes, x.Name)
				case *ast.VarDecl:
					if !seen[x.Name] {
						seen[x.Name] = true
						m.vars = append(m.vars, x.Name)
					}
				case *ast.ExprStmt:
					if as, ok := x.X.(*ast.Assign); ok {
						if id, ok := as.Target.(*ast.Ident); ok && !seen[id.Name] {
							seen[id.Name] = true
							m.vars = append(m.vars, id.Name)
						}
					}
				}
			}
		}
		moduleCache.Lock()
		moduleCache.m[p] = m
		moduleCache.Unlock()
		return m
	}
	return nil
}

func (c *cctx) moduleMembers(path, rank string) []CompletionItem {
	var items []CompletionItem
	m := c.s.module(c.uri, path)
	if m == nil {
		return items
	}
	from := "از " + path + " · from " + path
	for _, fn := range m.funcs {
		label, params := funcSignature(fn)
		it := CompletionItem{
			Label:            fn.Name,
			Kind:             KindFunction,
			Detail:           label,
			Documentation:    &MarkupContent{Kind: Markup, Value: "```mihan\n" + label + "\n```\n\n" + from + docSuffix(fn.Doc)},
			InsertTextFormat: InsertSnippet,
			SortText:         rank + fn.Name,
		}
		if len(params) == 0 {
			it.InsertText = fn.Name + "()"
		} else {
			it.InsertText = fn.Name + "($1)$0"
			it.Command = &Command{Title: "signature", Command: triggerHints}
		}
		items = append(items, it)
	}
	for _, cl := range m.classes {
		items = append(items, CompletionItem{Label: cl, Kind: KindClass, Detail: "darghan " + cl, Documentation: &MarkupContent{Kind: Markup, Value: from}, SortText: rank + cl})
	}
	for _, v := range m.vars {
		items = append(items, CompletionItem{Label: v, Kind: KindVariable, Detail: from, SortText: rank + v})
	}
	return items
}

func (c *cctx) importItems() []CompletionItem {
	var items []CompletionItem
	for _, im := range c.imports() {
		if im.alias != "" {
			items = append(items, CompletionItem{
				Label:            im.alias,
				Kind:             KindModule,
				Detail:           "biar \"" + im.path + "\"",
				Documentation:    &MarkupContent{Kind: Markup, Value: "**ماژول** · module `" + im.path + "`"},
				SortText:         "1" + im.alias,
				CommitCharacters: []string{"."},
			})
			continue
		}
		items = append(items, c.moduleMembers(im.path, "2")...)
	}
	return items
}

type snip struct {
	label, body, fa, en string
}

var snippets = []snip{
	{"agar", "agar ${1:shart}\n\t$0", "بلوک شرط", "if block"},
	{"agar…warna", "agar ${1:shart}\n\t${2}\nwarna\n\t$0", "شرط با شاخه‌ی وگرنه", "if / else"},
	{"warna agar", "warna agar ${1:shart}\n\t$0", "شرط بعدی", "else if"},
	{"warna", "warna\n\t$0", "شاخه‌ی وگرنه", "else block"},
	{"har…dar", "har ${1:x} dar ${2:list}\n\t$0", "حلقه روی هر عضو", "for each item"},
	{"har…az…ta", "har ${1:i} az ${2:0} ta ${3:10}\n\t$0", "حلقه‌ی شمارشی", "counted loop"},
	{"ta", "ta ${1:shart}\n\t$0", "حلقه تا وقتی شرط درست است", "while loop"},
	{"funct", "funct ${1:naam}(${2})\n\t${0:bargard hich}", "تعریف تابع", "function"},
	{"darghan", "darghan ${1:Naam}\n\tfunct __sakht__(${2:x})\n\t\tkhod.${2:x} = ${2:x}\n\n\tfunct ${3:kar}()\n\t\t$0", "تعریف کلاس با سازنده", "class with constructor"},
	{"koshish", "koshish\n\t${1}\ngiroft khata ${2:e}\n\t$0", "امتحان و گرفتن خطا", "try / catch"},
	{"koshish…nemigir", "koshish\n\t${1}\ngiroft khata ${2:e}\n\t${3}\nnemigir\n\t$0", "امتحان، گرفتن خطا و پایان", "try / catch / finally"},
	{"biar", "biar \"${1:abzar.mhn}\"$0", "آوردن ماژول", "import a module"},
	{"if", "agar ${1:shart}\n\t$0", "بلوک شرط (agar)", "if block"},
	{"if…else", "agar ${1:shart}\n\t${2}\nwarna\n\t$0", "شرط با وگرنه (agar / warna)", "if / else"},
	{"for…in", "har ${1:x} dar ${2:list}\n\t$0", "حلقه روی هر عضو (har … dar)", "for each item"},
	{"while", "ta ${1:shart}\n\t$0", "حلقه‌ی ta", "while loop"},
	{"def", "funct ${1:name}(${2})\n\t${0:bargard hich}", "تعریف تابع (funct)", "function"},
	{"class", "darghan ${1:Naam}\n\tfunct __sakht__(${2:x})\n\t\tkhod.${2:x} = ${2:x}\n\t$0", "تعریف کلاس (darghan)", "class"},
	{"try", "koshish\n\t${1}\ngiroft khata ${2:e}\n\t$0", "امتحان و گرفتن خطا (koshish)", "try / catch"},
}

var snippetOnce sync.Once
var snippetCache []CompletionItem

func snippetItems() []CompletionItem {
	snippetOnce.Do(func() {
		for _, s := range snippets {
			filter := strings.Split(strings.ReplaceAll(s.label, "…", " "), " ")[0]
			preview := strings.ReplaceAll(regexp.MustCompile(`\$\{\d+:([^}]*)\}|\$\d+`).ReplaceAllString(s.body, "$1"), "\t", "    ")
			snippetCache = append(snippetCache, CompletionItem{
				Label:            s.label,
				Kind:             KindSnippet,
				Detail:           s.fa + " · " + s.en,
				Documentation:    &MarkupContent{Kind: Markup, Value: "**" + s.fa + "** · " + s.en + "\n\n```mihan\n" + preview + "\n```"},
				InsertText:       s.body,
				InsertTextFormat: InsertSnippet,
				FilterText:       filter,
				SortText:         "3" + s.label,
			})
		}
	})
	return snippetCache
}

var staticOnce sync.Once
var staticAll, staticMid []CompletionItem

func lineStartOnly(k *token.Keyword) bool {
	switch k.Group {
	case token.GroupControl, token.GroupDeclaration:
		return true
	}
	return false
}

func builtinSig(b eval.BuiltinInfo) (string, []string) {
	n := builtinNotes[b.Name]
	var params []string
	for _, p := range strings.Split(n.params, ",") {
		if p = strings.TrimSpace(p); p != "" {
			params = append(params, p)
		}
	}
	return b.Name + "(" + strings.Join(params, ", ") + ")", params
}

func builtinDoc(b eval.BuiltinInfo, label string) string {
	sig, _ := builtinSig(b)
	if label != b.Name {
		sig = label + strings.TrimPrefix(sig, b.Name)
	}
	if b.Result != "" && b.Result != "any" {
		sig += " → " + b.Result
	}
	doc := ""
	if fa := builtinNotes[b.Name].fa; fa != "" {
		doc = "**" + fa + "**\n\n"
	}
	doc += b.Doc
	if b.Alias != "" {
		doc += "\n\n`" + b.Name + "` = `" + b.Alias + "`"
	}
	return doc + "\n\n```mihan\n" + sig + "\n```"
}

var plainBuiltins = map[string]bool{"chap": true, "chap_khat": true, "salam": true, "koshish": true}

func staticItems(atStart bool) []CompletionItem {
	staticOnce.Do(func() {
		seen := map[string]bool{}
		for _, b := range eval.Builtins() {
			if strings.Contains(b.Name, " ") {
				continue
			}
			sig, params := builtinSig(b)
			if b.Result != "" && b.Result != "any" {
				sig += " → " + b.Result
			}
			for i, label := range []string{b.Name, b.Alias} {
				if label == "" || seen[label] {
					continue
				}
				seen[label] = true
				it := CompletionItem{
					Label:         label,
					Kind:          KindFunction,
					Detail:        sig,
					Documentation: &MarkupContent{Kind: Markup, Value: builtinDoc(b, label)},
					SortText:      "4" + label,
				}
				if i == 1 {
					it.SortText = "6" + label
					it.Detail = "= " + b.Name + " · " + sig
				}
				switch {
				case plainBuiltins[b.Name]:
					it.InsertText = label + " "
				case b.MaxArgs == 0:
					it.InsertText = label + "()"
					it.InsertTextFormat = InsertSnippet
				default:
					it.InsertText = label + "($1)$0"
					it.InsertTextFormat = InsertSnippet
					it.Command = &Command{Title: "signature", Command: triggerHints}
				}
				_ = params
				staticAll = append(staticAll, it)
				staticMid = append(staticMid, it)
			}
		}
		for i := range token.Keywords {
			k := &token.Keywords[i]
			fa := keywordFA[k.Name]
			doc := k.Doc
			if doc == "" {
				doc = k.Name
			}
			if fa != "" {
				doc = "**" + fa + "**\n\n" + doc
			}
			kind := KindKeyword
			switch {
			case k.Name == "dorost" || k.Name == "ghalat" || k.Name == "hich":
				kind = KindConstant
			case k.Group == token.GroupType:
				kind = KindTypeParameter
			case k.Group == token.GroupBuiltin:
				kind = KindFunction
			}
			group := groupFA[k.Group] + " · " + k.Group.String()
			for j, word := range append([]string{k.Name}, k.Aliases...) {
				if seen[word] {
					continue
				}
				seen[word] = true
				it := CompletionItem{
					Label:         word,
					Kind:          kind,
					Detail:        group,
					Documentation: &MarkupContent{Kind: Markup, Value: doc},
					SortText:      "5" + word,
				}
				if j > 0 {
					it.Detail = "= " + k.Name + " · " + group
					it.SortText = "7" + word
				}
				staticAll = append(staticAll, it)
				if !lineStartOnly(k) {
					staticMid = append(staticMid, it)
				}
			}
		}
	})
	if atStart {
		return staticAll
	}
	return staticMid
}

var groupFA = map[token.Group]string{
	token.GroupControl:     "کنترل",
	token.GroupDeclaration: "تعریف",
	token.GroupOperator:    "عملگر",
	token.GroupType:        "نوع",
	token.GroupModifier:    "پیراینده",
	token.GroupBuiltin:     "آماده",
	token.GroupMisc:        "مقدار",
}
