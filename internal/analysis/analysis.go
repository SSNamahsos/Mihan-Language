package analysis

import (
	"sort"

	"mihan/internal/ast"
	"mihan/internal/diag"
)

type Decl struct {
	Name string

	Node ast.Node

	Pos diag.Pos

	Lo, Hi int

	Kind DeclKind

	TypeName string

	MadeBy string

	Scope *Scope
}

type DeclKind uint8

const (
	DeclVar DeclKind = iota

	DeclParam

	DeclFunc

	DeclClass

	DeclEnum

	DeclIface

	DeclCatch

	DeclMethod

	DeclFor

	DeclForIn
)

func (k DeclKind) String() string {
	switch k {
	case DeclVar:
		return "variable"
	case DeclParam:
		return "parameter"
	case DeclFunc:
		return "funct"
	case DeclClass:
		return "darghan"
	case DeclEnum:
		return "shenase"
	case DeclIface:
		return "dastoor"
	case DeclCatch:
		return "giroft name"
	case DeclMethod:
		return "method"
	case DeclFor:
		return "har variable"
	case DeclForIn:
		return "har variable"
	}
	return "name"
}

type Scope struct {
	Parent *Scope

	Name string

	Span diag.Span

	Decls []*Decl

	byName map[string][]*Decl

	Body *ast.Block

	children []*Scope
}

func (s *Scope) addChild(c *Scope) { s.children = append(s.children, c) }

func NewScope(parent *Scope, name string, span diag.Span) *Scope {
	return &Scope{Parent: parent, Name: name, Span: span, byName: map[string][]*Decl{}}
}

func (s *Scope) Add(d *Decl) {
	d.Scope = s
	s.Decls = append(s.Decls, d)
	s.byName[d.Name] = append(s.byName[d.Name], d)
}

func (s *Scope) Lookup(name string) *Decl {
	for sc := s; sc != nil; sc = sc.Parent {
		list := sc.byName[name]
		for i := len(list) - 1; i >= 0; i-- {
			return list[i]
		}
	}
	return nil
}

func (s *Scope) Names() []string {
	seen := map[string]bool{}
	var out []string
	for sc := s; sc != nil; sc = sc.Parent {

		for _, d := range sc.Decls {
			if !seen[d.Name] {
				seen[d.Name] = true
				out = append(out, d.Name)
			}
		}
	}
	return out
}

var Builtins = map[string]string{}

func IsBuiltin(name string) (string, bool) {
	doc, ok := Builtins[name]
	return doc, ok
}

type File struct {
	Tree *ast.File

	Root *Scope

	Uses []*Use

	Decls []*Decl
}

type Use struct {
	Name string

	Node ast.Expr

	Pos diag.Pos

	Lo, Hi int

	Scope *Scope

	Decl *Decl

	Builtin string
}

func Build(f *ast.File) *File {
	af := &File{Tree: f, Root: NewScope(nil, "", f.Span())}
	af.Root.Body = &ast.Block{Loc: f.Span()}
	for _, d := range f.Decls {
		af.declTop(d, af.Root)
	}
	af.resolve()
	return af
}

func (a *File) resolve() {
	for _, u := range a.Uses {
		if d := u.Scope.Lookup(u.Name); d != nil {
			u.Decl = d
			continue
		}
		if doc, ok := IsBuiltin(u.Name); ok {
			u.Builtin = doc
		}
	}
}

func (a *File) ScopeAt(pos diag.Pos) *Scope {
	best := a.Root
	bestSpan := a.Root.Span

	var walk func(s *Scope)
	walk = func(s *Scope) {
		if !containsPos(s.Span, pos) {
			return
		}
		best = s
		bestSpan = s.Span
		_ = bestSpan
		for _, d := range s.Decls {
			_ = d
		}
		for _, c := range s.children {
			walk(c)
		}
	}
	walk(a.Root)
	return best
}

func containsPos(span diag.Span, pos diag.Pos) bool {
	if pos.Off < span.Start.Off {
		return false
	}
	return pos.Off <= span.End.Off
}

func (a *File) DeclAt(pos diag.Pos) *Decl {
	for _, d := range a.Decls {
		if d.Pos.Off == pos.Off {
			return d
		}
	}
	return nil
}

func (a *File) UseAt(pos diag.Pos) *Use {
	for _, u := range a.Uses {
		if u.Pos.Off == pos.Off {
			return u
		}
	}
	return nil
}

func (a *File) Find(pos diag.Pos) (*Decl, *Use) {
	if d := a.DeclAt(pos); d != nil {
		return d, nil
	}
	if u := a.UseAt(pos); u != nil {
		return u.Decl, u
	}
	return nil, nil
}

func (a *File) FindWithin(lo, hi int) (*Decl, *Use) {

	for _, u := range a.Uses {
		if lo >= u.Lo && lo <= u.Hi {
			return u.Decl, u
		}
	}
	for _, d := range a.Decls {
		if lo >= d.Lo && lo <= d.Hi {
			return d, nil
		}
	}
	return nil, nil
}

func (a *File) DescribesName(pos diag.Pos) (name string, decl *Decl, builtin string) {
	if d, u := a.Find(pos); u != nil {
		return u.Name, d, u.Builtin
	} else if d != nil {
		return d.Name, d, ""
	}

	return "", nil, ""
}

func (a *File) Symbols() []*Decl {
	var out []*Decl
	for _, d := range a.Decls {
		switch d.Kind {
		case DeclFunc, DeclClass, DeclEnum, DeclIface, DeclMethod:
			out = append(out, d)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Pos.Off < out[j].Pos.Off })
	return out
}
