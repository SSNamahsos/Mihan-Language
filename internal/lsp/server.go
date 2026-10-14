package lsp

import (
	"strings"
	"unicode/utf8"

	"mihan/internal/analysis"
	"mihan/internal/ast"
	"mihan/internal/diag"
	"mihan/internal/eval"
	"mihan/internal/format"
	"mihan/internal/ir"
	"mihan/internal/parser"
	"mihan/internal/token"
)

type Server struct {
	conn *Conn

	files map[string]*open

	version string

	lastGood map[string]*analysis.File
}

type open struct {
	uri  string
	text string
	tree *ast.File

	facts *analysis.File

	bag *diag.Bag
}

func NewServer(conn *Conn, version string) *Server {
	s := &Server{conn: conn, files: map[string]*open{}, version: version, lastGood: map[string]*analysis.File{}}

	for _, b := range eval.Builtins() {
		analysis.Builtins[b.Name] = b.Doc
	}
	conn.SetHandler(s)
	return s
}

func (s *Server) Initialize(params *InitializeParams) *InitializeResult {
	return &InitializeResult{
		ServerInfo: ServerInfo{Name: "mihan-ls", Version: s.version},
		Capabilities: ServerCapabilities{
			TextDocumentSync:           1,
			HoverProvider:              true,
			DefinitionProvider:         true,
			DocumentSymbolProvider:     true,
			DocumentFormattingProvider: true,
			SignatureHelpProvider: &SignatureHelpOptions{
				TriggerCharacters:   []string{"(", ","},
				RetriggerCharacters: []string{","},
			},
			CompletionProvider: &CompletionOptions{

				TriggerCharacters: []string{".", "\""},
			},
		},
	}
}

func (s *Server) Shutdown() {}

func (s *Server) DidOpen(uri, text string) {
	s.set(uri, text)
	s.publish(uri)
}

func (s *Server) DidChange(uri, text string) {
	s.set(uri, text)
	s.publish(uri)
}

func (s *Server) DidClose(uri string) {
	delete(s.files, uri)
	delete(s.lastGood, uri)
}

func (s *Server) set(uri, text string) {
	f := &open{uri: uri, text: text, bag: &diag.Bag{}}
	f.tree = parser.Parse(uri, []byte(text), f.bag)
	if f.tree == nil {
		s.files[uri] = f
		return
	}

	ir.Lower(f.tree, f.bag)
	f.facts = analysis.Build(f.tree)
	if !f.bag.HasErrors() {
		s.lastGood[uri] = f.facts
	}
	s.files[uri] = f
}

func (s *Server) file(uri string) *open { return s.files[uri] }

func (s *Server) publish(uri string) {
	f := s.file(uri)
	if f == nil {
		return
	}
	s.conn.Notify("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnosticsOf(f.text, f.bag),
	})
}

func diagnosticsOf(text string, bag *diag.Bag) []Diagnostic {
	if bag == nil || !bag.HasErrors() {
		return []Diagnostic{}
	}
	out := make([]Diagnostic, 0, len(bag.Sorted()))
	for _, d := range bag.Sorted() {
		msg := d.FA
		out = append(out, Diagnostic{
			Range:    rangeOf(text, d.Span),
			Severity: 1,
			Code:     d.Code,
			Source:   "mihan",
			Message:  msg,
		})
	}
	return out
}

func kindOf(k analysis.DeclKind) CompletionItemKind {
	switch k {
	case analysis.DeclFunc:
		return KindFunction
	case analysis.DeclClass:
		return KindClass
	case analysis.DeclEnum:
		return KindEnum
	case analysis.DeclIface:
		return KindClass
	case analysis.DeclMethod:
		return KindMethod
	case analysis.DeclParam:
		return KindVariable
	}
	return KindVariable
}

func isNameRune(r rune) bool {
	return r == '_' ||
		(r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
		(r >= 0x80 && r != ' ')
}

func (s *Server) wordRange(text string, pos Position) (lo, hi int, word string) {
	lines := strings.Split(text, "\n")
	line := pos.Line
	if line < 0 || line >= len(lines) {
		return 0, 0, ""
	}
	runes := []rune(lines[line])
	i := pos.Character
	if i > len(runes) {
		i = len(runes)
	}
	start := i
	for start > 0 && isNameRune(runes[start-1]) {
		start--
	}
	end := i
	for end < len(runes) && isNameRune(runes[end]) {
		end++
	}
	if start == end {
		return 0, 0, ""
	}

	off := 0
	for n := 0; n < line; n++ {
		off += len(lines[n]) + 1
	}
	byteLo := off + byteLenOfRunes(runes[:start])
	byteHi := off + byteLenOfRunes(runes[:end])
	return byteLo, byteHi, string(runes[start:end])
}

func byteLenOfRunes(rs []rune) int {
	n := 0
	for _, r := range rs {
		n += utf8Len(r)
	}
	return n
}

func (s *Server) Hover(uri string, pos Position) *Hover {
	f := s.file(uri)
	if f == nil || f.facts == nil {
		return nil
	}
	lo, hi, word := s.wordRange(f.text, pos)
	if word == "" {
		return nil
	}
	d, u := f.facts.FindWithin(lo, hi)
	if u != nil && u.Builtin != "" {
		return &Hover{Contents: MarkupContent{Kind: Markup, Value: hoverBuiltin(word, u.Builtin)}}
	}
	if d != nil {
		r := Range{
			Start: s.positionOf(f.text, d.Pos.Off),
			End:   s.positionOf(f.text, d.Pos.Off+len([]rune(d.Name))),
		}
		return &Hover{
			Contents: MarkupContent{Kind: Markup, Value: hoverDecl(word, d)},
			Range:    &r,
		}
	}
	if u != nil {

		return &Hover{Contents: MarkupContent{Kind: Markup, Value: word}}
	}

	if k := token.LookupKeyword(word); k != nil {
		doc := k.Doc
		if doc == "" {
			doc = k.Name
		}
		return &Hover{Contents: MarkupContent{Kind: Markup, Value: hoverWord(word, k.Group.String(), doc)}}
	}
	return nil
}

func hoverDecl(name string, d *analysis.Decl) string {
	if label, _, ok := declSignature(d); ok {
		return "```mihan\n" + label + "\n```\n\nfunct, declared on line " + itoa(d.Pos.Line) + "."
	}
	kind := d.Kind.String()
	if d.TypeName != "" {
		kind += " of type " + d.TypeName
	}
	line := "```mihan\n" + name + "\n```\n\n" + kind
	if d.Node != nil {
		line += "\n\nDeclared on line " + itoa(d.Pos.Line) + "."
	}
	return line
}

func hoverBuiltin(name, doc string) string {
	return "```mihan\n" + name + "\n```\n\n" + doc
}

func hoverWord(name, group, doc string) string {
	return "```mihan\n" + name + "\n```\n\n" + group + "\n\n" + doc
}

func (s *Server) Definition(uri string, pos Position) []Location {
	f := s.file(uri)
	if f == nil || f.facts == nil {
		return nil
	}
	lo, hi, word := s.wordRange(f.text, pos)
	if word == "" {
		return nil
	}
	d, _ := f.facts.FindWithin(lo, hi)
	if d == nil {

		return nil
	}
	return []Location{{
		URI: uri,
		Range: Range{
			Start: s.positionOf(f.text, d.Pos.Off),
			End:   s.positionOf(f.text, d.Pos.Off+len([]rune(d.Name))),
		},
	}}
}

func (s *Server) positionOf2(off int) Position {
	f := s.files[""]
	_ = f
	return Position{}
}

func (s *Server) DocumentSymbol(uri string) []SymbolInformation {
	f := s.file(uri)
	if f == nil || f.facts == nil {
		return nil
	}
	out := make([]SymbolInformation, 0, 16)
	for _, d := range f.facts.Symbols() {
		start := s.positionOf(f.text, d.Pos.Off)
		end := start
		end.Character += len([]rune(d.Name))
		kind := SymbolFunction
		switch d.Kind {
		case analysis.DeclClass:
			kind = SymbolClass
		case analysis.DeclEnum:
			kind = SymbolEnum
		case analysis.DeclIface:
			kind = SymbolInterface
		case analysis.DeclMethod:
			kind = SymbolMethod
		}
		out = append(out, SymbolInformation{
			Name:     d.Name,
			Kind:     kind,
			Location: Location{URI: uri, Range: Range{Start: start, End: end}},
		})
	}
	return out
}

func (s *Server) Format(uri, text string) []TextEdit {
	bag := &diag.Bag{}
	tree := parser.Parse(uri, []byte(text), bag)
	if bag.HasErrors() {

		return nil
	}
	out := format.File(tree, []byte(text))
	if out == text {
		return nil
	}
	return []TextEdit{{
		Range: Range{
			Start: Position{Line: 0, Character: 0},
			End:   s.endPosition(text),
		},
		NewText: out,
	}}
}

func (s *Server) endPosition(text string) Position {
	lines := strings.Split(text, "\n")
	return Position{Line: len(lines) - 1, Character: len([]rune(lines[len(lines)-1]))}
}

func wordAt(text string, pos Position) string {
	line := lineOf(text, pos.Line)
	runes := []rune(line)
	i := pos.Character
	if i > len(runes) {
		i = len(runes)
	}
	start := i
	for start > 0 && isNameRune(runes[start-1]) {
		start--
	}
	end := i
	for end < len(runes) && isNameRune(runes[end]) {
		end++
	}
	return string(runes[start:end])
}

func lineOf(text string, line int) string {
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return ""
	}
	return lines[line]
}

func (s *Server) posOf(text string, p Position) diag.Pos {
	lines := strings.Split(text, "\n")
	line := p.Line
	if line < 0 {
		line = 0
	}
	if line >= len(lines) {
		line = len(lines) - 1
	}
	off := 0
	for i := 0; i < line; i++ {
		off += len(lines[i]) + 1
	}
	runes := []rune(lines[line])
	col := p.Character
	if col > len(runes) {
		col = len(runes)
	}

	byteOff := off
	for i := 0; i < col; i++ {
		byteOff += utf8Len(runes[i])
	}
	return diag.Pos{Line: line + 1, Col: col + 1, Off: byteOff}
}

func utf8Len(r rune) int {
	switch {
	case r < 0x80:
		return 1
	case r < 0x800:
		return 2
	case r < 0x10000:
		return 3
	}
	return 4
}

func (s *Server) positionOf(text string, off int) Position {
	if off > len(text) {
		off = len(text)
	}
	if off < 0 {
		off = 0
	}
	head := text[:off]
	line := strings.Count(head, "\n")
	col := off
	if nl := strings.LastIndexByte(head, '\n'); nl >= 0 {
		col = off - nl - 1
	}
	return Position{Line: line, Character: utf8.RuneCountInString(head[off-col:])}
}

func rangeOf(text string, span diag.Span) Range {
	s := &Server{}
	return Range{
		Start: s.positionOf(text, span.Start.Off),
		End:   s.positionOf(text, span.End.Off),
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}
