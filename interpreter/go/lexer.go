package main

import (
	"fmt"
	"strings"
	"unicode"
)

func PeekNextToken(lex *Lexer) (token *Token, err error) {
	lexCopy := *lex
	return (&lexCopy).GetNextToken()
}

func NewLexer(input string) *Lexer {
	if len(input) == 0 {
		return nil
	}

	firstRune := rune(input[0])

	return &Lexer{
		text:     input,
		pos:      0,
		currRune: &firstRune,
		Row:      1,
		Col:      1,
	}
}

type Lexer struct {
	text     string
	pos      int
	currRune *rune
	Row      int
	Col      int
}

func (lex *Lexer) GetNextToken() (token *Token, err error) {
	for lex.currRune != nil {
		// spaces
		if unicode.IsSpace(*lex.currRune) {
			lex.skipWhiteSpaces()
			continue
		}
		// comments
		if *lex.currRune == '{' {
			lex.advance()
			if err = lex.skipComment(); err != nil {
				return
			}
			continue
		}
		// numbers
		if unicode.IsDigit(*lex.currRune) {
			return lex.number(), nil
		}
		// id token
		if *lex.currRune == '_' || isAlpha(*lex.currRune) {
			return lex.id(), nil
		}
		// assign token
		nextRune := lex.peek()
		if nextRune != nil && *lex.currRune == ':' && *nextRune == '=' {
			lex.advance()
			lex.advance()
			return lex.staticToken(Assign), nil
		}
		// single-character token
		if kind, ok := TokenValues[string(*lex.currRune)]; ok {
			lex.advance()
			return lex.staticToken(kind), nil
		}
		err = lex.error(ErrorCodeUnknownRune)
		return
	}
	return lex.staticToken(EOF), nil
}

func (lex *Lexer) number() *Token {
	sb := strings.Builder{}
	for lex.currRune != nil && unicode.IsDigit(*lex.currRune) {
		sb.WriteRune(*lex.currRune)
		lex.advance()
	}

	if *lex.currRune == '.' {
		sb.WriteRune(*lex.currRune)
		lex.advance()

		for lex.currRune != nil && unicode.IsDigit(*lex.currRune) {
			sb.WriteRune(*lex.currRune)
			lex.advance()
		}
		return lex.dynamicToken(RealConst, sb.String())
	}
	return lex.dynamicToken(IntegerConst, sb.String())
}

func (lex *Lexer) id() *Token {
	sb := strings.Builder{}
	for lex.currRune != nil &&
		(*lex.currRune == '_' ||
			isAlpha(*lex.currRune) ||
			unicode.IsDigit(*lex.currRune)) {
		sb.WriteRune(*lex.currRune)
		lex.advance()
	}

	idName := normalizeKeyword(sb.String())

	if IsReservedKeyword(idName) {
		return lex.staticToken(TokenValues[idName])
	}
	return lex.dynamicToken(ID, idName)
}

func (lex *Lexer) skipWhiteSpaces() {
	for lex.currRune != nil && unicode.IsSpace(*lex.currRune) {
		lex.advance()
	}
}

// skipComment ignores all chars inside curly braces.
func (lex *Lexer) skipComment() (err error) {
	for *lex.currRune != '}' {
		lex.advance()
		if lex.currRune == nil {
			err = lex.error(ErrorCodeUnclosedComment)
			return
		}
	}
	lex.advance()
	return
}

func (lex *Lexer) advance() {
	if *lex.currRune == '\n' {
		lex.Row += 1
		lex.Col = 0
	}

	lex.pos += 1
	lex.Col += 1
	if lex.pos >= len(lex.text) {
		lex.currRune = nil
		return
	}

	nextRune := rune(lex.text[lex.pos])
	lex.currRune = &nextRune
}

func (lex *Lexer) peek() *rune {
	peekPos := lex.pos + 1
	if peekPos >= len(lex.text) {
		return nil
	}

	nextRune := rune(lex.text[peekPos])
	return &nextRune
}

func (lex *Lexer) error(code ErrorCode) error {
	var lexeme string
	if lex.currRune != nil {
		lexeme = string(*lex.currRune)
	}
	return Error{
		Code:   code,
		Module: ModuleLexer,
		Message: fmt.Sprintf("lexeme=%s,pos=(%d,%d)",
			lexeme, lex.Row, lex.Col),
	}
}

func (lex *Lexer) staticToken(kind TokenKind) *Token {
	return NewStaticToken(kind, lex.Row, lex.Col-len(kind.String()))
}

func (lex *Lexer) dynamicToken(kind TokenKind, value string) *Token {
	return NewDynamicToken(kind, value, lex.Row, lex.Col-len(value))
}

func normalizeKeyword(keyword string) string {
	return strings.ToUpper(keyword)
}

func NewStaticToken(kind TokenKind, row, col int) *Token {
	return &Token{Kind: kind, Value: TokenNames[kind], Row: row, Col: col}
}

func NewDynamicToken(kind TokenKind, value string, row, col int) *Token {
	return &Token{Kind: kind, Value: value, Row: row, Col: col}
}

type Token struct {
	Kind  TokenKind
	Value string
	Row   int
	Col   int
}

func (t Token) String() string {
	return fmt.Sprintf("(kind=%s,value=%s,pos=(%d,%d))",
		t.Kind.String(), t.Value, t.Row, t.Col)
}

type TokenKind int32

func (k TokenKind) String() string {
	if name, ok := TokenNames[k]; ok {
		return name
	}
	return "unknown keyword"
}

const (
	// single-character token kinds.
	Plus     TokenKind = 1
	Minus    TokenKind = 2
	Mul      TokenKind = 3
	FloatDiv TokenKind = 4
	LParen   TokenKind = 5
	RParen   TokenKind = 6
	Semi     TokenKind = 7
	Dot      TokenKind = 8
	Colon    TokenKind = 9
	Comma    TokenKind = 10
	// reserved keywords.
	Program    TokenKind = 1000
	Integer    TokenKind = 1001
	Real       TokenKind = 1002
	IntegerDiv TokenKind = 1003
	Var        TokenKind = 1004
	Procedure  TokenKind = 1005
	Begin      TokenKind = 1006
	End        TokenKind = 1007
	// misc.
	ID           TokenKind = 2001
	IntegerConst TokenKind = 2002
	RealConst    TokenKind = 2003
	Assign       TokenKind = 2004
	EOF          TokenKind = 2005
)

var TokenNames = map[TokenKind]string{
	Plus:         "+",
	Minus:        "-",
	Mul:          "*",
	FloatDiv:     "/",
	LParen:       "(",
	RParen:       ")",
	Semi:         ";",
	Dot:          ".",
	Colon:        ":",
	Comma:        ",",
	Program:      "PROGRAM",
	Integer:      "INTEGER",
	Real:         "REAL",
	IntegerDiv:   "DIV",
	Var:          "VAR",
	Procedure:    "PROCEDURE",
	Begin:        "BEGIN",
	End:          "END",
	ID:           "ID",
	IntegerConst: "INTEGER_CONST",
	RealConst:    "REAL_CONST",
	Assign:       ":=",
	EOF:          "EOF",
}

var TokenValues = map[string]TokenKind{
	"+":             Plus,
	"-":             Minus,
	"*":             Mul,
	"/":             FloatDiv,
	"(":             LParen,
	")":             RParen,
	";":             Semi,
	".":             Dot,
	":":             Colon,
	",":             Comma,
	"PROGRAM":       Program,
	"INTEGER":       Integer,
	"REAL":          Real,
	"DIV":           IntegerDiv,
	"VAR":           Var,
	"PROCEDURE":     Procedure,
	"BEGIN":         Begin,
	"END":           End,
	"ID":            ID,
	"INTEGER_CONST": IntegerConst,
	"REAL_CONST":    RealConst,
	":=":            Assign,
	"EOF":           EOF,
}

func IsReservedKeyword(name string) bool {
	tk := TokenValues[name]
	return tk >= 1000 && tk < 2000
}

func isAlpha(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func init() {
	if len(TokenNames) != len(TokenValues) {
		panic("TokenNames and TokenValues don't match")
	}
}
