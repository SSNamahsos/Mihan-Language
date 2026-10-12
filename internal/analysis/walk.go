package analysis

import (
	"mihan/internal/ast"
	"mihan/internal/diag"
)

func (a *File) declTop(d ast.Stmt, s *Scope) {
	switch x := d.(type) {
	case *ast.FuncDecl:
		a.add(s, x.Name, x.NamePos, DeclFunc, "", d)

		body := NewScope(s, x.Name, spanOf(x.Body, x.Span()))
		body.Body = x.Body
		s.addChild(body)
		a.addParams(body, x.Params)
		a.block(x.Body, body)
	case *ast.ClassDecl:
		a.add(s, x.Name, x.NamePos, DeclClass, "", d)

		for _, m := range x.Methods {
			md := &Decl{
				Name: m.Name, Node: m, Pos: m.NamePos, Kind: DeclMethod, Scope: s,
				Lo: m.NamePos.Off, Hi: m.NamePos.Off + len([]rune(m.Name)),
			}
			a.Decls = append(a.Decls, md)
		}
	case *ast.StructDecl:
		a.add(s, x.Name, x.NamePos, DeclClass, "", d)
	case *ast.EnumDecl:
		a.add(s, x.Name, x.NamePos, DeclEnum, "", d)
		for _, v := range x.Values {
			a.add(s, v.Name, v.NamePos, DeclVar, "", &ast.VarDecl{Loc: v.Span(), Name: v.Name, NamePos: v.NamePos})
		}
	case *ast.InterfaceDecl:
		a.add(s, x.Name, x.NamePos, DeclIface, "", d)
	}
	a.stmtExprs(d, s)
}

func (a *File) stmtExprs(d ast.Stmt, s *Scope) {
	switch x := d.(type) {
	case *ast.ExprStmt:
		if as, ok := x.X.(*ast.Assign); ok {
			a.assign(as, s)
			return
		}
		a.expr(x.X, s)
	case *ast.VarDecl:

		a.expr(x.Value, s)
		a.decl(s, x.Name, x.NamePos, DeclVar, x.TypeName, x)
	case *ast.ReturnStmt:
		a.expr(x.Value, s)
	case *ast.ThrowStmt:
		a.expr(x.Value, s)
	case *ast.TalignStmt:
		a.expr(x.X, s)
	case *ast.Block:
		a.block(x, s)
	case *ast.IfStmt:
		a.expr(x.Cond, s)
		a.block(x.Then, NewChild(s, "agar", spanOf(x.Then, x.Span())))
		if x.Else != nil {
			a.stmtExprs(x.Else, s)
		}
	case *ast.WhileStmt:
		a.expr(x.Cond, s)
		a.block(x.Body, NewChild(s, "ta", spanOf(x.Body, x.Span())))
	case *ast.ForStmt:
		a.expr(x.Lo, s)
		a.expr(x.Hi, s)
		inner := NewChild(s, "har", spanOf(x.Body, x.Span()))
		a.decl(inner, x.VarName, x.VarPos, DeclFor, "", x)
		a.expr(x.Filter, inner)
		a.block(x.Body, inner)
	case *ast.ForInStmt:
		a.expr(x.Iter, s)
		inner := NewChild(s, "har", spanOf(x.Body, x.Span()))
		a.decl(inner, x.VarName, x.VarPos, DeclForIn, "", x)
		a.expr(x.Filter, inner)
		a.block(x.Body, inner)
	case *ast.TryStmt:
		a.block(x.Body, NewChild(s, "koshish", spanOf(x.Body, x.Span())))
		if x.Catch != nil {
			inner := NewChild(s, "giroft", spanOf(x.Catch.Body, x.Catch.Body.Span()))
			a.decl(inner, x.Catch.Name, x.Catch.NamePos, DeclCatch, x.Catch.TypeName, x.Catch.Body)
			a.block(x.Catch.Body, inner)
		}
		if x.Finally != nil {
			a.block(x.Finally, NewChild(s, "nemigir", x.Finally.Span()))
		}
	}
}

func (a *File) assign(as *ast.Assign, s *Scope) {
	a.expr(as.Value, s)
	switch t := as.Target.(type) {
	case *ast.Ident:
		s.Add(&Decl{Name: t.Name, Node: as, Pos: t.Span().Start, Kind: DeclVar})
	case *ast.Member:
		a.expr(t.X, s)
	case *ast.Index:
		a.expr(t.X, s)
		a.expr(t.At, s)
	default:
		a.expr(as.Target, s)
	}
}

func (a *File) block(b *ast.Block, s *Scope) {
	if b == nil {
		return
	}
	for _, st := range b.Stmts {
		a.stmtExprs(st, s)
	}
}

func (a *File) addParams(s *Scope, params []*ast.Param) {
	for _, p := range params {
		a.decl(s, p.Name, p.NamePos, DeclParam, p.TypeName, nil)
	}
}

func (a *File) decl(s *Scope, name string, pos diag.Pos, kind DeclKind, typeName string, node ast.Node) {
	d := &Decl{
		Name: name, Node: node, Pos: pos, Kind: kind, TypeName: typeName,
		Lo: pos.Off, Hi: pos.Off + len([]rune(name)),
	}
	s.Add(d)
	a.Decls = append(a.Decls, d)
}

func (a *File) add(s *Scope, name string, pos diag.Pos, kind DeclKind, typeName string, node ast.Node) {
	a.decl(s, name, pos, kind, typeName, node)
}

func madeBy(e ast.Expr) string {
	c, ok := e.(*ast.Call)
	if !ok {
		return ""
	}
	if id, ok := c.Fn.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

func (a *File) use(name string, node ast.Expr, pos diag.Pos, s *Scope) {
	a.Uses = append(a.Uses, &Use{
		Name: name, Node: node, Pos: pos, Scope: s,
		Lo: pos.Off, Hi: pos.Off + len([]rune(name)),
	})
}

func (a *File) expr(e ast.Expr, s *Scope) {
	switch x := e.(type) {
	case nil:
		return
	case *ast.Ident:
		a.use(x.Name, x, x.Span().Start, s)
	case *ast.Group:
		a.expr(x.X, s)
	case *ast.Unary:
		a.expr(x.X, s)
	case *ast.Binary:
		a.expr(x.L, s)
		a.expr(x.R, s)
	case *ast.Ternary:
		a.expr(x.Cond, s)
		a.expr(x.Then, s)
		a.expr(x.Else, s)
	case *ast.ListLit:
		for _, it := range x.Elems {
			a.expr(it, s)
		}
	case *ast.SetLit:
		for _, it := range x.Elems {
			a.expr(it, s)
		}
	case *ast.DictLit:
		for _, en := range x.Entries {
			a.expr(en.Key, s)
			a.expr(en.Value, s)
		}
	case *ast.RangeExpr:
		a.expr(x.Lo, s)
		a.expr(x.Hi, s)
	case *ast.InterpLit:
		for _, seg := range x.Segs {
			a.expr(seg.Expr, s)
		}
	case *ast.Index:
		a.expr(x.X, s)
		a.expr(x.At, s)
	case *ast.Slice:
		a.expr(x.X, s)
		a.expr(x.Lo, s)
		a.expr(x.Hi, s)
	case *ast.Call:
		a.call(x.Fn, s, x.Loc)
		for _, arg := range x.Args {
			a.expr(arg.Expr, s)
		}
	case *ast.Member:
		a.expr(x.X, s)
		for _, arg := range x.Args {
			a.expr(arg.Expr, s)
		}
	case *ast.ListComp:
		a.expr(x.Iter, s)
		inner := NewChild(s, "har", x.Span())
		a.decl(inner, x.VarName, x.VarPos, DeclForIn, "", x)
		a.expr(x.Filter, inner)
		a.expr(x.Elem, inner)
	case *ast.Lambda:
		inner := NewChild(s, "lambda", x.Span())
		inner.Body = x.Body
		a.addParams(inner, x.Params)
		a.block(x.Body, inner)
		a.expr(x.Expr, inner)
	case *ast.Assign:
		a.assign(x, s)
	}
}

func (a *File) call(fn ast.Expr, s *Scope, span diag.Span) {
	if id, ok := fn.(*ast.Ident); ok {
		a.use(id.Name, id, span.Start, s)
		return
	}
	a.expr(fn, s)
}

func spanOf(b *ast.Block, fallback diag.Span) diag.Span {
	if b != nil {
		return b.Span()
	}
	return fallback
}

func NewChild(parent *Scope, name string, span diag.Span) *Scope {
	s := NewScope(parent, name, span)
	parent.addChild(s)
	return s
}
