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
	text := f.text
	off := s.posOf(text, pos).Off
	from := off - 4000
	if from < 0 {
		from = 0
	}
	if nl := strings.LastIndexByte(text[:from], '\n'); nl >= 0 && from > 0 {
		from = nl + 1
	}
	type open struct {
		at, commas int
		paren      bool
	}
	var stack []open
	scanCode(text[from:], off-from, func(j int) {
		switch text[from+j] {
		case '(', '[', '{':
			stack = append(stack, open{at: from + j, paren: text[from+j] == '('})
		case ')', ']', '}':
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case ',':
			if len(stack) > 0 {
				stack[len(stack)-1].commas++
			}
		}
	})
	for len(stack) > 0 && !stack[len(stack)-1].paren {
		stack = stack[:len(stack)-1]
	}
	if len(stack) == 0 {
		return nil
	}
	top := stack[len(stack)-1]
	head := []rune(text[:top.at])
	if nl := strings.LastIndexByte(text[:top.at], '\n'); nl >= 0 {
		head = []rune(text[nl+1 : top.at])
	}
	callee := receiverBefore(head, len(head))
	if callee == "" || !isNameRune([]rune(callee)[len([]rune(callee))-1]) {
		return nil
	}
	facts, apos := s.analyze(uri, text, pos)
	c := &cctx{s: s, uri: uri, text: text, off: off, facts: facts, lines: strings.Split(text, "\n"), line: pos.Line}
	if facts != nil {
		c.scope = facts.ScopeAt(apos)
	}
	sig := c.signatureOf(callee)
	if sig == nil {
		return nil
	}
	active := top.commas
	if n := len(sig.Parameters); n > 0 && active >= n {
		active = n - 1
	}
	return &SignatureHelp{Signatures: []SignatureInformation{*sig}, ActiveParameter: active}
}

func paramInfos(list []string) []ParameterInformation {
	out := make([]ParameterInformation, 0, len(list))
	for _, p := range list {
		out = append(out, ParameterInformation{Label: p})
	}
	return out
}

func (c *cctx) signatureOf(callee string) *SignatureInformation {
	if dot := strings.LastIndex(callee, "."); dot > 0 {
		recv, name := callee[:dot], callee[dot+1:]
		t := c.typeOfExpr(recv)
		for _, m := range methodTable(t.kind) {
			if m.name == name || m.alias == name {
				var ps []string
				for _, p := range strings.Split(m.params, ",") {
					if p = strings.TrimSpace(p); p != "" {
						ps = append(ps, p)
					}
				}
				label := name + "(" + strings.Join(ps, ", ") + ")"
				if m.result != "" && m.result != "any" {
					label += " → " + m.result
				}
				return &SignatureInformation{Label: label, Parameters: paramInfos(ps), Documentation: &MarkupContent{Kind: Markup, Value: "**" + m.fa + "**\n\n" + m.en}}
			}
		}
		if t.kind == tClass || t.kind == tStatic {
			for _, cl := range c.classChain(t.name) {
				for _, m := range cl.methods {
					if m.Name == name {
						l, p := funcSignature(m)
						return &SignatureInformation{Label: strings.TrimPrefix(l, "funct "), Parameters: p, Documentation: &MarkupContent{Kind: Markup, Value: "method of `" + cl.name + "`" + docSuffix(m.Doc)}}
					}
				}
			}
		}
		if t.kind == tModule {
			if mod := c.s.module(c.uri, t.name); mod != nil {
				for _, fn := range mod.funcs {
					if fn.Name == name {
						l, p := funcSignature(fn)
						return &SignatureInformation{Label: l, Parameters: p, Documentation: &MarkupContent{Kind: Markup, Value: "from `" + t.name + "`" + docSuffix(fn.Doc)}}
					}
				}
			}
		}
		return nil
	}
	if cl := c.classDecl(callee); cl != nil {
		for _, chain := range c.classChain(callee) {
			for _, m := range chain.methods {
				if m.Name == "__sakht__" || m.Name == chain.name {
					_, p := funcSignature(m)
					parts := make([]string, 0, len(p))
					for _, x := range p {
						parts = append(parts, x.Label)
					}
					return &SignatureInformation{Label: callee + "(" + strings.Join(parts, ", ") + ")", Parameters: p, Documentation: &MarkupContent{Kind: Markup, Value: "**سازنده‌ی " + callee + "** · constructor"}}
				}
			}
		}
		return &SignatureInformation{Label: callee + "()", Documentation: &MarkupContent{Kind: Markup, Value: "**سازنده‌ی " + callee + "** · constructor"}}
	}
	if c.scope != nil {
		if d := c.scope.Lookup(callee); d != nil {
			if label, params, ok := declSignature(d); ok {
				doc := ""
				if fn, ok := d.Node.(*ast.FuncDecl); ok {
					doc = strings.TrimSpace(fn.Doc)
				}
				si := &SignatureInformation{Label: label, Parameters: params}
				if doc != "" {
					si.Documentation = &MarkupContent{Kind: Markup, Value: doc}
				}
				return si
			}
		}
	}
	if fn := c.funcDecl(callee); fn != nil {
		l, p := funcSignature(fn)
		return &SignatureInformation{Label: l, Parameters: p}
	}
	for _, b := range eval.Builtins() {
		if b.Name == callee || b.Alias == callee {
			_, ps := builtinSig(b)
			label := callee + "(" + strings.Join(ps, ", ") + ")"
			if b.Result != "" && b.Result != "any" {
				label += " → " + b.Result
			}
			return &SignatureInformation{Label: label, Parameters: paramInfos(ps), Documentation: &MarkupContent{Kind: Markup, Value: builtinDoc(b, callee)}}
		}
	}
	for _, im := range c.imports() {
		if im.alias != "" {
			continue
		}
		if mod := c.s.module(c.uri, im.path); mod != nil {
			for _, fn := range mod.funcs {
				if fn.Name == callee {
					l, p := funcSignature(fn)
					return &SignatureInformation{Label: l, Parameters: p, Documentation: &MarkupContent{Kind: Markup, Value: "from `" + im.path + "`"}}
				}
			}
		}
	}
	return nil
}
