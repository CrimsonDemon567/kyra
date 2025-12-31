package lexer

import "unicode"

// TokenType describes the kind of token.
type TokenType string

const (
	// Special
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"
	NEWLINE TokenType = "NEWLINE"
	INDENT  TokenType = "INDENT"
	DEDENT  TokenType = "DEDENT"

	// Identifiers + literals
	IDENT  TokenType = "IDENT"
	NUMBER TokenType = "NUMBER"
	STRING TokenType = "STRING"

	// Keywords
	K_DEF    TokenType = "DEF"
	K_FUNC   TokenType = "FUNC"
	K_USE    TokenType = "USE"
	K_LET    TokenType = "LET"
	K_IF     TokenType = "IF"
	K_ELSE   TokenType = "ELSE"
	K_WHILE  TokenType = "WHILE"
	K_FOR    TokenType = "FOR"
	K_RETURN TokenType = "RETURN"
	K_EXIT   TokenType = "EXIT"
	K_PASS   TokenType = "PASS"
	K_TRUE   TokenType = "TRUE"
	K_FALSE  TokenType = "FALSE"

	// Types (not strictly required in the lexer, but convenient)
	K_I32    TokenType = "I32"
	K_I64    TokenType = "I64"
	K_F32    TokenType = "F32"
	K_F64    TokenType = "F64"
	K_BOOL   TokenType = "BOOL"
	K_STRING TokenType = "STRING_TYPE"
	K_VOID   TokenType = "VOID"

	// Operators
	ASSIGN   TokenType = "="
	PLUS     TokenType = "+"
	MINUS    TokenType = "-"
	STAR     TokenType = "*"
	SLASH    TokenType = "/"
	PERCENT  TokenType = "%"
	BANG     TokenType = "!"
	LT       TokenType = "<"
	GT       TokenType = ">"
	LE       TokenType = "<="
	GE       TokenType = ">="
	EQ       TokenType = "=="
	NEQ      TokenType = "!="
	AND      TokenType = "&&"
	OR       TokenType = "||"
	PLUS_EQ  TokenType = "+="
	MINUS_EQ TokenType = "-="
	MUL_EQ   TokenType = "*="
	DIV_EQ   TokenType = "/="

	// Delimiters
	COMMA     TokenType = ","
	COLON     TokenType = ":"
	DOT       TokenType = "."
	LPAREN    TokenType = "("
	RPAREN    TokenType = ")"
	LBRACE    TokenType = "{"
	RBRACE    TokenType = "}"
	LBRACKET  TokenType = "["
	RBRACKET  TokenType = "]"
	ARROW     TokenType = "->"
)

// Token represents a single lexical token.
type Token struct {
	Type    TokenType
	Lexeme  string
	Line    int
	Column  int
}

// Lexer converts source text into tokens.
type Lexer struct {
	src         []rune
	pos         int
	line        int
	col         int
	indentStack []int
	startOfLine bool
}

// New creates a new lexer for the given source.
func New(src string) *Lexer {
	return &Lexer{
		src:         []rune(src),
		line:        1,
		col:         0,
		indentStack: []int{0},
		startOfLine: true,
	}
}

// Lex tokenizes the entire input and returns a token slice.
func (l *Lexer) Lex() []Token {
	var tokens []Token

	for {
		tok := l.nextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}

	return tokens
}

func (l *Lexer) nextToken() Token {
	// Handle indentation only at the start of a line
	if l.startOfLine {
		l.startOfLine = false
		return l.lexIndentation()
	}

	l.skipWhitespaceExceptNewline()

	if l.isAtEnd() {
		// Emit DEDENTs for any remaining indentation
		if len(l.indentStack) > 1 {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			return Token{Type: DEDENT, Lexeme: "", Line: l.line, Column: l.col}
		}
		return Token{Type: EOF, Lexeme: "", Line: l.line, Column: l.col}
	}

	ch := l.peek()

	// Newline
	if ch == '\n' {
		l.advance()
		l.line++
		l.col = 0
		l.startOfLine = true
		return Token{Type: NEWLINE, Lexeme: "\n", Line: l.line - 1, Column: 0}
	}

	// Comments
	if ch == '#' {
		l.skipLineComment()
		return l.nextToken()
	}
	if ch == '/' && l.peekNext() == '*' {
		l.skipBlockComment()
		return l.nextToken()
	}

	// Identifiers / keywords
	if isLetter(ch) || ch == '_' {
		return l.lexIdentifierOrKeyword()
	}

	// Numbers
	if isDigit(ch) {
		return l.lexNumber()
	}

	// Strings: "..." , '...' , """..."""
	if ch == '"' {
		// Check for multiline """
		if l.peekNext() == '"' && l.peekThird() == '"' {
			return l.lexTripleString()
		}
		return l.lexString('"')
	}
	if ch == '\'' {
		return l.lexString('\'')
	}

	// Operators and delimiters
	return l.lexSymbol()
}

func (l *Lexer) isAtEnd() bool {
	return l.pos >= len(l.src)
}

func (l *Lexer) advance() rune {
	if l.isAtEnd() {
		return 0
	}
	ch := l.src[l.pos]
	l.pos++
	l.col++
	return ch
}

func (l *Lexer) peek() rune {
	if l.isAtEnd() {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) peekNext() rune {
	if l.pos+1 >= len(l.src) {
		return 0
	}
	return l.src[l.pos+1]
}

func (l *Lexer) peekThird() rune {
	if l.pos+2 >= len(l.src) {
		return 0
	}
	return l.src[l.pos+2]
}

func (l *Lexer) skipWhitespaceExceptNewline() {
	for !l.isAtEnd() {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) skipLineComment() {
	for !l.isAtEnd() && l.peek() != '\n' {
		l.advance()
	}
}

func (l *Lexer) skipBlockComment() {
	// Skip '/*'
	l.advance()
	l.advance()
	for !l.isAtEnd() {
		if l.peek() == '*' && l.peekNext() == '/' {
			l.advance()
			l.advance()
			break
		}
		ch := l.advance()
		if ch == '\n' {
			l.line++
			l.col = 0
		}
	}
}

func (l *Lexer) lexIndentation() Token {
	// Count spaces at the start of the line
	count := 0
	for !l.isAtEnd() {
		ch := l.peek()
		if ch == ' ' {
			count++
			l.advance()
		} else if ch == '\t' {
			// Tabs are treated as 4 spaces here
			count += 4
			l.advance()
		} else {
			break
		}
	}

	// Blank line or comment line
	if l.isAtEnd() || l.peek() == '\n' || l.peek() == '#' ||
		(l.peek() == '/' && l.peekNext() == '*') {
		l.startOfLine = false
		return l.nextToken()
	}

	currentIndent := l.indentStack[len(l.indentStack)-1]
	if count > currentIndent {
		l.indentStack = append(l.indentStack, count)
		return Token{Type: INDENT, Lexeme: "", Line: l.line, Column: 0}
	}
	if count < currentIndent {
		// Pop until we match or underflow
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
		return Token{Type: DEDENT, Lexeme: "", Line: l.line, Column: 0}
	}

	// Same indentation: just continue
	return l.nextToken()
}

func (l *Lexer) lexIdentifierOrKeyword() Token {
	startCol := l.col
	startPos := l.pos
	for !l.isAtEnd() && (isLetter(l.peek()) || isDigit(l.peek()) || l.peek() == '_') {
		l.advance()
	}
	lex := string(l.src[startPos:l.pos])

	switch lex {
	case "def":
		return Token{Type: K_DEF, Lexeme: lex, Line: l.line, Column: startCol}
	case "func":
		return Token{Type: K_FUNC, Lexeme: lex, Line: l.line, Column: startCol}
	case "use":
		return Token{Type: K_USE, Lexeme: lex, Line: l.line, Column: startCol}
	case "let":
		return Token{Type: K_LET, Lexeme: lex, Line: l.line, Column: startCol}
	case "if":
		return Token{Type: K_IF, Lexeme: lex, Line: l.line, Column: startCol}
	case "else":
		return Token{Type: K_ELSE, Lexeme: lex, Line: l.line, Column: startCol}
	case "while":
		return Token{Type: K_WHILE, Lexeme: lex, Line: l.line, Column: startCol}
	case "for":
		return Token{Type: K_FOR, Lexeme: lex, Line: l.line, Column: startCol}
	case "return":
		return Token{Type: K_RETURN, Lexeme: lex, Line: l.line, Column: startCol}
	case "exit":
		return Token{Type: K_EXIT, Lexeme: lex, Line: l.line, Column: startCol}
	case "pass":
		return Token{Type: K_PASS, Lexeme: lex, Line: l.line, Column: startCol}
	case "true":
		return Token{Type: K_TRUE, Lexeme: lex, Line: l.line, Column: startCol}
	case "false":
		return Token{Type: K_FALSE, Lexeme: lex, Line: l.line, Column: startCol}
	case "i32":
		return Token{Type: K_I32, Lexeme: lex, Line: l.line, Column: startCol}
	case "i64":
		return Token{Type: K_I64, Lexeme: lex, Line: l.line, Column: startCol}
	case "f32":
		return Token{Type: K_F32, Lexeme: lex, Line: l.line, Column: startCol}
	case "f64":
		return Token{Type: K_F64, Lexeme: lex, Line: l.line, Column: startCol}
	case "bool":
		return Token{Type: K_BOOL, Lexeme: lex, Line: l.line, Column: startCol}
	case "string":
		return Token{Type: K_STRING, Lexeme: lex, Line: l.line, Column: startCol}
	case "void":
		return Token{Type: K_VOID, Lexeme: lex, Line: l.line, Column: startCol}
	default:
		return Token{Type: IDENT, Lexeme: lex, Line: l.line, Column: startCol}
	}
}

func (l *Lexer) lexNumber() Token {
	startCol := l.col
	startPos := l.pos
	hasDot := false

	for !l.isAtEnd() {
		ch := l.peek()
		if isDigit(ch) {
			l.advance()
		} else if ch == '.' && !hasDot {
			hasDot = true
			l.advance()
		} else {
			break
		}
	}

	lex := string(l.src[startPos:l.pos])
	return Token{Type: NUMBER, Lexeme: lex, Line: l.line, Column: startCol}
}

func (l *Lexer) lexString(quote rune) Token {
	startCol := l.col
	l.advance() // consume opening quote
	startPos := l.pos

	for !l.isAtEnd() && l.peek() != quote {
		ch := l.advance()
		if ch == '\n' {
			l.line++
			l.col = 0
		}
	}
	lex := string(l.src[startPos:l.pos])

	if !l.isAtEnd() {
		l.advance() // closing quote
	}

	return Token{Type: STRING, Lexeme: lex, Line: l.line, Column: startCol}
}

func (l *Lexer) lexTripleString() Token {
	startCol := l.col
	// consume """
	l.advance()
	l.advance()
	l.advance()
	startPos := l.pos

	for !l.isAtEnd() {
		if l.peek() == '"' && l.peekNext() == '"' && l.peekThird() == '"' {
			break
		}
		ch := l.advance()
		if ch == '\n' {
			l.line++
			l.col = 0
		}
	}
	lex := string(l.src[startPos:l.pos])

	if !l.isAtEnd() {
		// consume closing """
		l.advance()
		l.advance()
		l.advance()
	}

	return Token{Type: STRING, Lexeme: lex, Line: l.line, Column: startCol}
}

func (l *Lexer) lexSymbol() Token {
	ch := l.advance()
	startCol := l.col - 1

	switch ch {
	case ',':
		return Token{Type: COMMA, Lexeme: ",", Line: l.line, Column: startCol}
	case ':':
		return Token{Type: COLON, Lexeme: ":", Line: l.line, Column: startCol}
	case '.':
		return Token{Type: DOT, Lexeme: ".", Line: l.line, Column: startCol}
	case '(':
		return Token{Type: LPAREN, Lexeme: "(", Line: l.line, Column: startCol}
	case ')':
		return Token{Type: RPAREN, Lexeme: ")", Line: l.line, Column: startCol}
	case '{':
		return Token{Type: LBRACE, Lexeme: "{", Line: l.line, Column: startCol}
	case '}':
		return Token{Type: RBRACE, Lexeme: "}", Line: l.line, Column: startCol}
	case '[':
		return Token{Type: LBRACKET, Lexeme: "[", Line: l.line, Column: startCol}
	case ']':
		return Token{Type: RBRACKET, Lexeme: "]", Line: l.line, Column: startCol}
	case '+':
		if l.peek() == '=' {
			l.advance()
			return Token{Type: PLUS_EQ, Lexeme: "+=", Line: l.line, Column: startCol}
		}
		return Token{Type: PLUS, Lexeme: "+", Line: l.line, Column: startCol}
	case '-':
		if l.peek() == '>' {
			l.advance()
			return Token{Type: ARROW, Lexeme: "->", Line: l.line, Column: startCol}
		}
		if l.peek() == '=' {
			l.advance()
			return Token{Type: MINUS_EQ, Lexeme: "-=", Line: l.line, Column: startCol}
		}
		return Token{Type: MINUS, Lexeme: "-", Line: l.line, Column: startCol}
	case '*':
		if l.peek() == '=' {
			l.advance()
			return Token{Type: MUL_EQ, Lexeme: "*=", Line: l.line, Column: startCol}
		}
		return Token{Type: STAR, Lexeme: "*", Line: l.line, Column: startCol}
	case '/':
		if l.peek() == '=' {
			l.advance()
			return Token{Type: DIV_EQ, Lexeme: "/=", Line: l.line, Column: startCol}
		}
		return Token{Type: SLASH, Lexeme: "/", Line: l.line, Column: startCol}
	case '%':
		return Token{Type: PERCENT, Lexeme: "%", Line: l.line, Column: startCol}
	case '=':
		if l.peek() == '=' {
			l.advance()
			return Token{Type: EQ, Lexeme: "==", Line: l.line, Column: startCol}
		}
		return Token{Type: ASSIGN, Lexeme: "=", Line: l.line, Column: startCol}
	case '!':
		if l.peek() == '=' {
			l.advance()
			return Token{Type: NEQ, Lexeme: "!=", Line: l.line, Column: startCol}
		}
		return Token{Type: BANG, Lexeme: "!", Line: l.line, Column: startCol}
	case '<':
		if l.peek() == '=' {
			l.advance()
			return Token{Type: LE, Lexeme: "<=", Line: l.line, Column: startCol}
		}
		return Token{Type: LT, Lexeme: "<", Line: l.line, Column: startCol}
	case '>':
		if l.peek() == '=' {
			l.advance()
			return Token{Type: GE, Lexeme: ">=", Line: l.line, Column: startCol}
		}
		return Token{Type: GT, Lexeme: ">", Line: l.line, Column: startCol}
	case '&':
		if l.peek() == '&' {
			l.advance()
			return Token{Type: AND, Lexeme: "&&", Line: l.line, Column: startCol}
		}
	case '|':
		if l.peek() == '|' {
			l.advance()
			return Token{Type: OR, Lexeme: "||", Line: l.line, Column: startCol}
		}
	}

	return Token{Type: ILLEGAL, Lexeme: string(ch), Line: l.line, Column: startCol}
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}
