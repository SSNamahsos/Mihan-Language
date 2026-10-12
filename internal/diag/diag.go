package diag

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Severity uint8

const (
	Error Severity = iota
	Warning
	Hint
)

func (s Severity) String() string {
	switch s {
	case Error:
		return "khata"
	case Warning:
		return "ehtiyat"
	case Hint:
		return "rahnamai"
	}
	return "?"
}

type Pos struct {
	Off  int
	Line int
	Col  int
}

func (p Pos) LineString() string {
	if p.Col <= 0 {
		return strconv.Itoa(p.Line)
	}
	return strconv.Itoa(p.Line) + ":" + strconv.Itoa(p.Col)
}

func (p Pos) String() string { return p.LineString() }

type Span struct {
	Start Pos
	End   Pos
}

func (s Span) Span() Span { return s }

func NewSpan(start, end Pos) Span { return Span{Start: start, End: end} }

type Diagnostic struct {
	Severity Severity
	Code     string
	FA       string
	EN       string
	Span     Span
	File     string
	Notes    []string
}

func (d *Diagnostic) Error() string { return d.Render(nil) }

func (d *Diagnostic) Message() string {
	return fmt.Sprintf("line %d: %s / %s", d.Span.Start.Line, d.FA, d.EN)
}

func (d *Diagnostic) Render(src []byte) string {
	var b strings.Builder
	sev := "error"
	switch d.Severity {
	case Warning:
		sev = "warning"
	case Hint:
		sev = "hint"
	}
	fmt.Fprintf(&b, "%s: %s\n", sev, d.Code)
	fmt.Fprintf(&b, "  %s\n", d.FA)
	fmt.Fprintf(&b, "  %s\n", d.EN)
	if src != nil {
		b.WriteString(renderExcerpt(src, d.Span))
	}
	for _, n := range d.Notes {
		fmt.Fprintf(&b, "  note: %s\n", n)
	}
	return b.String()
}

func renderExcerpt(src []byte, span Span) string {
	line, lineStart, lineNo := lineOf(src, span.Start.Off)
	col := utf8.RuneCount(src[lineStart:span.Start.Off]) + 1
	endCol := col
	if span.End.Off > span.Start.Off && span.End.Off <= len(src) {
		endCol = utf8.RuneCount(src[lineStart:span.End.Off]) + 1
	}
	if endCol <= col {
		endCol = col + 1
	}
	gutter := fmt.Sprintf("%4d | ", lineNo)
	var b strings.Builder
	b.WriteString(gutter)
	b.WriteString(strings.TrimRight(string(line), "\r\n"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat(" ", len(gutter)+col-1))
	b.WriteString(strings.Repeat("^", endCol-col))
	return b.String()
}

func lineOf(src []byte, off int) ([]byte, int, int) {
	if off > len(src) {
		off = len(src)
	}
	start := 0
	lineNo := 1
	for i := 0; i < off; i++ {
		if src[i] == '\n' {
			lineNo++
			start = i + 1
		}
	}
	end := start
	for end < len(src) && src[end] != '\n' {
		end++
	}
	return src[start:end], start, lineNo
}

func New(code string, span Span, fa, en string) *Diagnostic {
	return &Diagnostic{Severity: Error, Code: code, FA: fa, EN: en, Span: span}
}

func Errorf(code string, span Span, fa, en string, args ...any) *Diagnostic {
	return ErrorfArgs(code, span, fa, args, en, args)
}

func ErrorfArgs(code string, span Span, fa string, faArgs []any, en string, enArgs []any) *Diagnostic {
	return &Diagnostic{
		Severity: Error,
		Code:     code,
		FA:       fmt.Sprintf(fa, faArgs...),
		EN:       fmt.Sprintf(en, enArgs...),
		Span:     span,
	}
}

func (d *Diagnostic) WithNote(format string, args ...any) *Diagnostic {
	d.Notes = append(d.Notes, fmt.Sprintf(format, args...))
	return d
}

func (d *Diagnostic) WithFile(name string) *Diagnostic {
	d.File = name
	return d
}

type Bag struct {
	items []*Diagnostic
}

func (b *Bag) Add(d *Diagnostic) {
	if d != nil {
		b.items = append(b.items, d)
	}
}

func (b *Bag) AddAll(ds []*Diagnostic) {
	for _, d := range ds {
		b.Add(d)
	}
}

func (b *Bag) HasErrors() bool {
	for _, d := range b.items {
		if d.Severity == Error {
			return true
		}
	}
	return false
}

func (b *Bag) Len() int { return len(b.items) }

func (b *Bag) Sorted() []*Diagnostic {
	out := make([]*Diagnostic, len(b.items))
	copy(out, b.items)
	sort.SliceStable(out, func(i, j int) bool {
		a, c := out[i], out[j]
		if a.File != c.File {
			return a.File < c.File
		}
		if a.Span.Start.Line != c.Span.Start.Line {
			return a.Span.Start.Line < c.Span.Start.Line
		}
		return a.Span.Start.Col < c.Span.Start.Col
	})
	return out
}

func (b *Bag) Merge(o *Bag) {
	if o != nil {
		b.items = append(b.items, o.items...)
	}
}

func (b *Bag) Reset() { b.items = nil }
