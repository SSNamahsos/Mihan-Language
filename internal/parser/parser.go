package parser

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"mihan/internal/ast"
	"mihan/internal/diag"
	"mihan/internal/lexer"
	"mihan/internal/token"
)

const (
	maxErrors = 10

	maxDepth = 200
)

const prefixTight = 13

type Parser struct {
	file string
	toks []token.Token
	pos  int

	bag    *diag.Bag
	errors int
	depth  int

	furthest int

	plusEndsHeader bool

	inEnum bool

	generics []string
}

var bom = []byte{0xEF, 0xBB, 0xBF}

func Parse(file string, src []byte, bag *diag.Bag) *ast.File {
	if bag == nil {
		bag = &diag.Bag{}
	}
	src = bytes.TrimPrefix(src, bom)
	lx := lexer.New(file, src, bag)
	p := &Parser{file: file, toks: lx.Tokens(), bag: bag}
	f := p.parseFile()
	f.Path = file
	f.Source = src
	f.Comments = lx.Comments()
	return f
}

func (p *Parser) errf(code, fa, enFormat string, args ...any) *ast.ErrorExpr {

	faArgs := args
	if !strings.Contains(fa, "%") {
		faArgs = nil
	}
	return p.errfArgs(code, fa, faArgs, enFormat, args)
}

func (p *Parser) errfArgs(code, fa string, faArgs []any, enFormat string, enArgs []any) *ast.ErrorExpr {
	span := ast.SpanOf(p.cur().Start(), p.cur().EndPos())
	if p.errors < maxErrors {
		p.errors++
		p.furthest = p.pos
		p.bag.Add(diag.ErrorfArgs(code, span, fa, faArgs, enFormat, enArgs).WithFile(p.file))
	}

	return &ast.ErrorExpr{
		Loc:     span,
		Message: fmt.Sprintf(fa, faArgs...),
		English: fmt.Sprintf(enFormat, enArgs...),
	}
}

func (p *Parser) errAt(t token.Token, code, fa, enFormat string, args ...any) {
	if p.errors >= maxErrors {
		return
	}
	p.errors++
	span := ast.SpanOf(t.Span().Start, t.EndPos())
	p.bag.Add(diag.Errorf(code, span, fa, enFormat, args...).WithFile(p.file))
}

func (p *Parser) note(t token.Token, code, fa, enFormat string, args ...any) {
	p.bag.Add(diag.Errorf(code, ast.SpanOf(t.Span().Start, t.EndPos()), fa, enFormat, args...).WithFile(p.file))
}

func (p *Parser) errStmt(code, fa, enFormat string, args ...any) ast.Stmt {
	e := p.errf(code, fa, enFormat, args...)
	return &ast.ExprStmt{Loc: e.Loc, X: e}
}

func (p *Parser) cur() token.Token { return p.at_(0) }

func (p *Parser) at_(n int) token.Token {
	i := p.pos + n
	if i < 0 {
		i = 0
	}
	if i >= len(p.toks) {
		return p.toks[len(p.toks)-1]
	}
	return p.toks[i]
}

func (p *Parser) at(k token.Kind) bool { return p.cur().Kind == k }

func (p *Parser) atAny(kinds ...token.Kind) bool {
	for _, k := range kinds {
		if p.cur().Kind == k {
			return true
		}
	}
	return false
}

func (p *Parser) advance() token.Token {
	t := p.cur()
	if p.pos < len(p.toks)-1 {
		p.pos++
	}
	return t
}

func (p *Parser) accept(k token.Kind) bool {
	if p.at(k) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) expect(k token.Kind, what string) (token.Token, bool) {
	if p.at(k) {
		return p.advance(), true
	}
	p.errfArgs("E2001",
		"اینجا «%s» را می‌خواستم، ولی %s آمد", []any{farsi(what), p.cur().DescribeFarsi()},
		"I expected %s here, but found %s", []any{what, p.cur().Describe()})
	return p.cur(), false
}

var farsiPhrases = map[string]string{
	"the modul name":                        "نامِ ماژول",
	"the barname":                           "نامِ پرونده",
	"an opening brace":                      "«{» گشوده",
	"a closing brace":                       "«}» بسته",
	"the end of the indented block":         "پایانِ بلوکِ تورفتگی",
	"the end of the indented body":          "پایانِ بدنهٔ تورفتگی",
	"the end of the list of type names":     "پایانِ فهرستِ نامِ نوع‌ها",
	"the end of the type parameters":        "پایانِ پارامترهای نوع",
	"a closing parenthesis":                 "«)» بسته",
	"the module path":                       "راهِ ماژول",
	"a short name for the module":           "نامی کوتاه برای ماژول",
	"the name of the dastoor":               "نامِ دستور",
	"the name of the darghan":               "نامِ دایره",
	"the end of the list of base names":     "پایانِ فهرستِ نامِ پایه‌ها",
	"the name of the sakht":                 "نامِ ساخت",
	"the name of the shenase":               "نامِ شناسه",
	"a closing bracket":                     "«]» بسته",
	"a } to close the ${":                   "«}» برای بستنِ «${",
	"a colon between the key and the value": "«:» میان کلید و مقدار",
}

func farsi(what string) string {
	if f, ok := farsiPhrases[what]; ok {
		return f
	}
	return "این را"
}

func (p *Parser) start() diag.Pos { return p.cur().Start() }

func (p *Parser) finish(start diag.Pos) ast.Loc {
	end := p.cur().Start()
	if p.pos > 0 {
		end = p.at_(p.pos - 1).EndPos()
	}
	return ast.SpanOf(start, end)
}

func (p *Parser) skipSeps() {
	for p.atAny(token.NEWLINE, ";") {
		p.advance()
	}
}

func (p *Parser) endStatement() {
	if p.atAny(token.NEWLINE, ";") {
		p.advance()
	}
}

func (p *Parser) parseFile() *ast.File {
	f := &ast.File{}

	for p.at("modul") {
		p.advance()
		if t, ok := p.expect(token.STRING, "the modul name"); ok && t.Kind == token.STRING {
			f.Module = t.Text
		}
		p.endStatement()
	}
	for p.at("barname") {
		p.advance()
		if t, ok := p.expect(token.IDENT, "the barname"); ok {
			f.Program = t.Lit
		}
		p.endStatement()
	}
	f.Decls = p.parseStatementList()
	for _, d := range f.Decls {
		if im, ok := d.(*ast.ImportDecl); ok {
			f.Imports = append(f.Imports, im)
		}
	}
	return f
}

func (p *Parser) parseStatementList() []ast.Stmt {
	var out []ast.Stmt
	for {
		p.skipSeps()
		switch p.cur().Kind {
		case token.EOF, token.DEDENT, "}":
			return out
		}
		before := p.pos
		s := p.parseStatement()

		if m, ok := s.(multiDecl); ok {
			out = append(out, m...)
		} else if s != nil {
			out = append(out, s)
		}
		if p.pos == before {

			if p.at(token.EOF) {
				return out
			}
			p.advance()
		}
	}
}

type multiDecl []ast.Stmt

func (m multiDecl) Span() diag.Span {
	if len(m) == 0 {
		return diag.Span{}
	}
	return m[0].Span()
}
func (m multiDecl) NodeKind() string { return "MultiDecl" }
func (m multiDecl) StmtNode()        {}

func (p *Parser) parseStatement() ast.Stmt {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > maxDepth {
		return p.errStmt("E2002", "این برنامه خیلی عمیق شده است", "this program nests too deeply to read")
	}

	switch p.cur().Kind {
	case "agar":
		return p.parseIf()
	case "ta":
		return p.parseWhile()
	case "har":
		return p.parseFor()
	case "funct":
		return p.parseFuncDecl(nil)
	case "darghan":
		return p.parseClassDecl(nil)
	case "sakht":
		return p.parseStructDecl(nil)
	case "dastoor":
		return p.parseInterfaceDecl()
	case "shenase":
		return p.parseEnumDecl()
	case "biar":
		return p.parseImport()
	case "bargard":
		return p.parseReturn()
	case "shek":
		t := p.advance()
		p.endStatement()
		return &ast.BreakStmt{Loc: p.finish(t.Span().Start)}
	case "edame":
		t := p.advance()
		p.endStatement()
		return &ast.ContinueStmt{Loc: p.finish(t.Span().Start)}
	case "portab":
		return p.parseThrow()
	case "koshish":
		return p.parseTry()
	case "talign":
		t := p.advance()
		x := p.parseExpression()
		p.endStatement()
		return &ast.TalignStmt{Loc: p.finish(t.Span().Start), X: x}
	case "sabit", "baza", "mahram", "osta":
		return p.parseModified()
	case "{":
		b := p.parseBracedBlock()
		return b
	}
	return p.parseSimpleStatement()
}

type mods struct {
	sabit, baza, mahram, osta bool
}

func (p *Parser) parseModified() ast.Stmt {
	start := p.start()
	m := p.readMods()
	switch p.cur().Kind {
	case "funct":
		return p.parseFuncDecl(&m)
	case "darghan":
		return p.parseClassDecl(&m)
	case "sakht":
		return p.parseStructDecl(&m)
	}
	if m.osta {
		p.errf("E2003", "`osta` فقط روی تابع، darghan یا sakht می‌نشیند",
			"osta only belongs on a funct, darghan or sakht")
	}
	if m.mahram {
		p.errf("E2003", "`mahram` فقط روی تابع، darghan یا sakht می‌نشیند",
			"mahram only belongs on a funct, darghan or sakht")
	}
	if m.baza {
		p.errf("E2003", "`baza` فقط روی تابع، darghan یا sakht می‌نشیند",
			"baza only belongs on a funct, darghan or sakht")
	}
	st := p.parseSimpleStatement()
	if vd, ok := st.(*ast.VarDecl); ok {
		vd.Sabit = m.sabit
		return vd
	}
	if m.sabit {
		p.errf("E2004", "`sabit` پیش از نامی می‌آید که مقدار می‌گیرد",
			"sabit has to sit in front of a name that is being given a value")
	}
	_ = start
	return st
}

func (p *Parser) readMods() mods {
	var m mods
	for {
		switch p.cur().Kind {
		case "sabit":
			m.sabit = true
		case "baza":
			m.baza = true
		case "mahram":
			m.mahram = true
		case "osta":
			m.osta = true
		default:
			return m
		}
		p.advance()
	}
}

func (p *Parser) parseSimpleStatement() ast.Stmt {
	start := p.start()

	if p.atVerb() && !p.at_(1).Kind.IsDeclaration() && !p.at_(1).Kind.IsAssignment() && !p.curIsMethodBuiltin() && !p.atBracketAssign() {
		return p.parseVerbCall(start)
	}

	if isTypeWord(p.cur()) {

		if !p.at_(1).Kind.IsDeclaration() {
			if n := p.tryParseTypedDecl(start); n != nil {
				return n
			}
		}
	}

	if p.atAny("(", token.IDENT) || isSoftName(p.cur()) {
		if n := p.tryParseNameList(start); n != nil {
			return n
		}
	}

	x := p.parseExpression()
	if x == nil {
		return nil
	}
	if p.atKindIsAssign() {
		return p.finishAssign(x, start)
	}
	p.endStatement()
	return &ast.ExprStmt{Loc: p.finish(start), X: x}
}

func (p *Parser) atVerb() bool {
	kw := p.curKeyword()
	return kw != nil && kw.IsVerb()
}

func (p *Parser) atBracketAssign() bool {
	if p.pos+1 >= len(p.toks) || p.toks[p.pos+1].Kind != "[" {
		return false
	}
	depth := 0
	for i := p.pos + 1; i < len(p.toks); i++ {
		switch p.toks[i].Kind {
		case "[", "(", "{":
			depth++
		case "]", ")", "}":
			depth--
			if depth == 0 {
				return i+1 < len(p.toks) && p.toks[i+1].Kind.IsAssignment()
			}
		case token.NEWLINE, token.EOF:
			return false
		}
	}
	return false
}

func (p *Parser) atBracketAssignHere() bool {
	if !p.at("[") {
		return false
	}
	depth := 0
	for i := p.pos; i < len(p.toks); i++ {
		switch p.toks[i].Kind {
		case "[", "(", "{":
			depth++
		case "]", ")", "}":
			depth--
			if depth == 0 {
				return i+1 < len(p.toks) && p.toks[i+1].Kind.IsAssignment()
			}
		case token.NEWLINE, token.EOF:
			return false
		}
	}
	return false
}

func (p *Parser) curIsMethodBuiltin() bool {
	kw := p.curKeyword()
	return kw != nil && kw.IsVerb() && kw.Method
}

func (p *Parser) curKeyword() *token.Keyword {
	if !p.cur().Kind.IsKeyword() {
		return nil
	}
	return token.LookupKeyword(p.cur().Lit)
}

func (p *Parser) startsExpression() bool {

	if p.at("agar") {
		return canBeginValue(p.at_(1)) && p.warnaOnThisLine()
	}
	return canBeginValue(p.cur())
}

func (p *Parser) warnaOnThisLine() bool {
	depth := 0
	for i := p.pos + 1; i < len(p.toks); i++ {
		switch p.toks[i].Kind {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			depth--
		case token.NEWLINE, token.EOF:
			if depth <= 0 {
				return false
			}
		case "warna":
			if depth <= 0 {
				return true
			}
		}
	}
	return false
}

func canBeginValue(t token.Token) bool {
	switch k := t.Kind; k {
	case token.NEWLINE, token.EOF, token.DEDENT, token.INDENT,
		token.INTERP_END, token.STRING_END, token.COMMENT, token.ILLEGAL,
		",", ";", ":", ")", "]", "}", "=", ":=", "=>", "->", "::", "?":
		return false
	case "-", "+", "~", "!", "na", "(", "[", "{":

		return true
	}
	if t.Kind.IsPunct() || t.Kind.IsAssignment() {
		return false
	}

	if t.Kind.IsOperator() {
		return false
	}

	if t.Kind == "funct" {
		return true
	}
	if t.Kind.IsKeyword() {
		kw := token.LookupKeyword(t.Lit)
		if kw == nil {
			return false
		}

		switch kw.Group {
		case token.GroupBuiltin, token.GroupType, token.GroupMisc, token.GroupOperator:

			return true
		}
		return false
	}
	return true
}

func canPrefix(k token.Kind) bool {
	switch k {
	case "-", "+", "~", "!", "na":
		return true
	}
	return false
}

func (p *Parser) parseBareArgs() []ast.Arg {
	var args []ast.Arg
	for p.startsExpression() {
		args = append(args, ast.Arg{Expr: p.parseExpression()})
		if !p.accept(",") {
			break
		}
	}
	return args
}

func (p *Parser) gapBefore(cur token.Token) bool {
	if p.pos == 0 {
		return false
	}
	return p.toks[p.pos-1].EndPos().Col < cur.Start().Col
}

func (p *Parser) parseVerbCall(start diag.Pos) ast.Stmt {
	v := p.advance()
	fn := &ast.Ident{Loc: ast.SpanOf(v.Start(), v.EndPos()), Name: v.Lit}
	call := &ast.Call{Loc: p.finish(start), Fn: fn}

	if p.at("(") && !p.gapBefore(p.cur()) {
		p.advance()
		call.Args = p.parseArgsUntil(")")
	} else {
		call.Args = p.parseBareArgs()
	}
	p.endStatement()
	call.Bare = true
	call.Loc = p.finish(start)
	return &ast.ExprStmt{Loc: p.finish(start), X: call}
}

func (p *Parser) atKindIsAssign() bool { return p.cur().Kind.IsAssignment() }

func (p *Parser) finishAssign(x ast.Expr, start diag.Pos) ast.Stmt {
	op := p.advance()
	if g, ok := x.(*ast.Group); ok {
		p.note(p.cur(), "E2005", "نمی‌شود به چند جعبه هم‌زمان مقدار داد",
			"you cannot assign to a group of values like this")
		_ = g
	}
	if !assignable(x) {
		p.bag.Add(diag.Errorf("E2006", x.Span(),
			"«%s» جایی نیست که بشود به آن مقدار داد", "you cannot assign to %s",
			describeTarget(x)).WithFile(p.file))
	}

	rhs := p.parseExpression()
	p.endStatement()

	return &ast.ExprStmt{Loc: p.finish(start), X: &ast.Assign{Loc: p.finish(start), Op: op.Kind, Target: x, Value: rhs}}
}

func assignable(x ast.Expr) bool {
	switch x.(type) {
	case *ast.Ident, *ast.Index, *ast.Member:
		return true
	}
	return false
}

func describeTarget(x ast.Expr) string {
	switch t := x.(type) {
	case *ast.Call:
		return "the result of a call"
	case *ast.Binary:
		return "a calculation"
	case *ast.StrLit, *ast.IntLit, *ast.FloatLit, *ast.BoolLit, *ast.ListLit, *ast.DictLit:
		return "a fixed value"
	case *ast.Unary:
		return "a sign or negation"
	default:
		_ = t
		return "this"
	}
}

func isTypeWord(t token.Token) bool {
	if t.Kind == token.IDENT {
		return false
	}
	kw := token.LookupKeyword(t.Lit)
	if kw == nil {
		return false
	}
	return kw.Group == token.GroupType && kw.Name != "gong" && kw.Name != "hich"
}

func isSoftName(t token.Token) bool {
	if t.Kind == token.IDENT {
		return true
	}
	if !t.Kind.IsKeyword() {
		return false
	}
	kw := token.LookupKeyword(t.Lit)
	if kw == nil {
		return false
	}
	switch kw.Group {
	case token.GroupBuiltin, token.GroupType:
		return true
	}
	return false
}

func (p *Parser) tryParseTypedDecl(start diag.Pos) ast.Stmt {
	save := p.pos
	typeName := p.advance().Lit
	if !isSoftName(p.cur()) {

		if p.startsExpression() || p.atAny(token.NEWLINE, ";", token.EOF, "}", token.DEDENT) {
			p.errf("E2025", "بعد از نام نوع باید یک نام بیاید",
				"after the type %s I need a name to give that type to", typeName)
		}
		p.pos = save
		return nil
	}
	if !p.at_(1).Kind.IsAssignment() {

		name := p.advance()
		p.endStatement()
		return &ast.VarDecl{
			Loc:      p.finish(start),
			TypeName: canonical(typeName),
			Name:     name.Lit,
			NamePos:  name.Span().Start,
		}
	}
	name := p.advance()
	p.advance()
	value := p.parseExpression()
	p.endStatement()
	return &ast.VarDecl{
		Loc:      p.finish(start),
		TypeName: canonical(typeName),
		Name:     name.Lit,
		NamePos:  name.Span().Start,
		Value:    value,
	}
}

func (p *Parser) tryParseNameList(start diag.Pos) ast.Stmt {
	save := p.pos
	var names []token.Token
	for {
		if !isSoftName(p.cur()) {
			p.pos = save
			return nil
		}
		names = append(names, p.advance())
		if !p.accept(",") {
			break
		}
	}
	if !p.cur().Kind.IsDeclaration() {
		p.pos = save
		return nil
	}
	p.advance()
	rhs := p.parseExpression()
	p.endStatement()

	if len(names) == 1 {
		return &ast.VarDecl{Loc: p.finish(start), Name: names[0].Lit, NamePos: names[0].Start(), Value: rhs}
	}
	values := unpackValues(rhs)
	if values == nil {
		p.errAt(p.at_(p.pos-1), "E2007",
			"برای چند نام باید به همان تعداد مقدار بدهی",
			"naming %d things needs %d values on the right", len(names), len(names))
		return nil
	}
	if len(values) != len(names) {
		p.errAt(p.at_(p.pos-1), "E2008",
			"تعداد نام‌ها با تعداد مقدارها یکی نیست",
			"you named %d things but supplied %d values", len(names), len(values))
	}
	var out multiDecl
	for i, n := range names {
		if i >= len(values) {
			break
		}
		out = append(out, &ast.VarDecl{
			Loc:     p.finish(start),
			Name:    n.Lit,
			NamePos: n.Span().Start,
			Value:   values[i],
		})
	}
	return out
}

func unpackValues(e ast.Expr) []ast.Expr {
	switch t := e.(type) {
	case *ast.Call:
		out := make([]ast.Expr, 0, len(t.Args))
		for _, a := range t.Args {
			out = append(out, a.Expr)
		}
		return out
	case *ast.ListLit:
		return t.Elems
	case *ast.Group:
		return unpackValues(t.X)
	}
	return nil
}

func canonical(name string) string {
	if kw := token.LookupKeyword(name); kw != nil {
		return kw.Name
	}
	return name
}

func (p *Parser) parseIf() ast.Stmt {
	start := p.start()
	p.advance()
	cond := p.parseExpression()
	then := p.parseBlock()
	st := &ast.IfStmt{Loc: p.finish(start), Cond: cond, Then: then}
	if p.at("warna") {
		p.advance()
		if p.at("agar") {
			st.Else = p.parseIf()
		} else {
			st.Else = p.parseBlock()
		}
	}
	st.Loc = p.finish(start)
	return st
}

func (p *Parser) parseWhile() ast.Stmt {
	start := p.start()
	p.advance()
	cond := p.parseExpression()
	body := p.parseBlock()
	return &ast.WhileStmt{Loc: p.finish(start), Cond: cond, Body: body}
}

func (p *Parser) parseFor() ast.Stmt {
	start := p.start()
	p.advance()
	if !p.at(token.IDENT) {
		p.errf("E2010", "پس از `har` باید یک نام بیاید", "a for loop needs a name for the current item")
		return &ast.ForStmt{Loc: p.finish(start), Body: p.parseBlock()}
	}
	name := p.advance()

	if p.at("az") {
		p.advance()
		lo := p.parseExpression()
		if !p.at("ta") {
			p.errf("E2009", "پس از `az` باید `ta` بیاید", "after az I need ta to say where the range ends")
			return nil
		}
		p.advance()
		inclusive := false
		if p.at("+") {

			p.advance()
			inclusive = true
		}
		p.plusEndsHeader = true
		hi := p.parseExpression()
		p.plusEndsHeader = false
		if p.at("+") {
			p.advance()
			inclusive = true
		}
		filter := p.parseForFilter()
		body := p.parseBlock()
		return &ast.ForStmt{
			Loc: p.finish(start), VarName: name.Lit, VarPos: name.Span().Start,
			Lo: lo, Hi: hi, Inclusive: inclusive, Filter: filter, Body: body,
		}
	}

	if p.at("dar") {
		p.advance()
		iter := p.parseExpression()
		filter := p.parseForFilter()
		body := p.parseBlock()
		return &ast.ForInStmt{
			Loc: p.finish(start), VarName: name.Lit, VarPos: name.Span().Start,
			Iter: iter, Filter: filter, Body: body,
		}
	}
	p.errf("E2010", "پس از نام حلقه یک `az` یا `dar` بیاور",
		"a for loop needs az for a range or dar to walk a collection")
	return nil
}

func (p *Parser) parseForFilter() ast.Expr {
	if !p.at("ke") {
		return nil
	}
	p.advance()
	return p.parseExpression()
}

func (p *Parser) parseReturn() ast.Stmt {
	start := p.start()
	p.advance()
	var v ast.Expr
	if !p.atAny(token.NEWLINE, ";", token.EOF, "}", token.DEDENT) {
		v = p.parseExpression()
	}
	p.endStatement()
	return &ast.ReturnStmt{Loc: p.finish(start), Value: v}
}

func (p *Parser) parseThrow() ast.Stmt {
	start := p.start()
	p.advance()
	v := p.parseExpression()
	p.endStatement()
	return &ast.ThrowStmt{Loc: p.finish(start), Value: v}
}

func (p *Parser) parseTry() ast.Stmt {
	start := p.start()
	p.advance()
	body := p.parseBlock()
	st := &ast.TryStmt{Loc: p.finish(start), Body: body}
	if p.at("giroft") {
		p.advance()
		c := &ast.Catch{}
		if isTypeWord(p.cur()) && p.at_(1).Kind == token.IDENT {
			c.TypeName = canonical(p.advance().Lit)
		}
		if p.at(token.IDENT) {
			t := p.advance()
			c.Name = t.Lit
			c.NamePos = t.Span().Start
		}
		c.Body = p.parseBlock()
		st.Catch, st.HasCatch = c, true
	}
	if p.at("nemigir") {
		p.advance()
		st.Finally = p.parseBlock()
	}
	st.Loc = p.finish(start)
	return st
}

func (p *Parser) parseBlock() *ast.Block {
	p.accept(":")
	if p.at("{") {
		return p.parseBracedBlock()
	}
	return p.parseIndentedBlock()
}

func (p *Parser) parseBracedBlock() *ast.Block {
	start := p.start()
	p.expect("{", "an opening brace")
	stmts := p.parseStatementList()
	p.expect("}", "a closing brace")
	return &ast.Block{Loc: p.finish(start), Stmts: stmts, Braced: true}
}

func (p *Parser) parseIndentedBlock() *ast.Block {
	start := p.start()
	if !p.at(token.NEWLINE) {

		s := p.parseStatement()
		b := &ast.Block{Loc: p.finish(start)}
		if m, ok := s.(multiDecl); ok {
			b.Stmts = m
		} else if s != nil {
			b.Stmts = append(b.Stmts, s)
		}
		return b
	}
	p.advance()
	if !p.at(token.INDENT) {
		p.errf("E2011", "بدنه باید یک خط جلوتر و تورفته بنشیند",
			"this block has no body; put its statements on the next line, indented")
		return &ast.Block{Loc: p.finish(start)}
	}
	p.advance()
	stmts := p.parseStatementList()
	p.expect(token.DEDENT, "the end of the indented block")
	return &ast.Block{Loc: p.finish(start), Stmts: stmts}
}

func (p *Parser) parseFuncDecl(m *mods) ast.Stmt {
	start := p.start()
	p.advance()

	var generics []string
	if p.at("<") {
		p.advance()
		for p.at(token.IDENT) {
			generics = append(generics, p.advance().Lit)
			if !p.accept(",") {
				break
			}
		}
		p.expect(">", "the end of the list of type names")
	}

	if !isSoftName(p.cur()) {
		p.errf("E2026", "پس از funct باید یک نام بیاید", "a funct needs a name after it")
	}
	p.generics = generics
	nameTok := p.advance()
	params := p.parseParams()
	result := p.parseResultType()

	p.accept(":")
	p.accept(";")
	var body *ast.Block
	switch {
	case p.at("{"):
		body = p.parseBracedBlock()
	case p.at(token.NEWLINE):
		body = p.parseIndentedBlock()
	default:

		body = &ast.Block{Loc: p.finish(start)}
	}

	f := &ast.FuncDecl{
		Loc: p.finish(start), Name: nameTok.Lit, NamePos: nameTok.Span().Start,
		Params: params, Result: result, Body: body, Generics: generics,
	}
	if m != nil {
		f.Baza, f.Mahram, f.Osta = m.baza, m.mahram, m.osta
	}
	return f
}

func (p *Parser) parseParams() []*ast.Param {
	if !p.at("(") {
		p.errf("E2012", "پس از نام تابع باید `(` بیاید", "a funct needs a parameter list in parentheses")
		return nil
	}
	p.advance()
	var params []*ast.Param
	p.skipSeps()
	for !p.at(")") && !p.at(token.EOF) {

		typeFirst := ""
		if isTypeWord(p.cur()) && isSoftName(p.at_(1)) {
			typeFirst = p.advance().Lit
		}
		if !isSoftName(p.cur()) {
			p.errf("E2013", "هر پارامتر باید اسم داشته باشد", "every parameter needs a name, not just a type")
			break
		}
		t := p.advance()
		pm := &ast.Param{Name: t.Lit, NamePos: t.Span().Start, TypeName: typeFirst}
		if typeFirst == "" && p.accept(":") {
			pm.TypeName = p.parseTypeName()
		}
		if p.accept("=") {
			pm.Default = p.parseExpression()
		}
		params = append(params, pm)
		if !p.accept(",") {
			break
		}
		p.skipSeps()
	}
	p.skipSeps()
	p.expect(")", "a closing parenthesis")
	return params
}

func (p *Parser) parseResultType() string {
	if !p.accept("->") {
		return ""
	}

	if p.isTypeParameterHere() {
		return p.parseTypeName()
	}
	switch {
	case p.at(token.IDENT), p.at(token.INT), p.at(token.FLOAT), p.at(token.STRING),
		p.cur().Kind.IsOperator(), p.cur().Kind == "(", p.cur().Kind == "[":
		p.errf("E2027", "بعد از `->` باید نام نوع بیاید",
			"`->` is followed by the type a funct gives back, not the value; use bargard inside the body to send a value back")
	}
	return p.parseTypeName()
}

func (p *Parser) isTypeParameterHere() bool {
	if !p.at(token.IDENT) {
		return false
	}
	for _, g := range p.generics {
		if g == p.cur().Lit {
			return true
		}
	}
	return false
}

func (p *Parser) parseTypeName() string {
	var b []rune
	if p.at(token.IDENT) || p.cur().Kind.IsKeyword() {
		t := p.advance()

		b = append(b, []rune(canonical(t.Lit))...)
	} else {
		p.errf("E2014", "اینجا نام نوع لازم است", "I need a type name here")
		return ""
	}
	if p.at("<") {
		p.advance()
		var inner []rune
		for !p.at(">") && !p.at(token.EOF) {
			inner = append(inner, []rune(p.parseTypeName())...)
			if !p.accept(",") {
				break
			}
		}
		p.expect(">", "the end of the type parameters")
		b = append(b, '<')
		b = append(b, inner...)
		b = append(b, '>')
	}
	return string(b)
}

func (p *Parser) parseImport() ast.Stmt {
	start := p.start()
	p.advance()
	t, _ := p.expect(token.STRING, "the module path")
	p.endStatement()
	d := &ast.ImportDecl{Loc: p.finish(start), Path: t.Text}

	if p.at(token.IDENT) && p.cur().Lit == "as" {
		p.advance()
		if a, ok := p.expect(token.IDENT, "a short name for the module"); ok {
			d.Alias = a.Lit
		}
	}
	return d
}

func (p *Parser) parseInterfaceDecl() ast.Stmt {
	start := p.start()
	p.advance()
	name, _ := p.expect(token.IDENT, "the name of the dastoor")
	d := &ast.InterfaceDecl{Loc: p.finish(start), Name: name.Lit, NamePos: name.Span().Start}
	if p.at("{") {
		p.advance()
		for !p.at("}") && !p.at(token.EOF) {
			if p.at(token.IDENT) {
				d.Methods = append(d.Methods, p.advance().Lit)
				p.accept("(")
				p.skipTo(")")
			}
			p.skipSeps()
		}
		p.expect("}", "a closing brace")
	} else {
		p.accept(":")

		body := p.parseClassBody()
		if body != nil {
			for _, s := range body.Stmts {
				if fs, ok := s.(*ast.FuncDecl); ok {
					d.Methods = append(d.Methods, fs.Name)
				}
			}
		}
	}
	d.Loc = p.finish(start)
	return d
}

func (p *Parser) skipTo(k token.Kind) {
	for !p.at(token.EOF) {
		if p.at(k) {
			p.advance()
			return
		}
		p.advance()
	}
}

func (p *Parser) parseClassDecl(m *mods) ast.Stmt {
	start := p.start()
	p.advance()
	name, _ := p.expect(token.IDENT, "the name of the darghan")
	c := &ast.ClassDecl{Loc: p.finish(start), Name: name.Lit, NamePos: name.Span().Start, CtorName: "__sakht__"}
	if p.accept("<") {
		for p.at(token.IDENT) {
			c.Bases = append(c.Bases, p.advance().Lit)
			if !p.accept(",") {
				break
			}
		}

		p.accept(">")
	}
	p.accept(":")
	c.Fields, c.Methods = p.collectClassBody(p.parseClassBody(), "darghan")
	c.Loc = p.finish(start)
	if m != nil {
		c.Baza, c.Mahram = m.baza, m.mahram
	}
	return c
}

func (p *Parser) parseStructDecl(m *mods) ast.Stmt {
	start := p.start()
	p.advance()
	name, _ := p.expect(token.IDENT, "the name of the sakht")
	s := &ast.StructDecl{Loc: p.finish(start), Name: name.Lit, NamePos: name.Span().Start}
	if p.accept("<") {
		for p.at(token.IDENT) {
			s.Bases = append(s.Bases, p.advance().Lit)
			if !p.accept(",") {
				break
			}
		}
		p.expect(">", "the end of the list of base names")
	}
	p.accept(":")
	s.Fields, s.Methods = p.collectClassBody(p.parseClassBody(), "sakht")

	for _, meth := range s.Methods {
		p.bag.Add(diag.New("E2015", meth.Span(),
			"ساخت جای متد ندارد؛ برای این کار `darghan` بنویس",
			fmt.Sprintf("a sakht cannot hold the method %q; write a darghan instead", meth.Name)).
			WithNote("a sakht is for values that are just data, like a point or a date").
			WithFile(p.file))
	}
	s.Loc = p.finish(start)
	if m != nil {
		s.Baza, s.Mahram = m.baza, m.mahram
	}
	return s
}

func (p *Parser) collectClassBody(body *ast.Block, what string) ([]*ast.FieldDecl, []*ast.FuncDecl) {
	if body == nil {
		return nil, nil
	}
	var fields []*ast.FieldDecl
	var methods []*ast.FuncDecl
	for _, st := range body.Stmts {
		switch v := st.(type) {
		case *ast.FieldDecl:
			fields = append(fields, v)
		case *ast.FuncDecl:
			methods = append(methods, v)
		default:
			p.bag.Add(diag.New("E2019", st.Span(),
				fmt.Sprintf("داخل %s فقط فیلد و متد می‌آید", what),
				fmt.Sprintf("a %s body holds only fields and methods, not %s", what, st.NodeKind())).WithFile(p.file))
		}
	}
	return fields, methods
}

func (p *Parser) parseEnumDecl() ast.Stmt {
	start := p.start()
	p.advance()
	name, _ := p.expect(token.IDENT, "the name of the shenase")
	e := &ast.EnumDecl{Loc: p.finish(start), Name: name.Lit, NamePos: name.Span().Start}
	p.accept(":")
	p.inEnum = true
	body := p.parseClassBody()
	p.inEnum = false
	if body != nil {
		for _, st := range body.Stmts {
			ev, ok := st.(*ast.EnumValue)
			if !ok {
				p.bag.Add(diag.New("E2016", st.Span(),
					"داخل `shenase` فقط نام مقدارها می‌آید",
					fmt.Sprintf("a shenase body holds only named values, not %s", st.NodeKind())).WithFile(p.file))
				continue
			}
			e.Values = append(e.Values, ev)
		}
	}
	e.Loc = p.finish(start)
	return e
}

func (p *Parser) parseClassBody() *ast.Block {
	start := p.start()
	if p.at("{") {
		return p.parseBracedBlock()
	}
	if !p.at(token.NEWLINE) {
		s := p.parseClassMember()
		b := &ast.Block{Loc: p.finish(start)}
		if s != nil {
			b.Stmts = append(b.Stmts, s)
		}
		return b
	}
	p.advance()
	if !p.at(token.INDENT) {
		p.errf("E2017", "بدنه خالی است", "this declaration has no body; indent its members")
		return &ast.Block{Loc: p.finish(start)}
	}
	p.advance()
	var stmts []ast.Stmt
	for {
		p.skipSeps()
		if p.atAny(token.EOF, token.DEDENT, "}") {
			break
		}
		before := p.pos
		if s := p.parseClassMember(); s != nil {
			stmts = append(stmts, s)
		}
		if p.pos == before {
			break
		}
	}
	p.expect(token.DEDENT, "the end of the indented body")
	return &ast.Block{Loc: p.finish(start), Stmts: stmts}
}

func (p *Parser) parseClassMember() ast.Stmt {
	p.depth++
	defer func() { p.depth-- }()
	start := p.start()
	m := p.readMods()

	if p.at("funct") {
		return p.parseFuncDecl(&m)
	}

	if p.inEnum && p.at(token.IDENT) && p.at_(1).Kind.IsDeclaration() {
		nameTok := p.advance()
		p.advance()
		v := p.parseExpression()
		p.endStatement()
		return &ast.EnumValue{Loc: p.finish(start), Name: nameTok.Lit, NamePos: nameTok.Span().Start, Value: v}
	}

	if !p.at(token.IDENT) {
		p.errf("E2018", "اینجا فقط فیلد یا متد می‌آید", "only fields and methods can go here")
		return nil
	}
	nameTok := p.advance()

	if p.at("(") && !p.inEnum {
		p.advance()
		var params []*ast.Param
		p.skipSeps()
		for !p.at(")") && !p.at(token.EOF) {
			if !p.at(token.IDENT) {
				p.errf("E2013", "هر پارامتر باید اسم داشته باشد", "every parameter needs a name, not just a type")
				break
			}
			t := p.advance()
			pm := &ast.Param{Name: t.Lit, NamePos: t.Span().Start}
			if p.accept(":") {
				pm.TypeName = p.parseTypeName()
			}
			params = append(params, pm)
			if !p.accept(",") {
				break
			}
			p.skipSeps()
		}
		p.skipSeps()
		p.expect(")", "a closing parenthesis")
		result := p.parseResultType()

		if p.at(token.NEWLINE) && p.at_(1).Kind == token.INDENT {
			body := p.parseIndentedBlock()
			return &ast.FuncDecl{
				Loc: p.finish(start), Name: nameTok.Lit, NamePos: nameTok.Span().Start,
				Params: params, Result: result, Body: body,
				Baza: m.baza, Mahram: m.mahram,
			}
		}
		p.endStatement()
		return &ast.FuncDecl{
			Loc: p.finish(start), Name: nameTok.Lit, NamePos: nameTok.Span().Start,
			Params: params, Result: result, Body: &ast.Block{Loc: p.finish(start)},
			Baza: m.baza, Mahram: m.mahram,
		}
	}

	f := &ast.FieldDecl{
		Loc: p.finish(start), Name: nameTok.Lit, NamePos: nameTok.Span().Start,
		Baza: m.baza, Mahram: m.mahram, Osta: m.osta, Sabit: m.sabit,
	}
	if p.accept(":") {
		f.TypeName = p.parseTypeName()
	}
	if p.atKindIsAssign() {
		p.advance()
		f.Default = p.parseExpression()
	}
	p.endStatement()
	f.Loc = p.finish(start)
	return f
}

func (p *Parser) parseExpression() ast.Expr {

	if p.at("funct") {
		if lam := p.tryParseLambda(); lam != nil {
			if p.atKindIsAssign() {
				return p.assignAsExpr(lam, p.start())
			}
			return lam
		}
	}
	x := p.parseBinary(1)
	if p.atKindIsAssign() {
		return p.assignAsExpr(x, p.start())
	}
	return x
}

func (p *Parser) assignAsExpr(x ast.Expr, start diag.Pos) ast.Expr {
	op := p.advance()
	if !assignable(x) {
		p.bag.Add(diag.Errorf("E2006", x.Span(),
			"«%s» جایی نیست که بشود به آن مقدار داد", "you cannot assign to %s",
			describeTarget(x)).WithFile(p.file))
	}
	return &ast.Assign{Loc: p.finish(start), Op: op.Kind, Target: x, Value: p.parseExpression()}
}

func (p *Parser) parseBinary(minPrec int) ast.Expr {
	left := p.parseUnary()
	for {
		op := p.cur().Kind
		if op.IsPunct() || !isBinaryOp(op) {
			return left
		}

		if op == "+" && p.plusEndsHeader {
			if k := p.at_(1).Kind; k == token.NEWLINE || k == token.DEDENT || k == token.EOF {
				return left
			}
		}
		prec := token.Precedence(op)
		if prec < minPrec {
			return left
		}
		opTok := p.advance()
		var right ast.Expr
		if token.AssocOf(op) == token.Right {
			right = p.parseBinary(prec)
		} else {
			right = p.parseBinary(prec + 1)
		}
		loc := ast.SpanOf(left.Span().Start, right.Span().End)
		if op == ".." || op == "..+" {
			inc := op == "..+"
			left = &ast.RangeExpr{Loc: loc, Lo: left, Hi: right, Inclusive: inc}
			continue
		}
		left = &ast.Binary{Loc: loc, Op: op, OpPos: opTok.Span().Start, L: left, R: right}
	}
}

func isBinaryOp(k token.Kind) bool {
	if k.IsPunct() || k.IsAssignment() {
		return false
	}
	return token.Precedence(k) > 0 && token.Precedence(k) < 15
}

func (p *Parser) parseUnary() ast.Expr {
	k := p.cur().Kind
	switch k {
	case "na", "!":
		t := p.advance()

		x := p.parseBinary(token.Precedence("va") + 1)
		return &ast.Unary{Loc: ast.SpanOf(t.Span().Start, endOf(x)), Op: k, X: x}
	case "-", "+", "~":
		t := p.advance()
		x := p.parseBinary(prefixTight)
		return &ast.Unary{Loc: ast.SpanOf(t.Span().Start, endOf(x)), Op: k, X: x}
	}
	return p.parsePostfix(p.parsePrimary())
}

func endOf(e ast.Expr) diag.Pos {
	if e == nil {
		return diag.Pos{}
	}
	return e.Span().End
}

func (p *Parser) parsePostfix(x ast.Expr) ast.Expr {
	keywordName := false
	if id, ok := x.(*ast.Ident); ok {
		if kw := token.LookupKeyword(id.Name); kw != nil {
			keywordName = kw.IsVerb()
		}
	}
	for {
		switch {
		case p.at("("):
			call, ok := p.parseCallArgs(x)
			if !ok {
				return x
			}
			x = call
			keywordName = false
		case p.at("["):
			if keywordName && !p.atBracketAssignHere() {

				arg := p.parseListLit()
				x = &ast.Call{Loc: ast.SpanOf(x.Span().Start, p.prevEnd()), Fn: x, Args: []ast.Arg{{Expr: arg}}}
				keywordName = false
				continue
			}
			p.advance()
			x = p.parseIndexOrSlice(x)
		case p.at("{") && keywordName:

			arg := p.parseBraceLit()
			x = &ast.Call{Loc: ast.SpanOf(x.Span().Start, p.prevEnd()), Fn: x, Args: []ast.Arg{{Expr: arg}}}
			keywordName = false
		case p.at("."):
			p.advance()
			m, ok := p.parseMember(x)
			if !ok {
				return x
			}
			x = m
		case p.at("?") && p.at_(1).Kind == "?":

			p.advance()
			p.advance()
			rhs := p.parseExpression()
			fn := &ast.Ident{Loc: x.Span(), Name: "pishfarz"}
			x = &ast.Call{Loc: ast.SpanOf(x.Span().Start, p.prevEnd()), Fn: fn, Args: []ast.Arg{{Expr: x}, {Expr: rhs}}}
			return x
		case p.at("?"):

			p.advance()
		default:
			return x
		}
	}
}

func (p *Parser) parseCallArgs(fn ast.Expr) (ast.Expr, bool) {
	start := fn.Span().Start
	p.advance()
	args := p.parseArgsUntil(")")
	return &ast.Call{Loc: ast.SpanOf(start, p.prevEnd()), Fn: fn, Args: args}, true
}

func (p *Parser) parseArgsUntil(closing token.Kind) []ast.Arg {
	var args []ast.Arg
	p.skipSeps()
	for !p.at(closing) && !p.at(token.EOF) {
		a := ast.Arg{}

		if p.at(token.IDENT) && (p.at_(1).Kind == ":" && !p.at_(2).Kind.IsPunct() || p.at_(1).Kind == "=") {
			a.Name = p.advance().Lit
			p.advance()
		}
		a.Expr = p.parseExpression()
		args = append(args, a)
		if !p.accept(",") {
			break
		}
		p.skipSeps()
	}
	p.skipSeps()
	p.expect(closing, "a closing bracket")
	return args
}

func (p *Parser) parseIndexOrSlice(x ast.Expr) ast.Expr {
	start := x.Span().Start
	if p.at("]") {
		p.advance()
		return &ast.Index{Loc: ast.SpanOf(start, p.prevEnd()), X: x, At: &ast.ListLit{Loc: p.finish(start)}}
	}

	if p.at(":") {
		p.advance()
		var hi ast.Expr
		if !p.at("]") {
			hi = p.parseExpression()
		}
		p.expect("]", "a closing bracket")
		return &ast.Slice{Loc: ast.SpanOf(start, p.prevEnd()), X: x, Hi: hi}
	}

	if p.at("..") {
		p.advance()
		if p.at("+") {
			p.advance()
		}
		var hi ast.Expr
		if !p.at("]") {
			hi = p.parseExpression()
		}
		p.expect("]", "a closing bracket")
		return &ast.Slice{Loc: ast.SpanOf(start, p.prevEnd()), X: x, Hi: hi, Inclusive: true}
	}
	first := p.parseExpression()

	if p.at(":") {
		p.advance()
		var hi ast.Expr
		if !p.at("]") {
			hi = p.parseExpression()
		}
		p.expect("]", "a closing bracket")
		return &ast.Slice{Loc: ast.SpanOf(start, p.prevEnd()), X: x, Lo: first, Hi: hi}
	}
	if p.at("..") {
		p.advance()

		if p.at("+") {
			p.advance()
		}
		hi := p.parseExpression()
		p.expect("]", "a closing bracket")
		return &ast.Slice{Loc: ast.SpanOf(start, p.prevEnd()), X: x, Lo: first, Hi: hi, Inclusive: true}
	}

	if r, ok := first.(*ast.RangeExpr); ok {
		p.expect("]", "a closing bracket")
		return &ast.Slice{Loc: ast.SpanOf(start, p.prevEnd()), X: x, Lo: r.Lo, Hi: r.Hi, Inclusive: true}
	}
	p.expect("]", "a closing bracket")
	return &ast.Index{Loc: ast.SpanOf(start, p.prevEnd()), X: x, At: first}
}

func (p *Parser) parseMember(x ast.Expr) (ast.Expr, bool) {
	if !p.at(token.IDENT) && !p.cur().Kind.IsKeyword() {
		p.errf("E2020", "پس از نقطه یک نام لازم است", "a field or method name must follow the dot")
		return x, false
	}
	nameTok := p.advance()
	m := &ast.Member{
		Loc: ast.SpanOf(x.Span().Start, nameTok.EndPos()), X: x,
		Name: nameTok.Lit, NamePos: nameTok.Span().Start,
	}
	if p.at("(") {
		p.advance()
		m.Args = p.parseArgsUntil(")")
		m.Call = true
		m.Loc = ast.SpanOf(x.Span().Start, p.prevEnd())
	}
	return m, true
}

func (p *Parser) prevEnd() diag.Pos {
	if p.pos == 0 {
		return p.cur().Start()
	}
	return p.at_(p.pos - 1).EndPos()
}

func (p *Parser) parsePrimary() ast.Expr {
	t := p.cur()
	start := t.Span().Start
	switch t.Kind {
	case token.INT:
		p.advance()
		n, err := strconv.ParseInt(t.Text, 10, 64)
		if err != nil {
			u, uerr := strconv.ParseUint(t.Text, 10, 64)
			if uerr != nil {
				p.errf("E2021", "عدد %s از اندازه arz بزرگ‌تر است", "%s is larger than arz can hold", t.Text)
				n = 0
			} else {
				n = int64(u)
			}
		}
		return &ast.IntLit{Loc: ast.SpanOf(start, t.EndPos()), Value: n, Raw: t.Text}
	case token.FLOAT:
		p.advance()
		f, _ := strconv.ParseFloat(t.Text, 64)
		return &ast.FloatLit{Loc: ast.SpanOf(start, t.EndPos()), Value: f, Raw: t.Text}
	case token.STRING:
		p.advance()
		return &ast.StrLit{Loc: ast.SpanOf(start, t.EndPos()), Value: t.Text}
	case token.CHAR:
		p.advance()
		r := []rune(t.Text)
		var c rune
		if len(r) > 0 {
			c = r[0]
		}
		return &ast.CharLit{Loc: ast.SpanOf(start, t.EndPos()), Value: c}
	case token.INTERP:
		p.advance()
		return p.parseInterp(t)
	case "dorost":
		p.advance()
		return &ast.BoolLit{Loc: ast.SpanOf(start, t.EndPos()), Value: true}
	case "ghalat":
		p.advance()
		return &ast.BoolLit{Loc: ast.SpanOf(start, t.EndPos()), Value: false}
	case "hich":
		p.advance()
		return &ast.HichLit{Loc: ast.SpanOf(start, t.EndPos())}
	case token.IDENT:
		p.advance()
		return &ast.Ident{Loc: ast.SpanOf(start, t.EndPos()), Name: t.Lit}
	case "(":
		p.advance()
		x := p.parseExpression()
		p.expect(")", "a closing parenthesis")
		return &ast.Group{Loc: ast.SpanOf(start, p.prevEnd()), X: x}
	case "[":
		return p.parseListLit()
	case "{":
		return p.parseBraceLit()
	case "funct":
		if lam := p.tryParseLambda(); lam != nil {
			return lam
		}
		p.advance()
		return &ast.Ident{Loc: ast.SpanOf(start, t.EndPos()), Name: t.Lit}
	case "agar":

		p.advance()
		cond := p.parseExpression()
		then := p.parseExpression()
		if !p.at("warna") {
			p.errf("E2022", "پس از شرطِ `agar` باید `warna` بیاید",
				"a one-line agar needs a warna in the middle: agar cond then warna otherwise")
			return &ast.ErrorExpr{Loc: ast.SpanOf(start, p.prevEnd()), Message: "missing warna"}
		}
		p.advance()
		els := p.parseExpression()
		return &ast.Ternary{Loc: ast.SpanOf(start, endOf(els)), Cond: cond, Then: then, Else: els}
	}
	if isTypeWord(t) {
		p.advance()
		return &ast.Ident{Loc: ast.SpanOf(start, t.EndPos()), Name: t.Lit}
	}

	if t.Kind.IsKeyword() {
		if kw := token.LookupKeyword(t.Lit); kw != nil &&
			(kw.Group == token.GroupBuiltin || kw.Group == token.GroupType || kw.Group == token.GroupOperator) {
			p.advance()
			id := &ast.Ident{Loc: ast.SpanOf(start, t.EndPos()), Name: t.Lit}

			if p.atAny("=", ":=", "+=", "-=", "*=", "/=", "%=", "**=", "&=", "|=", "^=", "<<=", ">>=") {
				return id
			}
			if kw.IsVerb() && !kw.Method && p.startsExpression() && !p.atAny("(", "[", "{") {

				if canPrefix(p.cur().Kind) && p.cur().Col > t.EndPos().Col {
					return id
				}
				args := p.parseBareArgs()
				return &ast.Call{Loc: ast.SpanOf(start, p.prevEnd()), Fn: id, Args: args, Bare: true}
			}
			return id
		}
	}
	p.errfArgs("E2023",
		"%s اینجا نمی‌تواند بنشیند", []any{t.DescribeFarsi()},
		"%s cannot be used where a value is expected", []any{t.Describe()})
	p.advance()
	return &ast.ErrorExpr{Loc: ast.SpanOf(start, p.prevEnd()), Message: t.DescribeFarsi(), English: t.Describe()}
}

func (p *Parser) parseInterp(t token.Token) ast.Expr {
	lit := &ast.InterpLit{Loc: ast.SpanOf(t.Span().Start, t.EndPos())}
	for {
		switch p.cur().Kind {
		case token.STRING_END:
			p.advance()
			lit.Loc = ast.SpanOf(t.Span().Start, p.prevEnd())
			return lit
		case token.INTERP_TEXT:
			tk := p.advance()
			lit.Segs = append(lit.Segs, ast.InterpSeg{
				Loc:  ast.SpanOf(tk.Span().Start, tk.EndPos()),
				Text: tk.Text,
			})
		case token.INTERP_END:

			q := p.advance()
			p.bag.Add(diag.Errorf("E2024", ast.SpanOf(q.Span().Start, q.EndPos()),
				"داخل «${}» هیچ عبارتی نیست", "there is no expression between ${ and }").WithFile(p.file))
			lit.Segs = append(lit.Segs, ast.InterpSeg{Loc: ast.SpanOf(q.Span().Start, q.EndPos()), Expr: nil})
		case token.EOF:
			p.errf("E1005", "رشته بسته نشد", "this string is never closed")
			return lit
		default:
			sp := p.start()
			x := p.parseExpression()
			p.expect(token.INTERP_END, "a } to close the ${")
			lit.Segs = append(lit.Segs, ast.InterpSeg{
				Loc:  ast.SpanOf(sp, p.prevEnd()),
				Expr: x,
			})
		}
	}
}
func runeLen(s string) int { return len([]rune(s)) }

func (p *Parser) parseListLit() ast.Expr {
	start := p.start()
	p.advance()
	l := &ast.ListLit{}
	p.skipSeps()
	for !p.at("]") && !p.at(token.EOF) {
		l.Elems = append(l.Elems, p.parseExpression())
		if len(l.Elems) == 1 && p.at("har") {
			return p.parseListComp(start, l.Elems[0])
		}
		if !p.accept(",") {
			break
		}
		p.skipSeps()
	}
	p.skipSeps()
	p.expect("]", "a closing bracket")
	l.Loc = p.finish(start)
	return l
}

func (p *Parser) parseListComp(start diag.Pos, elem ast.Expr) ast.Expr {
	p.advance()
	c := &ast.ListComp{Elem: elem}
	if !p.at(token.IDENT) {
		p.errf("E2010", "پس از `har` باید یک نام بیاید", "a for loop needs a name for the current item")
		return c
	}
	name := p.advance()
	c.VarName, c.VarPos = name.Lit, name.Span().Start
	switch {
	case p.at("dar"):
		p.advance()
		c.Iter = p.parseExpression()
	case p.at("az"):
		rs := p.start()
		p.advance()
		lo := p.parseExpression()
		if !p.at("ta") {
			p.errf("E2009", "پس از `az` باید `ta` بیاید", "after az I need ta to say where the range ends")
			return c
		}
		p.advance()
		hi := p.parseExpression()
		inclusive := p.accept("+")
		c.Iter = &ast.RangeExpr{Loc: p.finish(rs), Lo: lo, Hi: hi, Inclusive: inclusive}
	default:
		p.errf("E2028", "در فهرست‌سازی پس از نام باید `dar` یا `az` بیاید",
			"in [x har x dar xs] the name must be followed by dar or az")
		return c
	}
	if p.at("ke") || p.at("agar") {
		p.advance()
		c.Filter = p.parseExpression()
	}
	p.skipSeps()
	p.expect("]", "a closing bracket")
	c.Loc = p.finish(start)
	return c
}

func (p *Parser) parseBraceLit() ast.Expr {
	start := p.start()
	p.advance()
	p.skipSeps()
	if p.at("}") {
		p.advance()
		return &ast.DictLit{Loc: p.finish(start)}
	}
	first := p.parseExpression()
	if p.at(":") {
		p.advance()
		val := p.parseExpression()
		d := &ast.DictLit{Entries: []ast.DictEntry{{Key: first, Value: val}}}
		for {
			p.skipSeps()
			if !p.accept(",") {
				break
			}
			p.skipSeps()
			if p.at("}") {
				break
			}
			k := p.parseExpression()
			p.expect(":", "a colon between the key and the value")
			v := p.parseExpression()
			d.Entries = append(d.Entries, ast.DictEntry{Key: k, Value: v})
		}
		p.skipSeps()
		p.expect("}", "a closing brace")
		d.Loc = p.finish(start)
		return d
	}
	s := &ast.SetLit{Elems: []ast.Expr{first}}
	for {
		p.skipSeps()
		if !p.accept(",") {
			break
		}
		p.skipSeps()
		if p.at("}") {
			break
		}
		s.Elems = append(s.Elems, p.parseExpression())
	}
	p.skipSeps()
	p.expect("}", "a closing brace")
	s.Loc = p.finish(start)
	return s
}

func (p *Parser) tryParseLambda() *ast.Lambda {
	save := p.pos
	start := p.start()
	p.advance()
	if !p.at("(") {
		p.pos = save
		return nil
	}
	p.advance()
	var params []*ast.Param
	p.skipSeps()
	for !p.at(")") && !p.at(token.EOF) {
		if !p.at(token.IDENT) {
			p.pos = save
			return nil
		}
		t := p.advance()
		pm := &ast.Param{Name: t.Lit, NamePos: t.Span().Start}
		if p.accept(":") {
			pm.TypeName = p.parseTypeName()
		}
		if p.accept("=") {
			pm.Default = p.parseExpression()
		}
		params = append(params, pm)
		if !p.accept(",") {
			break
		}
		p.skipSeps()
	}
	if !p.at(")") {
		p.pos = save
		return nil
	}
	p.advance()
	lam := &ast.Lambda{Loc: p.finish(start), Params: params}
	if p.accept("->") {

		if p.at(token.NEWLINE) {
			lam.Body = p.parseIndentedBlock()
			lam.Loc = p.finish(start)
			return lam
		}
		lam.Expr = p.parseExpression()
		lam.Loc = p.finish(start)
		return lam
	}
	if p.accept("=>") {
		lam.Expr = p.parseExpression()
		lam.Loc = p.finish(start)
		return lam
	}
	p.accept(":")
	switch {
	case p.at("{"):
		lam.Body = p.parseBracedBlock()
	case p.at(token.NEWLINE):
		lam.Body = p.parseIndentedBlock()
	default:

		p.pos = save
		return nil
	}
	lam.Loc = p.finish(start)
	return lam
}
