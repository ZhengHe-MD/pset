package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_Parse(t *testing.T) {
	tests := map[string]struct {
		givenSource string
		wantNode    ASTNode
	}{
		"noop": {
			givenSource: `
				PROGRAM Part10AST;
				BEGIN
				END.`,
			wantNode: &ProgramNode{
				name: "PART10AST",
				block: &BlockNode{
					declarations: nil,
					compoundStmt: &CompoundNode{
						children: []ASTNode{noop},
					},
				},
			},
		},
		"default case": {
			givenSource: `
				PROGRAM Part10AST;
				VAR
					a, b : INTEGER;
					y    : REAL;
				BEGIN {Part10AST}
					a := 2;
					b := 10 * a + 10 * a DIV 4;
					y := 20 / 7 + 3.14;
				END.  {Part10AST}
			`,
			wantNode: &ProgramNode{
				name: "PART10AST",
				block: &BlockNode{
					declarations: []ASTNode{
						&VarDeclNode{
							varNode: NewVarNode(NewDynamicToken(ID, "A", 4, 6)),
							typNode: NewTypNode(NewStaticToken(Integer, 4, 13)),
						},
						&VarDeclNode{
							varNode: NewVarNode(NewDynamicToken(ID, "B", 4, 9)),
							typNode: NewTypNode(NewStaticToken(Integer, 4, 13)),
						},
						&VarDeclNode{
							varNode: NewVarNode(NewDynamicToken(ID, "Y", 5, 6)),
							typNode: NewTypNode(NewStaticToken(Real, 5, 13)),
						},
					},
					compoundStmt: &CompoundNode{
						children: []ASTNode{
							&AssignNode{
								token: NewStaticToken(Assign, 7, 8),
								left:  NewVarNode(NewDynamicToken(ID, "A", 7, 6)),
								right: NewIntegerNumNode(NewDynamicToken(IntegerConst, "2", 7, 11)),
							},
							// b := 10 * a + 10 * a DIV 4;
							&AssignNode{
								token: NewStaticToken(Assign, 8, 8),
								left:  NewVarNode(NewDynamicToken(ID, "B", 8, 6)),
								right: &BinOpNode{
									token: NewStaticToken(Plus, 8, 18),
									left: &BinOpNode{
										token: NewStaticToken(Mul, 8, 14),
										left:  NewIntegerNumNode(NewDynamicToken(IntegerConst, "10", 8, 11)),
										right: NewVarNode(NewDynamicToken(ID, "A", 8, 16)),
										op:    Mul,
									},
									right: &BinOpNode{
										token: NewStaticToken(IntegerDiv, 8, 27),
										left: &BinOpNode{
											token: NewStaticToken(Mul, 8, 23),
											left:  NewIntegerNumNode(NewDynamicToken(IntegerConst, "10", 8, 20)),
											right: NewVarNode(NewDynamicToken(ID, "A", 8, 25)),
											op:    Mul,
										},
										right: NewIntegerNumNode(NewDynamicToken(IntegerConst, "4", 8, 31)),
										op:    IntegerDiv,
									},
									op: Plus,
								},
							},
							&AssignNode{
								token: NewStaticToken(Assign, 9, 8),
								left:  NewVarNode(NewDynamicToken(ID, "Y", 9, 6)),
								right: &BinOpNode{
									token: NewStaticToken(Plus, 9, 18),
									left: &BinOpNode{
										token: NewStaticToken(FloatDiv, 9, 14),
										left:  NewIntegerNumNode(NewDynamicToken(IntegerConst, "20", 9, 11)),
										right: NewIntegerNumNode(NewDynamicToken(IntegerConst, "7", 9, 16)),
										op:    FloatDiv,
									},
									right: NewRealNumNode(NewDynamicToken(RealConst, "3.14", 9, 20)),
									op:    Plus,
								},
							},
							noop,
						},
					},
				},
			},
		},
		"part12 case": {
			givenSource: `
				PROGRAM Part12;
				VAR
					a : INTEGER;
				PROCEDURE P1;
				VAR
					a : REAL;
					k : INTEGER;
					PROCEDURE P2;
					VAR
						a, z : INTEGER;
					BEGIN {P2}
						z := 777;
					END;  {P2}
				BEGIN {P1}
				END;  {P1}
				BEGIN {Part12}
					a := 10;
				END.  {Part12}
			`,
			wantNode: &ProgramNode{
				name: "PART12",
				block: &BlockNode{
					declarations: []ASTNode{
						&VarDeclNode{
							varNode: NewVarNode(NewDynamicToken(ID, "A", 4, 6)),
							typNode: NewTypNode(NewStaticToken(Integer, 4, 10)),
						},
						&ProcedureDeclNode{
							name: "P1",
							block: &BlockNode{
								declarations: []ASTNode{
									&VarDeclNode{
										varNode: NewVarNode(NewDynamicToken(ID, "A", 7, 6)),
										typNode: NewTypNode(NewStaticToken(Real, 7, 10)),
									},
									&VarDeclNode{
										varNode: NewVarNode(NewDynamicToken(ID, "K", 8, 6)),
										typNode: NewTypNode(NewStaticToken(Integer, 8, 10)),
									},
									&ProcedureDeclNode{
										name: "P2",
										block: &BlockNode{
											declarations: []ASTNode{
												&VarDeclNode{
													varNode: NewVarNode(NewDynamicToken(ID, "A", 11, 7)),
													typNode: NewTypNode(NewStaticToken(Integer, 11, 14)),
												},
												&VarDeclNode{
													varNode: NewVarNode(NewDynamicToken(ID, "Z", 11, 10)),
													typNode: NewTypNode(NewStaticToken(Integer, 11, 14)),
												},
											},
											compoundStmt: &CompoundNode{
												children: []ASTNode{
													&AssignNode{
														token: NewStaticToken(Assign, 13, 9),
														left:  NewVarNode(NewDynamicToken(ID, "Z", 13, 7)),
														right: NewIntegerNumNode(NewDynamicToken(IntegerConst, "777", 13, 12)),
													},
													noop,
												},
											},
										},
									},
								},
								compoundStmt: &CompoundNode{children: []ASTNode{noop}},
							},
						},
					},
					compoundStmt: &CompoundNode{
						children: []ASTNode{
							&AssignNode{
								token: NewStaticToken(Assign, 18, 8),
								left:  NewVarNode(NewDynamicToken(ID, "A", 18, 6)),
								right: NewIntegerNumNode(NewDynamicToken(IntegerConst, "10", 18, 11)),
							},
							noop,
						},
					},
				},
			},
		},
		"part 13: syntactically correct": {
			givenSource: `
				program Main;
					var x : integer;
				begin
					x := y;
				end.
			`,
			wantNode: &ProgramNode{
				name: "MAIN",
				block: &BlockNode{
					declarations: []ASTNode{
						&VarDeclNode{
							varNode: NewVarNode(NewDynamicToken(ID, "X", 3, 10)),
							typNode: NewTypNode(NewStaticToken(Integer, 3, 14)),
						},
					},
					compoundStmt: &CompoundNode{
						children: []ASTNode{
							&AssignNode{
								token: NewStaticToken(Assign, 5, 8),
								left:  NewVarNode(NewDynamicToken(ID, "X", 5, 6)),
								right: NewVarNode(NewDynamicToken(ID, "Y", 5, 11)),
							},
							noop,
						},
					},
				},
			},
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
			wantNode: &ProgramNode{
				name: "MAIN",
				block: &BlockNode{
					declarations: []ASTNode{
						&VarDeclNode{
							varNode: NewVarNode(NewDynamicToken(ID, "X", 3, 10)),
							typNode: NewTypNode(NewStaticToken(Real, 3, 16)),
						},
						&VarDeclNode{
							varNode: NewVarNode(NewDynamicToken(ID, "Y", 3, 13)),
							typNode: NewTypNode(NewStaticToken(Real, 3, 16)),
						},
						&ProcedureDeclNode{
							name: "ALPHA",
							params: []*ParamNode{
								NewParamNode(
									NewVarNode(NewDynamicToken(ID, "A", 4, 22)),
									NewTypNode(NewStaticToken(Integer, 4, 26))),
							},
							block: &BlockNode{
								declarations: []ASTNode{
									NewVarDeclNode(
										NewVarNode(NewDynamicToken(ID, "Y", 5, 11)),
										NewTypNode(NewStaticToken(Integer, 5, 15))),
								},
								// x := a + x + y;
								compoundStmt: &CompoundNode{
									children: []ASTNode{
										&AssignNode{
											token: NewStaticToken(Assign, 7, 9),
											left:  NewVarNode(NewDynamicToken(ID, "X", 7, 7)),
											right: &BinOpNode{
												token: NewStaticToken(Plus, 7, 18),
												left: &BinOpNode{
													token: NewStaticToken(Plus, 7, 14),
													left:  NewVarNode(NewDynamicToken(ID, "A", 7, 12)),
													right: NewVarNode(NewDynamicToken(ID, "X", 7, 16)),
													op:    Plus,
												},
												right: NewVarNode(NewDynamicToken(ID, "Y", 7, 20)),
												op:    Plus,
											},
										},
										noop,
									},
								},
							},
						},
					},
					compoundStmt: &CompoundNode{
						children: []ASTNode{noop},
					},
				},
			},
		},
		"multiple varDecls": {
			givenSource: `
				PROGRAM Part10AST;
					VAR x : INTEGER;
					VAR y, z : INTEGER;
				BEGIN
				END.`,
			wantNode: &ProgramNode{
				name: "PART10AST",
				block: &BlockNode{
					declarations: []ASTNode{
						&VarDeclNode{
							varNode: NewVarNode(NewDynamicToken(ID, "X", 3, 10)),
							typNode: NewTypNode(NewStaticToken(Integer, 3, 14)),
						},
						&VarDeclNode{
							varNode: NewVarNode(NewDynamicToken(ID, "Y", 4, 10)),
							typNode: NewTypNode(NewStaticToken(Integer, 4, 17)),
						},
						&VarDeclNode{
							varNode: NewVarNode(NewDynamicToken(ID, "Z", 4, 13)),
							typNode: NewTypNode(NewStaticToken(Integer, 4, 17)),
						},
					},
					compoundStmt: &CompoundNode{
						children: []ASTNode{noop},
					},
				},
			},
		},
		"part16: procedure call": {
			givenSource: `
				program Main;
				procedure Alpha(a : integer; b : integer);
					var x : integer;
				begin
					x := (a + b) * 2;
				end;
				begin { Main }
					Alpha(3 + 5, 7);  { procedure call }
				end.  { Main }
			`,
			wantNode: &ProgramNode{
				name: "MAIN",
				block: &BlockNode{
					declarations: []ASTNode{
						&ProcedureDeclNode{
							name: "ALPHA",
							params: []*ParamNode{
								NewParamNode(
									NewVarNode(NewDynamicToken(ID, "A", 3, 21)),
									NewTypNode(NewStaticToken(Integer, 3, 25))),
								NewParamNode(
									NewVarNode(NewDynamicToken(ID, "B", 3, 34)),
									NewTypNode(NewStaticToken(Integer, 3, 38))),
							},
							block: &BlockNode{
								declarations: []ASTNode{
									NewVarDeclNode(
										NewVarNode(NewDynamicToken(ID, "X", 4, 10)),
										NewTypNode(NewStaticToken(Integer, 4, 14))),
								},
								compoundStmt: &CompoundNode{
									children: []ASTNode{
										&AssignNode{
											token: NewStaticToken(Assign, 6, 8),
											left:  NewVarNode(NewDynamicToken(ID, "X", 6, 6)),
											right: &BinOpNode{
												token: NewStaticToken(Mul, 6, 19),
												left: &BinOpNode{
													token: NewStaticToken(Plus, 6, 14),
													left:  NewVarNode(NewDynamicToken(ID, "A", 6, 12)),
													right: NewVarNode(NewDynamicToken(ID, "B", 6, 16)),
													op:    Plus,
												},
												right: NewIntegerNumNode(NewDynamicToken(IntegerConst, "2", 6, 21)),
												op:    Mul,
											},
										},
										noop,
									},
								},
							},
						},
					},
					compoundStmt: &CompoundNode{
						children: []ASTNode{
							&ProcedureCallNode{
								token: NewDynamicToken(ID, "ALPHA", 9, 6),
								name:  "ALPHA",
								actualParams: []ASTNode{
									NewBinOpNode(
										NewStaticToken(Plus, 9, 14),
										NewIntegerNumNode(NewDynamicToken(IntegerConst, "3", 9, 12)),
										NewIntegerNumNode(NewDynamicToken(IntegerConst, "5", 9, 16))),
									NewIntegerNumNode(NewDynamicToken(IntegerConst, "7", 9, 19)),
								},
							},
							noop,
						},
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lexer := NewLexer(tc.givenSource)
			parser, err := NewParser(lexer)
			assert.NoError(t, err)
			node, err := parser.Parse()
			assert.NoError(t, err)
			assert.Equal(t, tc.wantNode, node)
		})
	}
}
