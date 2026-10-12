package token

import (
	"fmt"
	"unicode/utf8"

	"mihan/internal/diag"
)

func (t Token) Start() diag.Pos { return diag.Pos{Off: t.Off, Line: t.Line, Col: t.Col} }

func (t Token) EndPos() diag.Pos {
	return diag.Pos{Off: t.End, Line: t.Line, Col: t.Col + utf8.RuneCountInString(t.Lit)}
}

func (t Token) Span() diag.Span { return diag.Span{Start: t.Start(), End: t.EndPos()} }

type Kind string

const (
	ILLEGAL Kind = "illegal"
	EOF     Kind = "eof"
	COMMENT Kind = "comment"

	NEWLINE Kind = "newline"
	INDENT  Kind = "indent"
	DEDENT  Kind = "dedent"

	IDENT  Kind = "ident"
	INT    Kind = "int"
	FLOAT  Kind = "float"
	STRING Kind = "string"
	CHAR   Kind = "char"

	INTERP      Kind = "interp"
	INTERP_TEXT Kind = "interp_text"
	INTERP_END  Kind = "interp_end"
	STRING_END  Kind = "string_end"
)

func (k Kind) IsKeyword() bool { return isKeywordKind[k] }

func (k Kind) IsLiteral() bool {
	switch k {
	case INT, FLOAT, STRING, CHAR:
		return true
	}
	return false
}

type Token struct {
	Kind Kind
	Lit  string
	Text string
	Line int
	Col  int
	Pos  int
	Off  int
	End  int
}

func (t Token) String() string {
	if t.Lit != "" {
		return t.Lit
	}
	return string(t.Kind)
}

func (t Token) DescribeFarsi() string {
	switch t.Kind {
	case EOF:
		return "پایان پرونده"
	case NEWLINE:
		return "پایان سطر"
	case INDENT:
		return "بلوکِ تورفتگی"
	case DEDENT:
		return "پایانِ بلوکِ تورفتگی"
	case ILLEGAL:
		if t.Text != "" {
			return fmt.Sprintf("«%s»", t.Text)
		}
		return "نشانه‌ای که معنی ندارد"
	}
	if t.Kind.IsKeyword() {
		return fmt.Sprintf("کلیدواژهٔ «%s»", t.Lit)
	}
	switch t.Kind {
	case IDENT:
		return fmt.Sprintf("نامِ «%s»", t.Lit)
	case INT, FLOAT:
		return fmt.Sprintf("عددِ %s", t.Lit)
	case STRING:
		return "متن"
	}
	if t.Text != "" {
		return fmt.Sprintf("«%s»", t.Text)
	}
	return string(t.Kind)
}

func (t Token) Describe() string {
	switch t.Kind {
	case EOF:
		return "end of file"
	case NEWLINE:
		return "end of line"
	case INDENT:
		return "an indented block"
	case DEDENT:
		return "the end of an indented block"
	case ILLEGAL:
		if t.Text != "" {
			return fmt.Sprintf("%q", t.Text)
		}
		return "an illegal token"
	}
	if t.Kind.IsKeyword() {
		return fmt.Sprintf("keyword %q", t.Lit)
	}
	switch t.Kind {
	case IDENT:
		return fmt.Sprintf("name %q", t.Lit)
	case INT, FLOAT:
		return fmt.Sprintf("number %s", t.Text)
	case STRING, CHAR:
		return "a string"
	}
	if t.Kind.IsOperator() || t.Kind.IsPunct() {
		return fmt.Sprintf("%q", t.Lit)
	}
	return string(t.Kind)
}

func (t Token) Is(k Kind) bool { return t.Kind == k }
