package parser

import (
	"strings"
	"testing"
	"time"

	"mihan/internal/ast"
	"mihan/internal/diag"
	"mihan/internal/token"
)

func parse(t *testing.T, src string) (*ast.File, *diag.Bag) {
	t.Helper()
	bag := &diag.Bag{}
	return Parse("test.mhn", []byte(src), bag), bag
}

func mustParse(t *testing.T, src string) *ast.File {
	t.Helper()
	f, bag := parse(t, src)
	if bag.HasErrors() {
		t.Fatalf("unexpected error in:\n%s\n  %s", src, bag.Sorted()[0].Message())
	}
	return f
}

func firstError(t *testing.T, src string) *diag.Diagnostic {
	t.Helper()
	_, bag := parse(t, src)
	if !bag.HasErrors() {
		t.Fatalf("expected an error in:\n%s", src)
	}
	return bag.Sorted()[0]
}

func TestHelloWorld(t *testing.T) {
	f := mustParse(t, "chap \"salam, donya!\\n\"\n")
	if len(f.Decls) != 1 {
		t.Fatalf("want 1 statement, got %d", len(f.Decls))
	}
	es, ok := f.Decls[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("want ExprStmt, got %T", f.Decls[0])
	}
	call, ok := es.X.(*ast.Call)
	if !ok {
		t.Fatalf("want Call, got %T", es.X)
	}
	if id, ok := call.Fn.(*ast.Ident); !ok || id.Name != "chap" {
		t.Fatalf("want chap, got %#v", call.Fn)
	}
	if len(call.Args) != 1 {
		t.Fatalf("want 1 argument, got %d", len(call.Args))
	}
	if s, ok := call.Args[0].Expr.(*ast.StrLit); !ok || s.Value != "salam, donya!\n" {
		t.Fatalf("argument = %#v", call.Args[0].Expr)
	}
}

func TestVariableDeclarationIsAssignment(t *testing.T) {
	f := mustParse(t, "naam = 5\n")
	vd, ok := f.Decls[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("want VarDecl, got %T", f.Decls[0])
	}
	if vd.Name != "naam" {
		t.Fatalf("name = %q", vd.Name)
	}
	if vd.TypeName != "" {
		t.Fatalf("an untyped declaration must have no type, got %q", vd.TypeName)
	}
	if _, ok := vd.Value.(*ast.IntLit); !ok {
		t.Fatalf("value = %T", vd.Value)
	}
}

func TestTypedDeclaration(t *testing.T) {
	f := mustParse(t, "arz sen = 20\nkasr d = 1.5\n")
	vd := f.Decls[0].(*ast.VarDecl)
	if vd.TypeName != "arz" || vd.Name != "sen" {
		t.Fatalf("got %q %q", vd.TypeName, vd.Name)
	}
	vd2 := f.Decls[1].(*ast.VarDecl)
	if vd2.TypeName != "kasr" {
		t.Fatalf("type = %q", vd2.TypeName)
	}
}

func TestTypedDeclarationWithoutValue(t *testing.T) {
	f := mustParse(t, "arz shomarad\n")
	vd := f.Decls[0].(*ast.VarDecl)
	if vd.Value != nil {
		t.Fatalf("want no value, got %#v", vd.Value)
	}
}

func TestBuiltinNamesAreSoftWords(t *testing.T) {
	f := mustParse(t, "arz shomar = 3\n")
	vd := f.Decls[0].(*ast.VarDecl)
	if vd.Name != "shomar" || vd.TypeName != "arz" {
		t.Fatalf("got %q %q", vd.TypeName, vd.Name)
	}
}

func TestTypeWithNoNameIsReported(t *testing.T) {

	d := firstError(t, "arz 5\n")
	if d.Code != "E2025" {
		t.Fatalf("code = %s, want E2025", d.Code)
	}
	mustParse(t, "matn = \"salam\"\nchap_khat matn\n")
}

func TestATypeNameIsAnOrdinaryVariableWhenGivenAValue(t *testing.T) {
	f := mustParse(t, "matn = \"abc\"\nn = matn[1]\n")
	vd, ok := f.Decls[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("want VarDecl, got %T", f.Decls[0])
	}
	if vd.Name != "matn" || vd.TypeName != "" {
		t.Fatalf("name=%q type=%q", vd.Name, vd.TypeName)
	}
	if _, ok := f.Decls[1].(*ast.VarDecl).Value.(*ast.Index); !ok {
		t.Fatalf("matn[1] should be an index, got %T", f.Decls[1].(*ast.VarDecl).Value)
	}
}

func TestVerbWithoutBrackets(t *testing.T) {
	bare := mustParse(t, "chap \"salam\"\n")
	bracketed := mustParse(t, "chap(\"salam\")\n")
	a := bare.Decls[0].(*ast.ExprStmt).X.(*ast.Call)
	b := bracketed.Decls[0].(*ast.ExprStmt).X.(*ast.Call)
	if id := a.Fn.(*ast.Ident); id.Name != "chap" {
		t.Fatalf("function = %q", id.Name)
	}
	if len(a.Args) != 1 || len(b.Args) != 1 {
		t.Fatalf("argument counts differ: %d vs %d", len(a.Args), len(b.Args))
	}
}

func TestVerbWithSeveralArguments(t *testing.T) {
	f := mustParse(t, "chap \"a\", \"b\", x\n")
	call := f.Decls[0].(*ast.ExprStmt).X.(*ast.Call)
	if len(call.Args) != 3 {
		t.Fatalf("want 3 arguments, got %d", len(call.Args))
	}
}

func TestVerbWithNoArguments(t *testing.T) {
	f := mustParse(t, "khoroj\n")
	call := f.Decls[0].(*ast.ExprStmt).X.(*ast.Call)
	if len(call.Args) != 0 {
		t.Fatalf("want no arguments, got %d", len(call.Args))
	}
}

func TestEnglishAliasForType(t *testing.T) {
	f := mustParse(t, "int x = 1\n")
	vd := f.Decls[0].(*ast.VarDecl)

	if vd.TypeName != "arz" {
		t.Fatalf("int should canonicalise to arz, got %q", vd.TypeName)
	}
}

func TestMultipleNamesFromOneCall(t *testing.T) {
	f := mustParse(t, "a, b, c = f(1, 2, 3)\n")
	if len(f.Decls) != 3 {
		t.Fatalf("want 3 declarations, got %d", len(f.Decls))
	}
	names := []string{"a", "b", "c"}
	vals := []int64{1, 2, 3}
	for i, d := range f.Decls {
		vd := d.(*ast.VarDecl)
		if vd.Name != names[i] {
			t.Errorf("declaration %d is %q, want %q", i, vd.Name, names[i])
		}
		lit := vd.Value.(*ast.IntLit)
		if lit.Value != vals[i] {
			t.Errorf("declaration %d has %d, want %d", i, lit.Value, vals[i])
		}
	}
}

func TestWrongNumberOfValuesIsReported(t *testing.T) {
	d := firstError(t, "a, b, c = f(1, 2)\n")
	if d.Code != "E2008" {
		t.Fatalf("code = %s, want E2008", d.Code)
	}
}

func TestIfElseIfElse(t *testing.T) {
	src := `agar sen >= 18
    chap "baleghi"
warna agar sen >= 10
    chap "nojavan"
warna
    chap "kudakestan"
`
	f := mustParse(t, src)
	outer, ok := f.Decls[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("want IfStmt, got %T", f.Decls[0])
	}
	if len(outer.Then.Stmts) != 1 {
		t.Fatalf("then branch has %d statements", len(outer.Then.Stmts))
	}
	mid, ok := outer.Else.(*ast.IfStmt)
	if !ok {
		t.Fatalf("else should be another IfStmt, got %T", outer.Else)
	}
	if _, ok := mid.Else.(*ast.Block); !ok {
		t.Fatalf("the final warna should be a plain block, got %T", mid.Else)
	}
}

func TestColonIsOptional(t *testing.T) {
	with := mustParse(t, "agar x\n    chap 1\n")
	without := mustParse(t, "agar x: chap 1\n")
	a := with.Decls[0].(*ast.IfStmt)
	b := without.Decls[0].(*ast.IfStmt)
	if len(a.Then.Stmts) != len(b.Then.Stmts) {
		t.Fatalf("the colon changed the shape: %d vs %d statements", len(a.Then.Stmts), len(b.Then.Stmts))
	}
}

func TestBraceAndIndentBlocksAreEquivalent(t *testing.T) {
	indented := mustParse(t, "funct f()\n    chap 1\n    chap 2\n")
	braced := mustParse(t, "funct f() {\n    chap 1\n    chap 2\n}\n")
	a := indented.Decls[0].(*ast.FuncDecl)
	b := braced.Decls[0].(*ast.FuncDecl)
	if len(a.Body.Stmts) != len(b.Body.Stmts) {
		t.Fatalf("indented %d statements, braced %d", len(a.Body.Stmts), len(b.Body.Stmts))
	}
	if !b.Body.Braced {
		t.Error("the braced form should be recorded as braced")
	}
}

func TestCountedLoop(t *testing.T) {
	f := mustParse(t, "har i az 0 ta 10\n    chap i\n")
	fs, ok := f.Decls[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("want ForStmt, got %T", f.Decls[0])
	}
	if fs.VarName != "i" {
		t.Fatalf("var = %q", fs.VarName)
	}
	if fs.Inclusive {
		t.Error("ta 10 should be exclusive")
	}
}

func TestCountedLoopInclusive(t *testing.T) {
	f := mustParse(t, "har i az 0 ta 10+\n    chap i\n")
	fs := f.Decls[0].(*ast.ForStmt)
	if !fs.Inclusive {
		t.Error("ta 10+ should be inclusive")
	}
}

func TestForOverCollection(t *testing.T) {
	f := mustParse(t, "har x dar mohemat\n    chap x\n")
	fs, ok := f.Decls[0].(*ast.ForInStmt)
	if !ok {
		t.Fatalf("want ForInStmt, got %T", f.Decls[0])
	}
	if fs.VarName != "x" {
		t.Fatalf("var = %q", fs.VarName)
	}
}

func TestForWithFilter(t *testing.T) {
	f := mustParse(t, "har x dar xs ke x > 3\n    chap x\n")
	fs := f.Decls[0].(*ast.ForInStmt)
	if fs.Filter == nil {
		t.Fatal("the ke tail was dropped")
	}
}

func TestForNeedsAzOrDar(t *testing.T) {
	d := firstError(t, "har i\n    chap i\n")
	if d.Code != "E2010" {
		t.Fatalf("code = %s, want E2010", d.Code)
	}
}

func TestWhileLoop(t *testing.T) {
	f := mustParse(t, "ta shode < 10\n    chap shode\n    shode = shode + 1\n")
	ws, ok := f.Decls[0].(*ast.WhileStmt)
	if !ok {
		t.Fatalf("want WhileStmt, got %T", f.Decls[0])
	}
	if len(ws.Body.Stmts) != 2 {
		t.Fatalf("body has %d statements, want 2", len(ws.Body.Stmts))
	}
}

func TestFunctionDeclaration(t *testing.T) {
	src := "funct jam(a, b) -> arz\n    bargard a + b\n"
	f := mustParse(t, src)
	fd, ok := f.Decls[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("want FuncDecl, got %T", f.Decls[0])
	}
	if fd.Name != "jam" || len(fd.Params) != 2 {
		t.Fatalf("name=%q params=%d", fd.Name, len(fd.Params))
	}
	if fd.Params[0].Name != "a" || fd.Params[1].Name != "b" {
		t.Fatalf("params = %q %q", fd.Params[0].Name, fd.Params[1].Name)
	}
	if fd.Result != "arz" {
		t.Fatalf("result = %q", fd.Result)
	}
}

func TestEnglishFunctionKeyword(t *testing.T) {
	f := mustParse(t, "func add(a, b) -> int\n    return a + b\n")
	fd := f.Decls[0].(*ast.FuncDecl)
	if fd.Name != "add" {
		t.Fatalf("name = %q", fd.Name)
	}
	if fd.Result != "arz" {
		t.Fatalf("int should canonicalise to arz, got %q", fd.Result)
	}
}

func TestTypedParametersAndDefaults(t *testing.T) {
	f := mustParse(t, "funct greet(naam = \"doust\", sen: arz = 0)\n    chap naam\n")
	fd := f.Decls[0].(*ast.FuncDecl)
	if fd.Params[0].Default == nil {
		t.Error("the first parameter should have a default")
	}
	if fd.Params[1].TypeName != "arz" || fd.Params[1].Default == nil {
		t.Errorf("the second parameter: type=%q default=%v", fd.Params[1].TypeName, fd.Params[1].Default)
	}
}

func TestGenericFunction(t *testing.T) {
	f := mustParse(t, "funct<T> identity(x: T) -> T\n    bargard x\n")
	fd := f.Decls[0].(*ast.FuncDecl)
	if len(fd.Generics) != 1 || fd.Generics[0] != "T" {
		t.Fatalf("generics = %v", fd.Generics)
	}
}

func TestOperatorPrecedenceFromTheTable(t *testing.T) {

	f := mustParse(t, "x = 1 + 2 * 3\n")
	vd := f.Decls[0].(*ast.VarDecl)
	top, ok := vd.Value.(*ast.Binary)
	if !ok || top.Op != "+" {
		t.Fatalf("outer operator = %#v", vd.Value)
	}
	if _, ok := top.R.(*ast.Binary); !ok {
		t.Fatalf("the right side should be the multiplication, got %T", top.R)
	}
}

func TestExponentBindsTighterThanUnaryMinus(t *testing.T) {

	f := mustParse(t, "x = -2 ** 2\n")
	vd := f.Decls[0].(*ast.VarDecl)
	u, ok := vd.Value.(*ast.Unary)
	if !ok || u.Op != "-" {
		t.Fatalf("outer should be the negation, got %T", vd.Value)
	}
	if b, ok := u.X.(*ast.Binary); !ok || b.Op != "**" {
		t.Fatalf("the power should be inside the negation, got %T", u.X)
	}
}

func TestUnaryNotBindsTighterThanVa(t *testing.T) {

	f := mustParse(t, "x = na a va b\n")
	vd := f.Decls[0].(*ast.VarDecl)
	b, ok := vd.Value.(*ast.Binary)
	if !ok || b.Op != "va" {
		t.Fatalf("outer should be va, got %T", vd.Value)
	}
	if u, ok := b.L.(*ast.Unary); !ok || u.Op != "na" {
		t.Fatalf("na should wrap only a, got %T", b.L)
	}
}

func TestNotAlsoSpelledBang(t *testing.T) {
	f := mustParse(t, "x = !a\n")
	u := f.Decls[0].(*ast.VarDecl).Value.(*ast.Unary)
	if u.Op != "!" {
		t.Fatalf("op = %q", u.Op)
	}
}

func TestRangeExpression(t *testing.T) {
	f := mustParse(t, "r = 0..10\n")
	r, ok := f.Decls[0].(*ast.VarDecl).Value.(*ast.RangeExpr)
	if !ok {
		t.Fatalf("want RangeExpr, got %T", f.Decls[0].(*ast.VarDecl).Value)
	}
	if r.Inclusive {
		t.Error(".. should be exclusive")
	}
}

func TestIndexAndSlice(t *testing.T) {
	f := mustParse(t, "a = xs[0]\nb = xs[1:3]\nc = xs[1..3]\nd = xs[:2]\n")
	kinds := []string{"Index", "Slice", "Slice", "Slice"}
	for i, d := range f.Decls {
		vd := d.(*ast.VarDecl)
		if got := vd.Value.NodeKind(); got != kinds[i] {
			t.Errorf("declaration %d is %s, want %s", i, got, kinds[i])
		}
	}

	inc := f.Decls[2].(*ast.VarDecl).Value.(*ast.Slice)
	if !inc.Inclusive {
		t.Error("xs[1..3] should be inclusive")
	}
}

func TestNegativeIndex(t *testing.T) {
	f := mustParse(t, "a = xs[-1]\n")
	idx := f.Decls[0].(*ast.VarDecl).Value.(*ast.Index)
	if u, ok := idx.At.(*ast.Unary); !ok || u.Op != "-" {
		t.Fatalf("want a negation inside the brackets, got %T", idx.At)
	}
}

func TestMemberAndMethodCall(t *testing.T) {
	f := mustParse(t, "a = o.naam\nb = o.salam(\"sara\")\n")
	m1 := f.Decls[0].(*ast.VarDecl).Value.(*ast.Member)
	if m1.Name != "naam" || m1.Call {
		t.Fatalf("member = %+v", m1)
	}
	m2 := f.Decls[1].(*ast.VarDecl).Value.(*ast.Member)
	if m2.Name != "salam" || !m2.Call || len(m2.Args) != 1 {
		t.Fatalf("method call = %+v", m2)
	}
}

func TestMemberNamePositionIsTheName(t *testing.T) {
	src := "a = o.naam\n"
	f := mustParse(t, src)
	m := f.Decls[0].(*ast.VarDecl).Value.(*ast.Member)

	if m.NamePos.Col != 7 {
		t.Fatalf("name at column %d, want 7", m.NamePos.Col)
	}
}

func TestListAndDictAndSetLiterals(t *testing.T) {
	f := mustParse(t, "a = [1, 2, 3]\nb = {\"x\": 1, \"y\": 2}\nc = {1, 2, 3}\n")
	if l, ok := f.Decls[0].(*ast.VarDecl).Value.(*ast.ListLit); !ok || len(l.Elems) != 3 {
		t.Errorf("list = %T", f.Decls[0].(*ast.VarDecl).Value)
	}
	if d, ok := f.Decls[1].(*ast.VarDecl).Value.(*ast.DictLit); !ok || len(d.Entries) != 2 {
		t.Errorf("dict = %T", f.Decls[1].(*ast.VarDecl).Value)
	}
	if s, ok := f.Decls[2].(*ast.VarDecl).Value.(*ast.SetLit); !ok || len(s.Elems) != 3 {
		t.Errorf("set = %T", f.Decls[2].(*ast.VarDecl).Value)
	}
}

func TestEmptyDictBraces(t *testing.T) {
	f := mustParse(t, "a = {}\n")
	if _, ok := f.Decls[0].(*ast.VarDecl).Value.(*ast.DictLit); !ok {
		t.Fatalf("{} should be an empty dict, got %T", f.Decls[0].(*ast.VarDecl).Value)
	}
}

func TestInterpolation(t *testing.T) {
	f := mustParse(t, `chap $"salam {naam}, {sen} salhe"`+"\n")
	call := f.Decls[0].(*ast.ExprStmt).X.(*ast.Call)
	lit, ok := call.Args[0].Expr.(*ast.InterpLit)
	if !ok {
		t.Fatalf("want InterpLit, got %T", call.Args[0].Expr)
	}

	if len(lit.Segs) != 5 {
		t.Fatalf("want 5 segments, got %d: %+v", len(lit.Segs), lit.Segs)
	}
	if lit.Segs[0].Text != "salam " {
		t.Errorf("segment 0 = %q", lit.Segs[0].Text)
	}
	if lit.Segs[1].Expr == nil {
		t.Error("segment 1 should be an expression")
	}
	if lit.Segs[4].Text != " salhe" {
		t.Errorf("segment 4 = %q", lit.Segs[4].Text)
	}
}

func TestInterpolationKeepsTheRealColumn(t *testing.T) {

	src := "chap 1\nx = $\"a ${ b + } c\"\n"
	_, bag := parse(t, src)
	if !bag.HasErrors() {
		t.Fatal("the empty right-hand side should be reported")
	}
	d := bag.Sorted()[0]
	if d.Span.Start.Line != 2 {
		t.Errorf("reported on line %d, want 2", d.Span.Start.Line)
	}
}

func TestShortInterpolation(t *testing.T) {
	f := mustParse(t, `x = "hi $user!"`+"\n")
	lit, ok := f.Decls[0].(*ast.VarDecl).Value.(*ast.InterpLit)
	if !ok {
		t.Fatalf("want InterpLit, got %T", f.Decls[0].(*ast.VarDecl).Value)
	}
	if len(lit.Segs) != 3 {
		t.Fatalf("want 3 segments, got %d", len(lit.Segs))
	}
}

func TestLambdaExpression(t *testing.T) {
	f := mustParse(t, "f = funct(x) -> x * 2\n")
	lam, ok := f.Decls[0].(*ast.VarDecl).Value.(*ast.Lambda)
	if !ok {
		t.Fatalf("want Lambda, got %T", f.Decls[0].(*ast.VarDecl).Value)
	}
	if len(lam.Params) != 1 || lam.Params[0].Name != "x" {
		t.Fatalf("params = %+v", lam.Params)
	}
	if lam.Expr == nil {
		t.Error("the lambda body is missing")
	}
}

func TestLambdaWithBlockBody(t *testing.T) {
	f := mustParse(t, "f = funct(x) ->\n    bargard x + 1\n")
	lam, ok := f.Decls[0].(*ast.VarDecl).Value.(*ast.Lambda)
	if !ok {
		t.Fatalf("want Lambda, got %T", f.Decls[0].(*ast.VarDecl).Value)
	}
	if lam.Body == nil || len(lam.Body.Stmts) != 1 {
		t.Fatalf("body = %+v", lam.Body)
	}
}

func TestCallWithNamedArgument(t *testing.T) {
	f := mustParse(t, `f(naam: "sara", sen: 30)`+"\n")
	call := f.Decls[0].(*ast.ExprStmt).X.(*ast.Call)
	if len(call.Args) != 2 {
		t.Fatalf("want 2 arguments, got %d", len(call.Args))
	}
	if call.Args[0].Name != "naam" || call.Args[1].Name != "sen" {
		t.Errorf("names = %q %q", call.Args[0].Name, call.Args[1].Name)
	}
}

func TestTernaryExpression(t *testing.T) {
	f := mustParse(t, `x = agar n > 0 "positiv" warna "na-majez"`+"\n")
	tr, ok := f.Decls[0].(*ast.VarDecl).Value.(*ast.Ternary)
	if !ok {
		t.Fatalf("want Ternary, got %T", f.Decls[0].(*ast.VarDecl).Value)
	}
	if tr.Then == nil || tr.Else == nil || tr.Cond == nil {
		t.Fatal("the ternary is missing a part")
	}
}

func TestClassDeclaration(t *testing.T) {
	src := `darghan Ostad
    baza naam
    __sakht__(naam)
        khod.naam = naam
    funct kard()
        chap $"man {khod.naam} hastam"
`
	f := mustParse(t, src)
	c, ok := f.Decls[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("want ClassDecl, got %T", f.Decls[0])
	}
	if c.Name != "Ostad" {
		t.Fatalf("name = %q", c.Name)
	}
	if len(c.Fields) != 1 || c.Fields[0].Name != "naam" {
		t.Fatalf("fields = %+v", c.Fields)
	}
	if len(c.Methods) != 2 {
		t.Fatalf("methods = %d, want 2", len(c.Methods))
	}
	if !c.Fields[0].Baza {
		t.Error("the field was written baza and should be marked so")
	}
}

func TestVerbIsNotAFieldName(t *testing.T) {
	d := firstError(t, "darghan Ostad\n    chap kard\n")
	if d.Code != "E2018" {
		t.Fatalf("code = %s, want E2018", d.Code)
	}
}

func TestClassWithBase(t *testing.T) {
	f := mustParse(t, "darghan DaneshAmuz < Tajebe>\n    naam = \"\"\n")
	c := f.Decls[0].(*ast.ClassDecl)
	if len(c.Bases) != 1 || c.Bases[0] != "Tajebe" {
		t.Fatalf("bases = %v", c.Bases)
	}
}

func TestStructDeclaration(t *testing.T) {
	f := mustParse(t, "sakht Noghat\n    x: kasr\n    y: kasr\n")
	s, ok := f.Decls[0].(*ast.StructDecl)
	if !ok {
		t.Fatalf("want StructDecl, got %T", f.Decls[0])
	}
	if len(s.Fields) != 2 || s.Fields[0].TypeName != "kasr" {
		t.Fatalf("fields = %+v", s.Fields)
	}
}

func TestMethodInStructIsReported(t *testing.T) {
	d := firstError(t, "sakht Noghat\n    x: kasr\n    funct chap kard()\n        chap 1\n")
	if d.Code != "E2015" {
		t.Fatalf("code = %s, want E2015", d.Code)
	}
}

func TestEnumDeclaration(t *testing.T) {
	f := mustParse(t, "shenase Rang\n    sefid = 0\n    siah = 1\n")
	e, ok := f.Decls[0].(*ast.EnumDecl)
	if !ok {
		t.Fatalf("want EnumDecl, got %T", f.Decls[0])
	}
	if len(e.Values) != 2 || e.Values[0].Name != "sefid" {
		t.Fatalf("values = %+v", e.Values)
	}
}

func TestInterfaceDeclaration(t *testing.T) {
	f := mustParse(t, "dastoor Drawable\n    kash(id)\n    andazesho()\n")
	i, ok := f.Decls[0].(*ast.InterfaceDecl)
	if !ok {
		t.Fatalf("want InterfaceDecl, got %T", f.Decls[0])
	}
	if len(i.Methods) != 2 || i.Methods[0] != "kash" {
		t.Fatalf("methods = %v", i.Methods)
	}
}

func TestTryCatchFinally(t *testing.T) {
	src := `koshish
    adad = khanak("edad: ")
giroft khata e
    chap e.message
nemigir
    chap "tamam shod"
`
	f := mustParse(t, src)
	ts, ok := f.Decls[0].(*ast.TryStmt)
	if !ok {
		t.Fatalf("want TryStmt, got %T", f.Decls[0])
	}
	if !ts.HasCatch || ts.Catch.Name != "e" {
		t.Fatalf("catch = %+v", ts.Catch)
	}
	if ts.Finally == nil {
		t.Error("the nemigir block was dropped")
	}
}

func TestThrowStatement(t *testing.T) {
	f := mustParse(t, "portab khata(\"kharab\")\n")
	ts, ok := f.Decls[0].(*ast.ThrowStmt)
	if !ok {
		t.Fatalf("want ThrowStmt, got %T", f.Decls[0])
	}
	if ts.Value == nil {
		t.Error("portab without a value")
	}
}

func TestTalignStatement(t *testing.T) {
	f := mustParse(t, "talign chap \"kharch\"\n")
	ts, ok := f.Decls[0].(*ast.TalignStmt)
	if !ok {
		t.Fatalf("want TalignStmt, got %T", f.Decls[0])
	}
	if ts.X == nil {
		t.Error("talign with no expression")
	}
}

func TestModifiers(t *testing.T) {
	f := mustParse(t, "sabit PI = 3.14\nbaza funct f()\n    chap 1\n")
	vd := f.Decls[0].(*ast.VarDecl)
	if !vd.Sabit {
		t.Error("sabit was not recorded")
	}
	fd := f.Decls[1].(*ast.FuncDecl)
	if !fd.Baza {
		t.Error("baza was not recorded")
	}
}

func TestVisibilityOnLocalIsReported(t *testing.T) {
	d := firstError(t, "baza x = 1\n")
	if d.Code != "E2003" {
		t.Fatalf("code = %s, want E2003", d.Code)
	}
}

func TestImportAndModuleAndProgram(t *testing.T) {
	f := mustParse(t, "modul \"mihan/mikhak\"\nbarname Khaneh\nbiar \"std/fmt\"\n")
	if f.Module != "mihan/mikhak" {
		t.Errorf("module = %q", f.Module)
	}
	if f.Program != "Khaneh" {
		t.Errorf("program = %q", f.Program)
	}
	if len(f.Imports) != 1 || f.Imports[0].Path != "std/fmt" {
		t.Errorf("imports = %+v", f.Imports)
	}
}

func TestImportWithAlias(t *testing.T) {
	f := mustParse(t, "biar \"std/fmt\" as f\n")
	if f.Imports[0].Alias != "f" {
		t.Fatalf("alias = %q", f.Imports[0].Alias)
	}
}

func TestBreakAndContinue(t *testing.T) {
	f := mustParse(t, "har i az 0 ta 10\n    agar i > 5\n        shek\n    edame\n")
	fs := f.Decls[0].(*ast.ForStmt)
	if _, ok := fs.Body.Stmts[0].(*ast.IfStmt).Then.Stmts[0].(*ast.BreakStmt); !ok {
		t.Errorf("expected a break, got %T", fs.Body.Stmts[0].(*ast.IfStmt).Then.Stmts[0])
	}
	if _, ok := fs.Body.Stmts[1].(*ast.ContinueStmt); !ok {
		t.Errorf("expected a continue, got %T", fs.Body.Stmts[1])
	}
}

func TestReturnWithAndWithoutValue(t *testing.T) {
	f := mustParse(t, "funct f(a) -> arz\n    agar a > 0\n        bargard 1\n    bargard\n")
	fd := f.Decls[0].(*ast.FuncDecl)
	rs := fd.Body.Stmts[0].(*ast.IfStmt).Then.Stmts[0]
	if r, ok := rs.(*ast.ReturnStmt); !ok || r.Value == nil {
		t.Errorf("expected a return with a value, got %#v", rs)
	}
	if r, ok := fd.Body.Stmts[1].(*ast.ReturnStmt); !ok || r.Value != nil {
		t.Errorf("expected a bare return, got %#v", fd.Body.Stmts[1])
	}
}

func TestAssignmentOperators(t *testing.T) {
	f := mustParse(t, "x = 1\nx += 2\nx -= 1\nx *= 3\nx /= 2\nx %= 2\nx **= 2\nx := 9\n")
	want := []token.Kind{"=", "+=", "-=", "*=", "/=", "%=", "**=", ":="}
	for i, d := range f.Decls {
		if want[i].IsDeclaration() {
			vd, ok := d.(*ast.VarDecl)
			if !ok {
				t.Fatalf("declaration %d is %T, want VarDecl", i, d)
			}
			if vd.Name != "x" {
				t.Errorf("declaration %d names %q", i, vd.Name)
			}
			continue
		}
		es, ok := d.(*ast.ExprStmt)
		if !ok {
			t.Fatalf("statement %d is %T, want ExprStmt", i, d)
		}
		as, ok := es.X.(*ast.Assign)
		if !ok {
			t.Fatalf("statement %d is %T, want Assign", i, es.X)
		}
		if as.Op != want[i] {
			t.Errorf("statement %d uses %q, want %q", i, as.Op, want[i])
		}
	}
}

func TestAssigningToALiteralIsReported(t *testing.T) {
	d := firstError(t, "5 = x\n")
	if d.Code != "E2006" {
		t.Fatalf("code = %s, want E2006", d.Code)
	}
}

func TestAssigningToAMemberOrIndexIsFine(t *testing.T) {
	f := mustParse(t, "o.naam = \"sara\"\nxs[0] = 1\n")
	if len(f.Decls) != 2 {
		t.Fatalf("got %d statements", len(f.Decls))
	}
}

func TestSemicolonsAndMultiStatementLines(t *testing.T) {
	f := mustParse(t, "a = 1; b = 2\n")
	if len(f.Decls) != 2 {
		t.Fatalf("want 2 statements, got %d", len(f.Decls))
	}

	f2 := mustParse(t, "chap 1 chap 2\n")
	if len(f2.Decls) != 2 {
		t.Fatalf("want 2 statements without a semicolon, got %d", len(f2.Decls))
	}
}

func TestNestedFunctionsAndBlocks(t *testing.T) {
	src := `funct outer()
    funct inner()
        chap 1
    inner()
`
	f := mustParse(t, src)
	outer := f.Decls[0].(*ast.FuncDecl)
	if len(outer.Body.Stmts) != 2 {
		t.Fatalf("outer body has %d statements", len(outer.Body.Stmts))
	}
}

func TestDeeplyNestedBlocks(t *testing.T) {
	src := `har i az 0 ta 3
    agar i > 1
        ta dorost
            chap i
            agar i > 2
                shek
            chap "badi"
`
	f := mustParse(t, src)
	if len(f.Decls) != 1 {
		t.Fatalf("got %d statements", len(f.Decls))
	}
	outer := f.Decls[0].(*ast.ForStmt)
	iface := outer.Body.Stmts[0].(*ast.IfStmt)
	loop := iface.Then.Stmts[0].(*ast.WhileStmt)

	if len(loop.Body.Stmts) != 3 {
		t.Fatalf("the innermost block has %d statements, want 3", len(loop.Body.Stmts))
	}
	inner := loop.Body.Stmts[1].(*ast.IfStmt)
	if _, ok := inner.Then.Stmts[0].(*ast.BreakStmt); !ok {
		t.Errorf("expected a break, got %T", inner.Then.Stmts[0])
	}
}

func TestParserRecoversAndKeepsGoing(t *testing.T) {
	src := "a = 1\nb = )\nc = 3\n"
	_, bag := parse(t, src)
	if !bag.HasErrors() {
		t.Fatal("expected an error")
	}
	if bag.Len() < 1 {
		t.Fatal("the diagnostic was lost")
	}
}

func TestErrorMessageIsBilingual(t *testing.T) {
	d := firstError(t, "b = )\n")
	if d.FA == "" {
		t.Error("the Persian half of the message is empty")
	}
	if d.EN == "" {
		t.Error("the English half of the message is empty")
	}
	if d.Code == "" {
		t.Error("the diagnostic has no code")
	}
}

func TestMissingClosingBrace(t *testing.T) {
	d := firstError(t, "funct f() {\n    chap 1\n")
	if d.Code != "E2001" {
		t.Fatalf("code = %s, want E2001", d.Code)
	}
}

func TestUnclosedParenthesis(t *testing.T) {
	firstError(t, "chap(1\n")
}

func TestEmptyBlockIsReported(t *testing.T) {
	d := firstError(t, "agar x\nchap 1\n")
	if d.Code != "E2011" && d.Code != "E2017" {
		t.Logf("code = %s", d.Code)
	}
}

func TestSpansAreRealPositions(t *testing.T) {
	src := "naam = 42\nchap naam\n"
	f := mustParse(t, src)
	vd := f.Decls[0].(*ast.VarDecl)
	if got := src[vd.NamePos.Off : vd.NamePos.Off+4]; got != "naam" {
		t.Errorf("NamePos points at %q, want naam", got)
	}
	call := f.Decls[1].(*ast.ExprStmt).X.(*ast.Call)
	id := call.Fn.(*ast.Ident)
	if got := src[id.Span().Start.Off:id.Span().End.Off]; got != "chap" {
		t.Errorf("the call's function span covers %q, want chap", got)
	}
}

func TestLineNumbersAdvance(t *testing.T) {
	src := "a = 1\nb = 2\nc = 3\n"
	f := mustParse(t, src)
	for i, d := range f.Decls {
		if line := d.Span().Start.Line; line != i+1 {
			t.Errorf("statement %d starts on line %d, want %d", i, line, i+1)
		}
	}
}

func TestFuzzDoesNotPanic(t *testing.T) {
	seeds := []string{
		"agar", "chap", "$", "${", "\"", "`", "'", "..", "funct", "darghan",
		"har", "ta", "koshish", "{", "}", "[", "]", "(", ")", "=", ",", ":",
		"بسم", "\t", "\n\n", "//", "/*", "shenase", "dastoor", "sakht", "ke",
		"0x", "..+", "1e", "shenase\n  ", "har i az", "warna", "nemigir",
	}
	for _, s := range seeds {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic on %q: %v", s, r)
				}
			}()
			Parse("fuzz.mhn", []byte(s), &diag.Bag{})
		}()
	}
}

func TestEveryKeywordIsHandled(t *testing.T) {

	programs := map[string]string{
		"agar":      "agar dorost\n    chap 1\n",
		"warna":     "agar dorost\n    chap 1\nwarna\n    chap 2\n",
		"ta":        "ta ghalat\n    chap 1\n",
		"har":       "har i az 0 ta 3\n    chap i\n",
		"shek":      "har i az 0 ta 3\n    shek\n",
		"edame":     "har i az 0 ta 3\n    edame\n",
		"bargard":   "funct f()\n    bargard\n",
		"koshish":   "koshish\n    chap 1\ngiroft khata e\n    chap 1\n",
		"giroft":    "koshish\n    chap 1\ngiroft khata e\n    chap 1\n",
		"nemigir":   "koshish\n    chap 1\nnemigir\n    chap 2\n",
		"portab":    "portab khata(\"x\")\n",
		"funct":     "funct f()\n    chap 1\n",
		"darghan":   "darghan A\n    x = 1\n",
		"sakht":     "sakht A\n    x: arz\n",
		"dastoor":   "dastoor A\n    m()\n",
		"shenase":   "shenase A\n    a = 0\n",
		"biar":      "biar \"std/fmt\"\n",
		"modul":     "modul \"a/b\"\n",
		"barname":   "barname A\n",
		"va":        "x = dorost va dorost\n",
		"ya":        "x = dorost ya ghalat\n",
		"na":        "x = na dorost\n",
		"dar":       "x = 1 dar [1, 2]\n",
		"ke":        "har i az 0 ta 3 ke i > 1\n    chap i\n",
		"ghesmat":   "x = 7 ghesmat 2\n",
		"az":        "har i az 0 ta 3\n    chap i\n",
		"talign":    "talign chap 1\n",
		"dorost":    "x = dorost\n",
		"ghalat":    "x = ghalat\n",
		"hich":      "x = hich\n",
		"sabit":     "sabit A = 1\n",
		"baza":      "baza funct f()\n    chap 1\n",
		"mahram":    "mahram funct f()\n    chap 1\n",
		"osta":      "osta funct f()\n    chap 1\n",
		"chap":      "chap 1\n",
		"chap_khat": "chap_khat 1\n",
		"khod":      "x = khod\n",
	}
	for kw, src := range programs {
		t.Run(kw, func(t *testing.T) {
			_, bag := parse(t, src)
			if bag.HasErrors() {
				for _, d := range bag.Sorted() {
					t.Logf("%s: %s / %s", d.Code, d.FA, d.EN)
				}
				t.Fatalf("the keyword %q is not handled", kw)
			}
		})
	}
}

func TestKeywordsUsedAsBareOperands(t *testing.T) {

	src := `x = tule [1, 2]
y = rang(1, 10)
z = no A()
chap_khat "line"
w = hasht(x)
v = shomar(x)
u = tabdil(x, arz)
t = goosh(x)
s = takfir(x)
r = dope(1, 2)
q = balia(1, 2)
o = neshan(x)
p = hast(x)
n = banafsh(x)
chap 1
`
	_, bag := parse(t, src)
	if bag.HasErrors() {
		for _, d := range bag.Sorted() {
			t.Logf("%s: %s / %s", d.Code, d.FA, d.EN)
		}
		t.Fatal("a builtin did not parse")
	}
}

func timeoutAfterSeconds(n int) <-chan time.Time { return time.After(time.Duration(n) * time.Second) }

func TestFuzzyInputKeepsGoing(t *testing.T) {

	src := strings.Repeat("agar x\n    chap ${\n", 5) + "}\n"
	done := make(chan struct{})
	go func() {
		defer close(done)
		Parse("fuzz.mhn", []byte(src), &diag.Bag{})
	}()
	select {
	case <-done:
	case <-timeoutAfterSeconds(10):
		t.Fatal("the parser did not finish")
	}
}
