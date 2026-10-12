package lexer

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"mihan/internal/diag"
	"mihan/internal/token"
)

const TabWidth = 4

type Lexer struct {
	src  []byte
	file string
	off  int
	line int
	col  int

	atLineStart bool
	indents     []int
	parenDepth  int
	braceDepth  int
	bracePos    []diag.Pos

	comments []token.Token
	bag      *diag.Bag
	toks     []token.Token
	done     bool
}

func New(file string, src []byte, bag *diag.Bag) *Lexer {
	if bag == nil {
		bag = &diag.Bag{}
	}
	return &Lexer{
		src:         src,
		file:        file,
		line:        1,
		col:         1,
		atLineStart: true,
		indents:     []int{0},
		bag:         bag,
	}
}

func (l *Lexer) pos() diag.Pos { return diag.Pos{Off: l.off, Line: l.line, Col: l.col} }

func (l *Lexer) span(p diag.Pos) diag.Span { return diag.Span{Start: p, End: l.pos()} }

func (l *Lexer) errf(code string, p diag.Pos, fa, enFormat string, args ...any) {
	l.bag.Add(diag.Errorf(code, l.span(p), fa, enFormat, args...).WithFile(l.file))
}

func (l *Lexer) eof() bool { return l.off >= len(l.src) }

func (l *Lexer) peek() rune {
	if l.off >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRune(l.src[l.off:])
	return r
}

func (l *Lexer) peekAt(n int) rune {
	p := l.off
	for i := 0; i < n; i++ {
		if p >= len(l.src) {
			return 0
		}
		_, w := utf8.DecodeRune(l.src[p:])
		p += w
	}
	if p >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRune(l.src[p:])
	return r
}

func (l *Lexer) next() rune {
	if l.eof() {
		return 0
	}
	r, w := utf8.DecodeRune(l.src[l.off:])
	l.off += w
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func (l *Lexer) suspended() bool { return l.parenDepth > 0 || l.braceDepth > 0 }

func (l *Lexer) Tokens() []token.Token {
	for !l.done && !l.eof() {
		l.scanLine()
	}

	if n := len(l.toks); n > 0 && l.toks[n-1].Kind != token.NEWLINE && l.toks[n-1].Kind != token.DEDENT {
		l.push(token.NEWLINE, "\n", "", l.pos())
	}
	for len(l.indents) > 1 {
		l.indents = l.indents[:len(l.indents)-1]
		l.push(token.DEDENT, "", "", l.pos())
	}
	l.push(token.EOF, "", "", l.pos())
	return l.toks
}

func (l *Lexer) Comments() []token.Token { return l.comments }

func (l *Lexer) push(k token.Kind, lit, text string, p diag.Pos) {
	l.toks = append(l.toks, token.Token{
		Kind: k, Lit: lit, Text: text,
		Line: p.Line, Col: p.Col, Pos: p.Off, Off: p.Off, End: l.off,
	})
}

func (l *Lexer) scanLine() {
	if l.suspended() {
		l.scanToken()
		return
	}
	if l.atLineStart {
		if !l.measureIndent() {
			return
		}
	}
	l.atLineStart = false
	for !l.eof() && (l.peek() == ' ' || l.peek() == '\t') {
		l.next()
	}
	if l.eof() {
		return
	}
	l.scanToken()
}

func (l *Lexer) measureIndent() bool {
	width := 0
	p := l.pos()
	for !l.eof() {
		switch l.peek() {
		case ' ':
			width++
			l.next()
		case '\t':
			width = (width/TabWidth + 1) * TabWidth
			l.next()
		default:
			goto measured
		}
	}
measured:
	if l.eof() {
		l.done = true
		return false
	}
	if r := l.peek(); r == '\n' || r == '\r' {

		l.next()
		l.atLineStart = true
		return false
	}

	if l.peek() == '}' {
		return true
	}
	top := l.indents[len(l.indents)-1]
	switch {
	case width > top:
		l.indents = append(l.indents, width)
		l.push(token.INDENT, "", "", p)
	case width < top:
		for len(l.indents) > 1 && width < l.indents[len(l.indents)-1] {
			l.indents = l.indents[:len(l.indents)-1]
			l.push(token.DEDENT, "", "", p)
		}
		if got := l.indents[len(l.indents)-1]; got != width {
			l.errf("E1002", p,
				"تورفتگی %d با تورفتگی %d جور در نمی‌آید", "indentation of %d does not line up with %d",
				width, got)
		}
	}
	return true
}

func (l *Lexer) endLine() {
	if l.parenDepth == 0 && !l.eof() {
		l.push(token.NEWLINE, "\n", "", l.pos())
	}
	l.atLineStart = true
}

func (l *Lexer) scanToken() {

	for !l.eof() && (l.peek() == ' ' || l.peek() == '\t') {
		l.next()
	}
	if l.eof() {
		l.done = true
		return
	}
	r := l.peek()
	switch {
	case r == '\n' || r == '\r':
		l.next()
		l.endLine()
	case r == '#':
		l.lexLineComment("#")
	case r == '/' && l.peekAt(1) == '/':
		l.lexLineComment("//")
	case r == '/' && l.peekAt(1) == '*':
		l.lexBlockComment()
	case r == '"' || r == '`':
		l.lexString(false)
	case r == '$' && (l.peekAt(1) == '"' || l.peekAt(1) == '`'):
		l.lexString(true)
	case r == '\'':
		l.lexChar()
	case isDigit(r):
		l.lexNumber()
	case isIdentStart(r):
		l.lexWord()
	case l.lexOperator():
	default:
		p := l.pos()
		bad := l.next()
		l.errf("E1001", p, "نویسه %q شناخته نشد", "unrecognised character %q", string(bad))
	}
}

func (l *Lexer) lexLineComment(marker string) {
	p := l.pos()
	start := l.off
	for !l.eof() && l.peek() != '\n' {
		l.next()
	}
	lit := string(l.src[start:l.off])
	if marker == "//" {
		lit = marker + lit
	}
	tk := l.make(token.COMMENT, lit, "", p)
	l.comments = append(l.comments, tk)
}

func (l *Lexer) lexBlockComment() {
	p := l.pos()
	start := l.off
	l.next()
	l.next()
	for !l.eof() {
		if l.peek() == '*' && l.peekAt(1) == '/' {
			l.next()
			l.next()
			break
		}
		l.next()
	}
	l.comments = append(l.comments, l.make(token.COMMENT, string(l.src[start:l.off]), "", p))
}

func (l *Lexer) make(k token.Kind, lit, text string, p diag.Pos) token.Token {
	return token.Token{Kind: k, Lit: lit, Text: text, Line: p.Line, Col: p.Col, Pos: p.Off, Off: p.Off, End: l.off}
}

func (l *Lexer) lexWord() {
	p := l.pos()
	start := l.off
	for !l.eof() && isIdentPart(l.peek()) {
		l.next()
	}
	word := string(l.src[start:l.off])
	if kw := token.LookupKeyword(word); kw != nil {

		l.push(token.Kind(kw.Name), word, "", p)
		return
	}
	l.push(token.IDENT, word, word, p)
}

func (l *Lexer) lexNumber() {
	p := l.pos()
	start := l.off
	isFloat := false

	if l.peek() == '0' && !isDigit(l.peekAt(1)) && l.peekAt(1) != '.' {
		base := 0
		switch lower(l.peekAt(1)) {
		case 'x':
			base = 16
		case 'o':
			base = 8
		case 'b':
			base = 2
		}
		if base > 0 {
			l.next()
			l.next()
			for !l.eof() && (isBaseDigit(l.peek(), base) || l.peek() == '_') {
				l.next()
			}
			l.finishNumber(p, start, token.INT)
			return
		}
	}

	for !l.eof() && (isDigit(l.peek()) || l.peek() == '_') {
		l.next()
	}
	if l.peek() == '.' && isDigit(l.peekAt(1)) {
		isFloat = true
		l.next()
		for !l.eof() && (isDigit(l.peek()) || l.peek() == '_') {
			l.next()
		}
	}
	if r := lower(l.peek()); r == 'e' {
		n := 1
		if s := l.peekAt(1); s == '+' || s == '-' {
			n = 2
		}
		if isDigit(l.peekAt(n)) {
			isFloat = true
			for i := 0; i < n+1; i++ {
				l.next()
			}
			for !l.eof() && (isDigit(l.peek()) || l.peek() == '_') {
				l.next()
			}
		}
	}
	if !l.eof() && isIdentStart(l.peek()) {
		l.errf("E1003", p,
			"پس از عدد %s فاصله لازم است", "number %s needs a space before the name that follows it",
			string(l.src[start:l.off]))
	}
	if isFloat {
		l.finishNumber(p, start, token.FLOAT)
	} else {
		l.finishNumber(p, start, token.INT)
	}
}

func (l *Lexer) finishNumber(p diag.Pos, start int, k token.Kind) {
	raw := string(l.src[start:l.off])
	clean := strings.ReplaceAll(raw, "_", "")
	if k == token.INT {
		if v, err := strconv.ParseUint(clean, 0, 64); err == nil {
			l.push(token.INT, raw, strconv.FormatUint(v, 10), p)
			return
		}
	}
	if v, err := strconv.ParseFloat(clean, 64); err == nil {
		l.push(token.FLOAT, raw, strconv.FormatFloat(v, 'g', -1, 64), p)
		return
	}
	l.errf("E1007", p, "عدد %s خوانده نشد", "%s is not a number I can read", raw)
	l.push(token.INT, raw, "0", p)
}

func (l *Lexer) lexChar() {
	p := l.pos()
	l.next()
	var b strings.Builder
	for {
		if l.eof() || l.peek() == '\n' {
			l.errf("E1004", p, "کاراکتر بسته نشد", "this character literal is never closed")
			break
		}
		if l.peek() == '\'' {
			l.next()
			break
		}
		if l.peek() == '\\' {
			b.WriteRune(l.lexEscape())
			continue
		}
		b.WriteRune(l.next())
	}
	if rs := []rune(b.String()); len(rs) != 1 {
		l.errf("E1004", p,
			"یک کاراکتر باید داخل ' بنشیند", "a character literal holds exactly one character, not %d",
			len(rs))
	}
	l.push(token.CHAR, string(l.src[p.Off:l.off]), b.String(), p)
}

func (l *Lexer) lexString(interp bool) {
	p := l.pos()
	if interp {
		l.next()
	}
	quote := l.peek()
	l.next()
	raw := quote == '`'

	startIdx := len(l.toks)
	l.push(token.INTERP, string(quote), "", p)

	var cur strings.Builder
	curPos := l.pos()
	holes := 0
	flush := func() {
		if cur.Len() > 0 {
			l.push(token.INTERP_TEXT, cur.String(), cur.String(), curPos)
			cur.Reset()
		}
	}

	for {
		if l.eof() || (!raw && l.peek() == '\n') {
			l.errf("E1005", p, "رشته بسته نشد", "this string is never closed")
			break
		}
		r := l.peek()
		if r == quote {
			l.next()
			break
		}
		if !raw && r == '\\' {
			cur.WriteRune(l.lexEscape())
			continue
		}
		if !raw && r == '{' && interp {

			if l.peekAt(1) == '{' {
				l.next()
				cur.WriteRune(l.next())
				continue
			}

			flush()
			l.next()
			holes++
			curPos = l.pos()
			l.lexHoleTokens()
			continue
		}
		if !raw && r == '}' && interp {

			if l.peekAt(1) == '}' {
				l.next()
				cur.WriteRune(l.next())
				continue
			}
		}
		if !raw && r == '$' {
			switch {
			case l.peekAt(1) == '{':
				flush()
				l.next()
				l.next()
				holes++
				curPos = l.pos()
				l.lexHoleTokens()
			case isIdentStart(l.peekAt(1)):
				flush()
				l.next()
				holes++
				curPos = l.pos()
				l.lexIdentTokens()
			default:

				cur.WriteRune(l.next())
			}
			continue
		}
		cur.WriteRune(l.next())
	}
	flush()

	lit := string(l.src[p.Off:l.off])
	if holes == 0 {
		var text string
		for _, tk := range l.toks[startIdx:] {
			text += tk.Text
		}
		l.toks = l.toks[:startIdx]
		l.push(token.STRING, lit, text, p)
		return
	}
	l.push(token.STRING_END, string(quote), "", l.pos())
}

func (l *Lexer) lexHoleTokens() {
	depth := 0
	for !l.eof() {

		for !l.eof() && (l.peek() == ' ' || l.peek() == '\t' || l.peek() == '\n' || l.peek() == '\r') {
			l.next()
		}
		if l.eof() {
			break
		}
		r := l.peek()
		if r == '}' && depth == 0 {
			q := l.pos()
			l.next()
			l.push(token.INTERP_END, "}", "", q)
			return
		}
		switch r {
		case '{', '}':

			q := l.pos()
			l.next()
			if r == '{' {
				depth++
			} else {
				depth--
			}
			l.push(token.Kind(string(r)), string(r), string(r), q)
		default:
			if !l.lexOneTokenInHole() {
				l.next()
			}
		}
	}
	l.errf("E1008", l.pos(), "کران ${ بسته نشد", "this ${ was never closed")
}

func (l *Lexer) lexIdentTokens() {
	if !l.lexOneTokenInHole() {
		l.next()
	}
	p := l.pos()
	l.push(token.INTERP_END, "}", "", p)
}

func (l *Lexer) lexOneTokenInHole() bool {
	if l.eof() {
		return false
	}
	savedParen, savedBrace := l.parenDepth, l.braceDepth
	savedLineStart, savedBraces := l.atLineStart, l.bracePos
	l.parenDepth, l.braceDepth, l.atLineStart, l.bracePos = 1, 0, false, nil
	before := len(l.toks)
	l.scanToken()
	l.parenDepth, l.braceDepth, l.atLineStart, l.bracePos = savedParen, savedBrace, savedLineStart, savedBraces
	return len(l.toks) > before
}

func (l *Lexer) lexEscape() rune {
	l.next()
	if l.eof() {
		return '\\'
	}
	r := l.next()
	switch r {
	case 'n':
		return '\n'
	case 't':
		return '\t'
	case 'r':
		return '\r'
	case '0':
		return 0
	case 'a':
		return 7
	case 'b':
		return 8
	case 'f':
		return 12
	case 'v':
		return 11
	case '\\', '\'', '"', '$', '{', '}':
		return r
	case 'u', 'U':
		n := 4
		if r == 'U' {
			n = 8
		}
		if l.off+n > len(l.src) {
			return 'u'
		}
		v, ok := parseHex(l.src[l.off : l.off+n])
		if !ok {
			return 'u'
		}
		for i := 0; i < n; i++ {
			l.next()
		}
		return rune(v)
	case 'x':
		if l.off+2 > len(l.src) {
			return 'x'
		}
		v, ok := parseHex(l.src[l.off : l.off+2])
		if !ok {
			return 'x'
		}
		l.next()
		l.next()
		return rune(v)
	}
	return r
}

func (l *Lexer) lexOperator() bool {
	maxLen := token.MaxSymbolLen()
	if l.off+maxLen > len(l.src) {
		maxLen = len(l.src) - l.off
	}
	for n := maxLen; n > 0; n-- {
		sym := string(l.src[l.off : l.off+n])
		if token.LookupOperator(sym) == nil {
			continue
		}
		p := l.pos()
		for i := 0; i < n; i++ {
			l.next()
		}
		l.push(token.Kind(sym), sym, sym, p)
		switch sym {
		case "(", "[":
			l.parenDepth++
		case ")", "]":
			if l.parenDepth > 0 {
				l.parenDepth--
			} else {
				l.errf("E1006", p, "برای %q جفت پیدا نشد", "%q has no opening bracket here", sym)
			}
		case "{":
			l.braceDepth++
			l.bracePos = append(l.bracePos, p)
		case "}":
			if l.braceDepth > 0 {
				l.braceDepth--
				l.bracePos = l.bracePos[:len(l.bracePos)-1]
			} else {
				l.errf("E1006", p, "برای %q جفت پیدا نشد", "%q has no opening bracket here", sym)
			}
		}
		return true
	}
	return false
}

func lower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

func isBaseDigit(r rune, base int) bool {
	if r == '_' {
		return false
	}
	v := 0
	switch {
	case r >= '0' && r <= '9':
		v = int(r - '0')
	case r >= 'a' && r <= 'f':
		v = int(r-'a') + 10
	case r >= 'A' && r <= 'F':
		v = int(r-'A') + 10
	default:
		return false
	}
	return v < base
}

func parseHex(b []byte) (int, bool) {
	v := 0
	for _, c := range b {
		d := -1
		switch {
		case c >= '0' && c <= '9':
			d = int(c - '0')
		case c >= 'a' && c <= 'f':
			d = int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			d = int(c-'A') + 10
		}
		if d < 0 {
			return 0, false
		}
		v = v*16 + d
	}
	return v, true
}

func isIdentStart(r rune) bool { return r == '_' || unicode.IsLetter(r) }

func isIdentPart(r rune) bool { return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) }
