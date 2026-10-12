package ast

import (
	"sync/atomic"

	"mihan/internal/diag"
	"mihan/internal/token"
)

type Node interface {
	Span() diag.Span

	NodeKind() string
}

type Loc = diag.Span

func SpanOf(a, b diag.Pos) Loc { return diag.Span{Start: a, End: b} }

type Expr interface {
	Node
	ExprNode()
}

type Ident struct {
	Loc
	Name string

	FileScope bool
	slot      atomic.Pointer[Slot]
}

type Slot struct{ Scope, Binding any }

func (x *Ident) LoadSlot() *Slot { return x.slot.Load() }

func (x *Ident) StoreSlot(s *Slot) { x.slot.Store(s) }

func (*Ident) ExprNode()        {}
func (*Ident) NodeKind() string { return "Ident" }

type IntLit struct {
	Loc
	Value int64
	Raw   string
}

func (*IntLit) ExprNode()        {}
func (*IntLit) NodeKind() string { return "IntLit" }

type FloatLit struct {
	Loc
	Value float64
	Raw   string
}

func (*FloatLit) ExprNode()        {}
func (*FloatLit) NodeKind() string { return "FloatLit" }

type StrLit struct {
	Loc
	Value string
}

func (*StrLit) ExprNode()        {}
func (*StrLit) NodeKind() string { return "StrLit" }

type CharLit struct {
	Loc
	Value rune
}

func (*CharLit) ExprNode()        {}
func (*CharLit) NodeKind() string { return "CharLit" }

type BoolLit struct {
	Loc
	Value bool
}

func (*BoolLit) ExprNode()        {}
func (*BoolLit) NodeKind() string { return "BoolLit" }

type HichLit struct{ Loc }

func (*HichLit) ExprNode()        {}
func (*HichLit) NodeKind() string { return "HichLit" }

type InterpSeg struct {
	Loc

	Text string
	Expr Expr
}

type InterpLit struct {
	Loc
	Segs []InterpSeg
}

func (*InterpLit) ExprNode()        {}
func (*InterpLit) NodeKind() string { return "InterpLit" }

type ListLit struct {
	Loc
	Elems []Expr
}

func (*ListLit) ExprNode()        {}
func (*ListLit) NodeKind() string { return "ListLit" }

type DictEntry struct {
	Key   Expr
	Value Expr
}

type DictLit struct {
	Loc
	Entries []DictEntry
}

func (*DictLit) ExprNode()        {}
func (*DictLit) NodeKind() string { return "DictLit" }

type SetLit struct {
	Loc
	Elems []Expr
}

func (*SetLit) ExprNode()        {}
func (*SetLit) NodeKind() string { return "SetLit" }

type Unary struct {
	Loc
	Op token.Kind
	X  Expr
}

func (*Unary) ExprNode()        {}
func (*Unary) NodeKind() string { return "Unary" }

type Binary struct {
	Loc
	Op    token.Kind
	OpPos diag.Pos
	L, R  Expr
}

func (*Binary) ExprNode()        {}
func (*Binary) NodeKind() string { return "Binary" }

type Assign struct {
	Loc
	Op     token.Kind
	Target Expr
	Value  Expr
}

func (*Assign) ExprNode()        {}
func (*Assign) NodeKind() string { return "Assign" }

type Ternary struct {
	Loc
	Cond Expr
	Then Expr
	Else Expr
}

func (*Ternary) ExprNode()        {}
func (*Ternary) NodeKind() string { return "Ternary" }

type Arg struct {
	Name string
	Expr Expr
}

type Call struct {
	Loc
	Fn   Expr
	Args []Arg

	Bare bool

	Spread bool
}

func (*Call) ExprNode()        {}
func (*Call) NodeKind() string { return "Call" }

type Index struct {
	Loc
	X  Expr
	At Expr
}

func (*Index) ExprNode()        {}
func (*Index) NodeKind() string { return "Index" }

type Slice struct {
	Loc
	X Expr

	Lo, Hi Expr

	Inclusive bool
}

func (*Slice) ExprNode()        {}
func (*Slice) NodeKind() string { return "Slice" }

type Member struct {
	Loc
	X    Expr
	Name string

	NamePos diag.Pos
	Call    bool
	Args    []Arg
	Spread  bool
}

func (*Member) ExprNode()        {}
func (*Member) NodeKind() string { return "Member" }

type Lambda struct {
	Loc
	Params []*Param
	Body   *Block
	Expr   Expr
}

func (*Lambda) ExprNode()        {}
func (*Lambda) NodeKind() string { return "Lambda" }

type ListComp struct {
	Loc
	Elem    Expr
	VarName string
	VarPos  diag.Pos
	Iter    Expr
	Filter  Expr
}

func (*ListComp) ExprNode()        {}
func (*ListComp) NodeKind() string { return "ListComp" }

type RangeExpr struct {
	Loc
	Lo, Hi    Expr
	Inclusive bool
}

func (*RangeExpr) ExprNode()        {}
func (*RangeExpr) NodeKind() string { return "RangeExpr" }

type Group struct {
	Loc
	X Expr
}

func (*Group) ExprNode()        {}
func (*Group) NodeKind() string { return "Group" }

type ErrorExpr struct {
	Loc

	Message string

	English string
}

func (*ErrorExpr) ExprNode()        {}
func (*ErrorExpr) NodeKind() string { return "ErrorExpr" }

type Stmt interface {
	Node
	StmtNode()
}

type Block struct {
	Loc
	Stmts []Stmt

	Braced bool
}

func (*Block) StmtNode()        {}
func (*Block) NodeKind() string { return "Block" }

type ExprStmt struct {
	Loc
	X Expr
}

func (*ExprStmt) StmtNode()        {}
func (*ExprStmt) NodeKind() string { return "ExprStmt" }

type Param struct {
	Name     string
	NamePos  diag.Pos
	TypeName string
	Default  Expr
}

type VarDecl struct {
	Loc
	TypeName string
	Name     string
	NamePos  diag.Pos
	Value    Expr
	Sabit    bool
}

func (*VarDecl) StmtNode()        {}
func (*VarDecl) NodeKind() string { return "VarDecl" }

type IfStmt struct {
	Loc
	Cond Expr
	Then *Block
	Else Stmt
}

func (*IfStmt) StmtNode()        {}
func (*IfStmt) NodeKind() string { return "IfStmt" }

type WhileStmt struct {
	Loc
	Cond Expr
	Body *Block
}

func (*WhileStmt) StmtNode()        {}
func (*WhileStmt) NodeKind() string { return "WhileStmt" }

type ForStmt struct {
	Loc
	VarName string
	VarPos  diag.Pos
	Lo, Hi  Expr

	Inclusive bool

	Filter Expr
	Body   *Block
}

func (*ForStmt) StmtNode()        {}
func (*ForStmt) NodeKind() string { return "ForStmt" }

type ForInStmt struct {
	Loc
	VarName string
	VarPos  diag.Pos
	Iter    Expr
	Filter  Expr
	Body    *Block
}

func (*ForInStmt) StmtNode()        {}
func (*ForInStmt) NodeKind() string { return "ForInStmt" }

type FuncDecl struct {
	Loc
	Name    string
	NamePos diag.Pos
	Params  []*Param

	Result string
	Body   *Block

	Generics []string

	Baza, Mahram bool
	Osta         bool
	Doc          string
}

func (*FuncDecl) StmtNode()        {}
func (*FuncDecl) NodeKind() string { return "FuncDecl" }

type ReturnStmt struct {
	Loc
	Value Expr
}

func (*ReturnStmt) StmtNode()        {}
func (*ReturnStmt) NodeKind() string { return "ReturnStmt" }

type BreakStmt struct{ Loc }

func (*BreakStmt) StmtNode()        {}
func (*BreakStmt) NodeKind() string { return "BreakStmt" }

type ContinueStmt struct{ Loc }

func (*ContinueStmt) StmtNode()        {}
func (*ContinueStmt) NodeKind() string { return "ContinueStmt" }

type TalignStmt struct {
	Loc
	X Expr
}

func (*TalignStmt) StmtNode()        {}
func (*TalignStmt) NodeKind() string { return "TalignStmt" }

type TryStmt struct {
	Loc
	Body     *Block
	Catch    *Catch
	Finally  *Block
	HasCatch bool
}

type Catch struct {
	Name     string
	NamePos  diag.Pos
	TypeName string
	Body     *Block
}

func (*TryStmt) StmtNode()        {}
func (*TryStmt) NodeKind() string { return "TryStmt" }

type ThrowStmt struct {
	Loc
	Value Expr
}

func (*ThrowStmt) StmtNode()        {}
func (*ThrowStmt) NodeKind() string { return "ThrowStmt" }

type ClassDecl struct {
	Loc
	Name         string
	NamePos      diag.Pos
	Bases        []string
	Fields       []*FieldDecl
	Methods      []*FuncDecl
	Doc          string
	Baza, Mahram bool

	CtorName string
}

func (*ClassDecl) StmtNode()        {}
func (*ClassDecl) NodeKind() string { return "ClassDecl" }

type FieldDecl struct {
	Loc
	Name         string
	NamePos      diag.Pos
	TypeName     string
	Default      Expr
	Baza, Mahram bool
	Osta         bool
	Sabit        bool
}

func (*FieldDecl) StmtNode()        {}
func (*FieldDecl) NodeKind() string { return "FieldDecl" }

type StructDecl struct {
	Loc
	Name    string
	NamePos diag.Pos

	Bases  []string
	Fields []*FieldDecl

	Methods      []*FuncDecl
	Doc          string
	Baza, Mahram bool
}

func (*StructDecl) StmtNode()        {}
func (*StructDecl) NodeKind() string { return "StructDecl" }

type EnumDecl struct {
	Loc
	Name    string
	NamePos diag.Pos
	Values  []*EnumValue
	Doc     string
}

func (*EnumDecl) StmtNode()        {}
func (*EnumDecl) NodeKind() string { return "EnumDecl" }

type EnumValue struct {
	Loc
	Name    string
	NamePos diag.Pos
	Value   Expr
}

func (*EnumValue) StmtNode()        {}
func (*EnumValue) NodeKind() string { return "EnumValue" }

type InterfaceDecl struct {
	Loc
	Name    string
	NamePos diag.Pos
	Methods []string
	Doc     string
}

func (*InterfaceDecl) StmtNode()        {}
func (*InterfaceDecl) NodeKind() string { return "InterfaceDecl" }

type ImportDecl struct {
	Loc
	Path  string
	Alias string
}

func (*ImportDecl) StmtNode()        {}
func (*ImportDecl) NodeKind() string { return "ImportDecl" }

type ModuleDecl struct {
	Loc
	Path string
}

func (*ModuleDecl) StmtNode()        {}
func (*ModuleDecl) NodeKind() string { return "ModuleDecl" }

type ProgramDecl struct {
	Loc
	Name string
}

func (*ProgramDecl) StmtNode()        {}
func (*ProgramDecl) NodeKind() string { return "ProgramDecl" }

type File struct {
	Path     string
	Source   []byte
	Decls    []Stmt
	Module   string
	Program  string
	Imports  []*ImportDecl
	Comments []token.Token
}

func (f *File) Span() diag.Span {
	if len(f.Decls) == 0 {
		return diag.Span{}
	}
	s := f.Decls[0].Span()
	e := f.Decls[len(f.Decls)-1].Span()
	return diag.Span{Start: s.Start, End: e.End}
}

func (f *File) Start() diag.Pos  { return f.Span().Start }
func (f *File) NodeKind() string { return "File" }
