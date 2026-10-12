package lsp

import (
	"strings"

	"mihan/internal/analysis"
	"mihan/internal/ast"
	"mihan/internal/eval"
	"mihan/internal/format"
)

type SignatureInformation struct {
	Label         string                 `json:"label"`
	Documentation *MarkupContent         `json:"documentation,omitempty"`
	Parameters    []ParameterInformation `json:"parameters,omitempty"`
}

type ParameterInformation struct {
	Label string `json:"label"`
}

type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature"`
	ActiveParameter int                    `json:"activeParameter"`
}

type SignatureHelpOptions struct {
	TriggerCharacters   []string `json:"triggerCharacters,omitempty"`
	RetriggerCharacters []string `json:"retriggerCharacters,omitempty"`
}

func funcSignature(fn *ast.FuncDecl) (string, []ParameterInformation) {
	var params []ParameterInformation
	parts := make([]string, 0, len(fn.Params))
	for _, p := range fn.Params {
		s := p.Name
		if p.TypeName != "" {
			s = p.TypeName + " " + s
		}
		if p.Default != nil {
			s += " = " + strings.TrimSpace(format.Expr(p.Default))
		}
		parts = append(parts, s)
		params = append(params, ParameterInformation{Label: s})
	}
	label := "funct " + fn.Name + "(" + strings.Join(parts, ", ") + ")"
	if fn.Result != "" {
		label += " -> " + fn.Result
	}
	return label, params
}

func declSignature(d *analysis.Decl) (string, []ParameterInformation, bool) {
	if d == nil {
		return "", nil, false
	}
	if fn, ok := d.Node.(*ast.FuncDecl); ok {
		l, p := funcSignature(fn)
		return l, p, true
	}
	return "", nil, false
}

func (s *Server) SignatureHelp(uri string, pos Position) *SignatureHelp {
	f := s.file(uri)
	if f == nil {
		return nil
	}
	line := []rune(lineOf(f.text, pos.Line))
	end := pos.Character
	if end > len(line) {
		end = len(line)
	}

	depth, commas, open := 0, 0, -1
	inText := false
	for i := end - 1; i >= 0; i-- {
		c := line[i]
		if c == '"' {
			inText = !inText
			continue
		}
		if inText {
			continue
		}
		switch c {
		case ')', ']', '}':
			depth++
		case '(', '[', '{':
			if depth == 0 {
				if c == '(' {
					open = i
				}
				i = -1
				continue
			}
			depth--
		case ',':
			if depth == 0 {
				commas++
			}
		}
		if open >= 0 {
			break
		}
	}
	if open <= 0 {
		return nil
	}
	j := open
	for j > 0 && isNameRune(line[j-1]) {
		j--
	}
	name := string(line[j:open])
	if name == "" {
		return nil
	}
	help := &SignatureHelp{ActiveParameter: commas}
	if f.facts != nil {
		scope := f.facts.ScopeAt(s.posOf(f.text, pos))
		if d := scope.Lookup(name); d != nil {
			if label, params, ok := declSignature(d); ok {
				help.Signatures = []SignatureInformation{{Label: label, Parameters: params}}
				return help
			}
		}
	}
	for _, b := range eval.Builtins() {
		if b.Name == name || b.Alias == name {
			help.Signatures = []SignatureInformation{{
				Label:         b.Name + "(…)",
				Documentation: &MarkupContent{Kind: Markup, Value: b.Doc},
			}}
			return help
		}
	}
	return nil
}
