package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScopeMarker_Mark(t *testing.T) {
	tests := map[string]struct {
		givenSource string
		wantSource  string
	}{
		"single varDecl": {
			givenSource: `
				 program SymTab;
					  var x : integer;
				 begin
				 end.
			`,
			wantSource: `
				 PROGRAM SYMTAB0;
					VAR X1 : INTEGER;
				 BEGIN
				 END. {END OF SYMTAB}
			`,
		},
		"multiple varDecls": {
			givenSource: `
				 program SymTab;
					  var x, y : integer;
				 begin
				 end.
			`,
			wantSource: `
				 PROGRAM SYMTAB0;
					VAR X1 : INTEGER;
					VAR Y1 : INTEGER;
				 BEGIN
				 END. {END OF SYMTAB}
			`,
		},
		"part14-case": {
			givenSource: `
				program Main;
					var b, x, y : real;
					var z : integer;
					procedure AlphaA(a : integer);
						var b : integer;
						procedure Beta(c : integer);
							var y : integer;
						    procedure Gamma(c : integer);
								var x : integer;
							begin { Gamma }
								x := a + b + c + x + y + z;
							end;  { Gamma }
						begin { Beta }
						end;  { Beta }
					begin { AlphaA }
					end;  { AlphaA }
					procedure AlphaB(a : integer);
						var c : real;
					begin { AlphaB }
						c := a + b;
					end;  { AlphaB }
				begin { Main }
				end.  { Main }
			`,
			wantSource: `
				PROGRAM MAIN0;
					VAR B1 : REAL;
					VAR X1 : REAL;
					VAR Y1 : REAL;
					VAR Z1 : INTEGER;
					PROCEDURE ALPHAA1(A2 : INTEGER);
						VAR B2 : INTEGER;
						PROCEDURE BETA2(C3 : INTEGER);
							VAR Y3 : INTEGER;
						    PROCEDURE GAMMA3(C4 : INTEGER);
								VAR X4 : INTEGER;
							BEGIN
								X4 := A2 + B2 + C4 + X4 + Y3 + Z1;
							END;
						BEGIN
						END;
					BEGIN
					END;
					PROCEDURE ALPHAB1(A2 : INTEGER);
						VAR C2 : REAL;
					BEGIN
						C2 := A2 + B1;
					END;
				BEGIN
				END. {END OF MAIN}
			`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			w := bytes.NewBuffer(nil)
			marker := NewScopeMarker(w)
			err := marker.Mark(tc.givenSource)
			assert.NoError(t, err)

			assert.Equal(t, strings.Fields(tc.wantSource), strings.Fields(w.String()))
		})
	}
}
