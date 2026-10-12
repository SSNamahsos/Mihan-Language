package format

import (
	"bytes"
	"fmt"
	"strings"

	"mihan/internal/ast"
	"mihan/internal/token"
)

type Source = []byte

func Expr(e ast.Expr) string {
	var b strings.Builder
	p := &printer{b: &b, indentUnit: "    "}
	p.expr(e, 0)
	return b.String()
}

func File(f *ast.File, src []byte) string {
	var b strings.Builder
	comments := commentIndex(f.Comments)
	p := &printer{b: &b, comments: comments, indentUnit: "    "}
	first := true
	for _, st := range f.Decls {
		if !first && p.wantsBlankBefore(st, src) {
			b.WriteByte('\n')
		}
		first = false
		p.stmt(st, 0)
	}

	p.leading(1<<30, 0)
	out := b.String()
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}

	lines := strings.Split(out, "\n")
	var kept []string
	for i, l := range lines {
		if l == "" && i+1 < len(lines) && lines[i+1] == "" {
			continue
		}
		kept = append(kept, l)
	}
	return strings.Join(kept, "\n")
}

type comment struct {
	off   int
	text  string
	block bool
}

func commentIndex(toks []token.Token) []comment {
	out := make([]comment, 0, len(toks))
	for _, t := range toks {
		out = append(out, comment{off: t.Off, text: t.Lit, block: strings.HasPrefix(t.Lit, "/*")})
	}
	return out
}

type printer struct {
	b          *strings.Builder
	comments   []comment
	next       int
	indentUnit string
}

func (p *printer) pad(level int) {
	for i := 0; i < level; i++ {
		p.b.WriteString(p.indentUnit)
	}
}

func (p *printer) leading(off int, level int) {
	for p.next < len(p.comments) {
		c := p.comments[p.next]
		if c.off >= off {
			return
		}
		p.next++
		if c.block {
			p.pad(level)
			p.b.WriteString(c.text)
			p.b.WriteByte('\n')
			continue
		}
		p.pad(level)
		p.b.WriteString(c.text)
		p.b.WriteByte('\n')
	}
}

func (p *printer) wantsBlankBefore(st ast.Stmt, src []byte) bool {
	_ = src
	switch st.(type) {
	case *ast.FuncDecl, *ast.ClassDecl, *ast.StructDecl, *ast.EnumDecl, *ast.InterfaceDecl:
		return true
	}
	return false
}

func (p *printer) stmt(st ast.Stmt, level int) {

	p.leading(st.Span().Start.Off, level)
	switch s := st.(type) {
	case *ast.ExprStmt:
		p.pad(level)
		p.expr(s.X, level)
		p.b.WriteByte('\n')

	case *ast.VarDecl:
		p.pad(level)
		if s.TypeName != "" {
			p.b.WriteString(s.TypeName)
			p.b.WriteByte(' ')
		}
		p.b.WriteString(s.Name)

		if s.Value == nil {
			p.b.WriteByte('\n')
			break
		}
		p.b.WriteString(" = ")
		p.expr(s.Value, level)
		p.b.WriteByte('\n')

	case *ast.IfStmt:
		p.pad(level)
		p.b.WriteString("agar ")
		p.expr(s.Cond, level)
		p.b.WriteByte('\n')
		p.block(s.Then, level+1)
		if s.Else != nil {
			p.pad(level)
			p.b.WriteString("warna")
			if inner, ok := s.Else.(*ast.IfStmt); ok {
				p.b.WriteByte('\n')
				p.stmt(inner, level+1)
				return
			}
			if blk, ok := s.Else.(*ast.Block); ok {
				p.b.WriteByte('\n')
				p.block(blk, level+1)
				return
			}
			p.b.WriteByte('\n')
			p.stmt(s.Else, level+1)
		}

	case *ast.WhileStmt:
		p.pad(level)
		p.b.WriteString("ta ")
		p.expr(s.Cond, level)
		p.b.WriteByte('\n')
		p.block(s.Body, level+1)

	case *ast.ForStmt:
		p.pad(level)
		fmt.Fprintf(p.b, "har %s az ", s.VarName)
		p.expr(s.Lo, level)
		p.b.WriteString(" ta ")
		p.expr(s.Hi, level)
		if s.Inclusive {
			p.b.WriteByte('+')
		}
		if s.Filter != nil {
			p.b.WriteString(" ke ")
			p.expr(s.Filter, level)
		}
		p.b.WriteByte('\n')
		p.block(s.Body, level+1)

	case *ast.ForInStmt:
		p.pad(level)
		fmt.Fprintf(p.b, "har %s dar ", s.VarName)
		p.expr(s.Iter, level)
		if s.Filter != nil {
			p.b.WriteString(" ke ")
			p.expr(s.Filter, level)
		}
		p.b.WriteByte('\n')
		p.block(s.Body, level+1)

	case *ast.FuncDecl:
		p.pad(level)
		p.b.WriteString("funct ")
		p.b.WriteString(s.Name)
		p.b.WriteByte('(')
		for i, pm := range s.Params {
			if i > 0 {
				p.b.WriteString(", ")
			}
			p.b.WriteString(pm.Name)
			if pm.TypeName != "" {
				p.b.WriteString(": ")
				p.b.WriteString(pm.TypeName)
			}
			if pm.Default != nil {
				p.b.WriteString(" = ")
				p.expr(pm.Default, level)
			}
		}
		p.b.WriteByte(')')
		if s.Result != "" {
			p.b.WriteString(" -> ")
			p.b.WriteString(s.Result)
		}
		p.b.WriteByte('\n')
		p.block(s.Body, level+1)

	case *ast.ClassDecl:
		p.pad(level)
		p.b.WriteString("darghan ")
		p.b.WriteString(s.Name)
		if len(s.Bases) > 0 {
			p.b.WriteString(" < ")
			p.b.WriteString(strings.Join(s.Bases, ", "))
			p.b.WriteString(">")
		}
		p.b.WriteByte('\n')
		p.bodyBlock(s.Fields, s.Methods, level+1)

	case *ast.StructDecl:
		p.pad(level)
		p.b.WriteString("sakht ")
		p.b.WriteString(s.Name)
		p.b.WriteByte('\n')
		p.bodyBlock(s.Fields, nil, level+1)

	case *ast.ReturnStmt:
		p.pad(level)
		p.b.WriteString("bargard")
		if s.Value != nil {
			p.b.WriteByte(' ')
			p.expr(s.Value, level)
		}
		p.b.WriteByte('\n')

	case *ast.BreakStmt:
		p.pad(level)
		p.b.WriteString("shek\n")

	case *ast.ContinueStmt:
		p.pad(level)
		p.b.WriteString("edame\n")

	case *ast.Block:
		p.block(s, level)

	default:
		p.pad(level)
		p.b.WriteString(st.NodeKind())
		p.b.WriteByte('\n')
	}
}

func (p *printer) assign(s *ast.Assign, level int) {
	p.expr(s.Target, level)
	p.b.WriteByte(' ')
	p.b.WriteString(string(s.Op))
	p.b.WriteByte(' ')
	p.expr(s.Value, level)
	p.b.WriteByte('\n')
}

func (p *printer) bodyBlock(fields []*ast.FieldDecl, methods []*ast.FuncDecl, level int) {

	if len(fields) == 1 && methods == nil && fields[0].Default != nil && fields[0].TypeName == "" {
		p.pad(level - 1)
		p.b.WriteString(": ")
		p.b.WriteString(fields[0].Name)
		p.b.WriteString(" = ")
		p.expr(fields[0].Default, level)
		p.b.WriteByte('\n')
		return
	}
	for _, f := range fields {
		p.pad(level)
		p.b.WriteString(f.Name)
		if f.TypeName != "" {
			p.b.WriteString(": ")
			p.b.WriteString(f.TypeName)
		}
		if f.Default != nil {
			p.b.WriteString(" = ")
			p.expr(f.Default, level)
		}
		p.b.WriteByte('\n')
	}
	for _, m := range methods {
		p.stmt(m, level)
	}
}

func (p *printer) block(b *ast.Block, level int) {
	if b == nil || len(b.Stmts) == 0 {
		p.pad(level - 1)
		p.b.WriteString("    # hich dar ina nist\n")
		return
	}
	if b.Braced {
		p.pad(level - 1)
		p.b.WriteString("{\n")
		for _, st := range b.Stmts {
			p.stmt(st, level)
		}
		p.pad(level - 1)
		p.b.WriteString("}\n")
		return
	}
	for _, st := range b.Stmts {
		p.stmt(st, level)
	}
}

func (p *printer) expr(e ast.Expr, level int) {
	switch x := e.(type) {
	case nil:
		return
	case *ast.Ident:
		p.b.WriteString(x.Name)
	case *ast.IntLit:
		p.b.WriteString(x.Raw)
	case *ast.FloatLit:
		p.b.WriteString(x.Raw)
	case *ast.StrLit:
		p.b.WriteString(quoteString(x.Value))
	case *ast.CharLit:
		p.b.WriteString(fmt.Sprintf("'%c'", x.Value))
	case *ast.BoolLit:
		if x.Value {
			p.b.WriteString("dorost")
		} else {
			p.b.WriteString("ghalat")
		}
	case *ast.HichLit:
		p.b.WriteString("hich")
	case *ast.InterpLit:
		var b strings.Builder
		for _, seg := range x.Segs {
			if seg.Expr == nil {
				b.WriteString(seg.Text)
				continue
			}
			b.WriteByte('{')
			sub := &printer{b: &b, indentUnit: p.indentUnit}
			sub.expr(seg.Expr, 0)
			b.WriteByte('}')
		}
		p.b.WriteByte('$')
		p.b.WriteString(quoteString(b.String()))
	case *ast.ListLit:
		p.b.WriteByte('[')
		for i, el := range x.Elems {
			if i > 0 {
				p.b.WriteString(", ")
			}
			p.expr(el, level)
		}
		p.b.WriteByte(']')
	case *ast.DictLit:
		p.b.WriteByte('{')
		for i, entry := range x.Entries {
			if i > 0 {
				p.b.WriteString(", ")
			}
			p.expr(entry.Key, level)
			p.b.WriteString(": ")
			p.expr(entry.Value, level)
		}
		p.b.WriteByte('}')
	case *ast.SetLit:
		p.b.WriteByte('{')
		for i, el := range x.Elems {
			if i > 0 {
				p.b.WriteString(", ")
			}
			p.expr(el, level)
		}
		p.b.WriteByte('}')
	case *ast.Group:
		p.b.WriteByte('(')
		p.expr(x.X, level)
		p.b.WriteByte(')')
	case *ast.Unary:
		p.b.WriteString(string(x.Op))
		p.expr(x.X, level)
	case *ast.Binary:
		p.expr(x.L, level)
		p.b.WriteByte(' ')
		p.b.WriteString(string(x.Op))
		p.b.WriteByte(' ')
		p.expr(x.R, level)
	case *ast.Assign:
		p.expr(x.Target, level)
		p.b.WriteByte(' ')
		p.b.WriteString(string(x.Op))
		p.b.WriteByte(' ')
		p.expr(x.Value, level)
	case *ast.Ternary:
		p.b.WriteString("agar ")
		p.expr(x.Cond, level)
		p.b.WriteByte(' ')
		p.expr(x.Then, level)
		p.b.WriteString(" warna ")
		p.expr(x.Else, level)
	case *ast.ListComp:
		p.b.WriteByte('[')
		p.expr(x.Elem, level)
		p.b.WriteString(" har " + x.VarName)
		if r, ok := x.Iter.(*ast.RangeExpr); ok {
			p.b.WriteString(" az ")
			p.expr(r.Lo, level)
			p.b.WriteString(" ta ")
			p.expr(r.Hi, level)
			if r.Inclusive {
				p.b.WriteString(" +")
			}
		} else {
			p.b.WriteString(" dar ")
			p.expr(x.Iter, level)
		}
		if x.Filter != nil {
			p.b.WriteString(" ke ")
			p.expr(x.Filter, level)
		}
		p.b.WriteByte(']')
	case *ast.RangeExpr:
		p.expr(x.Lo, level)
		if x.Inclusive {
			p.b.WriteString("..+")
		} else {
			p.b.WriteString("..")
		}
		p.expr(x.Hi, level)
	case *ast.Index:
		p.expr(x.X, level)
		p.b.WriteByte('[')
		p.expr(x.At, level)
		p.b.WriteByte(']')
	case *ast.Slice:
		p.expr(x.X, level)
		p.b.WriteByte('[')
		if x.Lo != nil {
			p.expr(x.Lo, level)
		}
		if x.Inclusive {
			p.b.WriteString("..")
			if x.Hi != nil {
				p.expr(x.Hi, level)
			}
		} else {
			p.b.WriteString(":")
			if x.Hi != nil {
				p.expr(x.Hi, level)
			}
		}
		p.b.WriteByte(']')
	case *ast.Member:
		p.expr(x.X, level)
		p.b.WriteByte('.')
		p.b.WriteString(x.Name)
		if x.Call {
			p.b.WriteByte('(')
			for i, a := range x.Args {
				if i > 0 {
					p.b.WriteString(", ")
				}
				if a.Name != "" {
					p.b.WriteString(a.Name)
					p.b.WriteString(": ")
				}
				p.expr(a.Expr, level)
			}
			p.b.WriteByte(')')
		}
	case *ast.Call:
		p.expr(x.Fn, level)
		if x.Bare {

			if len(x.Args) > 0 {
				p.b.WriteByte(' ')
			}
			for i, a := range x.Args {
				if i > 0 {
					p.b.WriteString(", ")
				}
				if a.Name != "" {
					p.b.WriteString(a.Name)
					p.b.WriteString(": ")
				}
				p.expr(a.Expr, level)
			}
			break
		}
		p.b.WriteByte('(')
		for i, a := range x.Args {
			if i > 0 {
				p.b.WriteString(", ")
			}
			if a.Name != "" {
				p.b.WriteString(a.Name)
				p.b.WriteString(": ")
			}
			p.expr(a.Expr, level)
		}
		p.b.WriteByte(')')
	case *ast.Lambda:
		p.b.WriteString("funct(")
		for i, pm := range x.Params {
			if i > 0 {
				p.b.WriteString(", ")
			}
			p.b.WriteString(pm.Name)
		}
		p.b.WriteString(") -> ")
		if x.Expr != nil {
			p.expr(x.Expr, level)
		} else if x.Body != nil {
			p.b.WriteByte('\n')
			p.block(x.Body, level+1)
		}
	default:
		p.b.WriteString(fmt.Sprintf("<%s>", e.NodeKind()))
	}
}

func quoteString(s string) string {
	var b bytes.Buffer
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
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
