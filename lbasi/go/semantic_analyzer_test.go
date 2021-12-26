package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSemanticAnalyzer_Visit(t *testing.T) {
	tests := map[string]struct {
		givenSource      string
		wantError        bool
		wantErrorMessage string
	}{
		"part13: var declarations only": {
			givenSource: `
				program SymTab2;
					var x, y : INTEGER;
				begin
				end.
			`,
		},
		"part13: declarations + references": {
			givenSource: `
				program SymTab4;
					var x, y : integer;

				begin
					x := x + y;
				end.
			`,
		},
		"part13: reference without declaration": {
			givenSource: `
				program SymTab5;
				var x : integer;
				begin
					x := y;
				end.
			`,
			wantError:        true,
			wantErrorMessage: `<Error: module=SemanticAnalyzer,code=IdNotFound,message="token:(kind=ID,value=Y,pos=(5,11))">`,
		},
		"part13: duplication declarations": {
			givenSource: `
				program SymTab6;
				var
					x, y : integer;
					y : real;
				begin
					x := x + y;
				end.
			`,
			wantError:        true,
			wantErrorMessage: `<Error: module=SemanticAnalyzer,code=DuplicateId,message="token:(kind=ID,value=Y,pos=(5,6))">`,
		},
		"part14: nested scopes": {
			givenSource: `
				program Main;
					var x, y: real;
					procedure Alpha(a : integer);
						var y : integer;
					begin
						x := a + x + y;
					end;
				begin { Main }
				end.  { Main }
			`,
		},
		"part16: actualParams mismatch case 1": {
			givenSource: `
				program Main;
					 procedure Alpha(a : integer; b : integer);
						 var x : integer;
					 begin
						 x := (a + b) * 2;
					 end;
				begin { Main }
					Alpha(1, 3 + 5, 7);  { procedure call }
				end.  { Main }`,
			wantError:        true,
			wantErrorMessage: `<Error: module=SemanticAnalyzer,code=ArgumentsMismatch,message="token:(kind=ID,value=ALPHA,pos=(9,6))">`,
		},
		"part16: actualParams mismatch case 2": {
			givenSource: `
				program Main;
					 procedure Alpha(a : integer; b : integer);
						 var x : integer;
					 begin
						 x := (a + b) * 2;
					 end;
				begin { Main }
					Alpha(1);  { procedure call }
				end.  { Main }`,
			wantError:        true,
			wantErrorMessage: `<Error: module=SemanticAnalyzer,code=ArgumentsMismatch,message="token:(kind=ID,value=ALPHA,pos=(9,6))">`,
		},
		// TODO: when we support type definition syntax
		//"declare a symbol with unknown type": {},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lexer := NewLexer(tc.givenSource)
			parser, err := NewParser(lexer)
			assert.NoError(t, err)
			root, err := parser.Parse()
			assert.NoError(t, err)
			analyzer := NewSemanticAnalyzer()
			err = Walk(analyzer, root)
			if tc.wantError {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErrorMessage, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
