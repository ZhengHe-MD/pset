package main

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer_GetNextToken(t *testing.T) {
	tests := map[string]struct {
		givenText  string
		wantTokens []*Token
	}{
		"simple": {
			givenText: "BEGIN a := 2; END.",
			wantTokens: []*Token{
				NewStaticToken(Begin, 1, 1),
				NewDynamicToken(ID, "A", 1, 7),
				NewStaticToken(Assign, 1, 9),
				NewDynamicToken(IntegerConst, "2", 1, 12),
				NewStaticToken(Semi, 1, 13),
				NewStaticToken(End, 1, 15),
				NewStaticToken(Dot, 1, 18),
			},
		},
		"variable with underscores": {
			givenText: "BEGIN _num := 2; END.",
			wantTokens: []*Token{
				NewStaticToken(Begin, 1, 1),
				NewDynamicToken(ID, "_NUM", 1, 7),
				NewStaticToken(Assign, 1, 12),
				NewDynamicToken(IntegerConst, "2", 1, 15),
				NewStaticToken(Semi, 1, 16),
				NewStaticToken(End, 1, 18),
				NewStaticToken(Dot, 1, 21),
			},
		},
		"complex": {
			givenText: `
				BEGIN
					BEGIN
						number := 2;
						a := number;
						b := 10 * a + 10 * number DIV 4;
						c := a - - b
					END;
				x := 11;
				END.
			`,
			wantTokens: []*Token{
				NewStaticToken(Begin, 2, 5),
				NewStaticToken(Begin, 3, 6),
				// number := 2;
				NewDynamicToken(ID, "NUMBER", 4, 7), NewStaticToken(Assign, 4, 14),
				NewDynamicToken(IntegerConst, "2", 4, 17), NewStaticToken(Semi, 4, 18),
				// a := number
				NewDynamicToken(ID, "A", 5, 7), NewStaticToken(Assign, 5, 9),
				NewDynamicToken(ID, "NUMBER", 5, 12), NewStaticToken(Semi, 5, 18),
				// b := 10 * a + 10 * number DIV 4;
				NewDynamicToken(ID, "B", 6, 7), NewStaticToken(Assign, 6, 9),
				NewDynamicToken(IntegerConst, "10", 6, 12), NewStaticToken(Mul, 6, 15),
				NewDynamicToken(ID, "A", 6, 17), NewStaticToken(Plus, 6, 19),
				NewDynamicToken(IntegerConst, "10", 6, 21), NewStaticToken(Mul, 6, 24),
				NewDynamicToken(ID, "NUMBER", 6, 26), NewStaticToken(IntegerDiv, 6, 33),
				NewDynamicToken(IntegerConst, "4", 6, 37), NewStaticToken(Semi, 6, 38),
				// c := a - - b
				NewDynamicToken(ID, "C", 7, 7), NewStaticToken(Assign, 7, 9),
				NewDynamicToken(ID, "A", 7, 12), NewStaticToken(Minus, 7, 14),
				NewStaticToken(Minus, 7, 16), NewDynamicToken(ID, "B", 7, 18),
				// END;
				NewStaticToken(End, 8, 6), NewStaticToken(Semi, 8, 9),
				// x := 11;
				NewDynamicToken(ID, "X", 9, 5), NewStaticToken(Assign, 9, 7),
				NewDynamicToken(IntegerConst, "11", 9, 10), NewStaticToken(Semi, 9, 12),
				// END.
				NewStaticToken(End, 10, 5), NewStaticToken(Dot, 10, 8),
			},
		},
		"var declarations": {
			givenText: `
				VAR
					number     : INTEGER;
					a, b, c, x : INTEGER;
					y          : REAL;
			`,
			wantTokens: []*Token{
				NewStaticToken(Var, 2, 5),
				NewDynamicToken(ID, "NUMBER", 3, 6), NewStaticToken(Colon, 3, 17), NewStaticToken(Integer, 3, 19), NewStaticToken(Semi, 3, 26),
				NewDynamicToken(ID, "A", 4, 6), NewStaticToken(Comma, 4, 7), NewDynamicToken(ID, "B", 4, 9), NewStaticToken(Comma, 4, 10), NewDynamicToken(ID, "C", 4, 12), NewStaticToken(Comma, 4, 13), NewDynamicToken(ID, "X", 4, 15), NewStaticToken(Colon, 4, 17), NewStaticToken(Integer, 4, 19), NewStaticToken(Semi, 4, 26),
				NewDynamicToken(ID, "Y", 5, 6), NewStaticToken(Colon, 5, 17), NewStaticToken(Real, 5, 19), NewStaticToken(Semi, 5, 23),
			},
		},
		"multiple var declarations": {
			givenText: `
				VAR	 number     : INTEGER;
				VAR	 a, b, c, x : INTEGER;
				VAR	 y          : REAL;
			`,
			wantTokens: []*Token{
				NewStaticToken(Var, 2, 5), NewDynamicToken(ID, "NUMBER", 2, 10), NewStaticToken(Colon, 2, 21), NewStaticToken(Integer, 2, 23), NewStaticToken(Semi, 2, 30),
				NewStaticToken(Var, 3, 5), NewDynamicToken(ID, "A", 3, 10), NewStaticToken(Comma, 3, 11), NewDynamicToken(ID, "B", 3, 13), NewStaticToken(Comma, 3, 14), NewDynamicToken(ID, "C", 3, 16), NewStaticToken(Comma, 3, 17), NewDynamicToken(ID, "X", 3, 19), NewStaticToken(Colon, 3, 21), NewStaticToken(Integer, 3, 23), NewStaticToken(Semi, 3, 30),
				NewStaticToken(Var, 4, 5), NewDynamicToken(ID, "Y", 4, 10), NewStaticToken(Colon, 4, 21), NewStaticToken(Real, 4, 23), NewStaticToken(Semi, 4, 27),
			},
		},
		"program": {
			givenText: `
				PROGRAM Part10;
				BEGIN
				END.
			`,
			wantTokens: []*Token{
				NewStaticToken(Program, 2, 5), NewDynamicToken(ID, "PART10", 2, 13), NewStaticToken(Semi, 2, 19),
				NewStaticToken(Begin, 3, 5),
				NewStaticToken(End, 4, 5), NewStaticToken(Dot, 4, 8),
			},
		},
		"ignore comments": {
			givenText: `
				BEGIN
				{ writeln('a = ', a); }
				{ writeln('b = ', b); }
				{ writeln('c = ', c); }
				{ writeln('number = ', number); }
				{ writeln('x = ', x); }
				{ writeln('y = ', y); }
				END;
			`,
			wantTokens: []*Token{
				NewStaticToken(Begin, 2, 5),
				NewStaticToken(End, 9, 5), NewStaticToken(Semi, 9, 8),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer(tc.givenText)
			for _, wantToken := range tc.wantTokens {
				givenToken, err := l.GetNextToken()
				assert.NoError(t, err)
				assert.Equal(t, wantToken, givenToken)
			}
			lastToken, err := l.GetNextToken()
			assert.NoError(t, err)
			assert.Equal(t, EOF, lastToken.Kind)
		})
	}

	t.Run("unknown rune", func(t *testing.T) {
		l := NewLexer("✅")
		token, err := l.GetNextToken()
		assert.Nil(t, token)
		assert.Equal(t, `<Error: module=Lexer,code=UnknownRune,message="lexeme=â,pos=(1,1)">`, err.Error())
	})

	t.Run("unclosed comment", func(t *testing.T) {
		l := NewLexer("{haha")
		token, err := l.GetNextToken()
		log.Println(token)
		assert.Equal(t, `<Error: module=Lexer,code=UnclosedComment,message="lexeme=,pos=(1,6)">`, err.Error())
	})
}

func TestPeekNextToken(t *testing.T) {
	source := `VAR
					number     : INTEGER;
					a, b, c, x : INTEGER;
					y          : REAL;`
	lexer := NewLexer(source)
	token, err := lexer.GetNextToken()
	assert.NoError(t, err)
	assert.Equal(t, Var, token.Kind)
	nextToken, err := PeekNextToken(lexer)
	assert.NoError(t, err)
	assert.Equal(t, ID, nextToken.Kind)
	expectedNextToken, err := lexer.GetNextToken()
	assert.NoError(t, err)
	assert.Equal(t, expectedNextToken, nextToken)
}
