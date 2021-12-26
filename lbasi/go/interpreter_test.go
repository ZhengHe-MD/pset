package main

//func TestInterpreter_Interpret(t *testing.T) {
//	tests := map[string]struct {
//		givenProgram    string
//		wantCallStack map[string]float64
//	}{
//		"noop": {
//			givenProgram: `
//				PROGRAM p;
//				BEGIN
//				END.
//			`,
//			wantGlobalScope: map[string]float64{},
//		},
//		"default test": {
//			givenProgram: `
//				PROGRAM p;
//				VAR
//					number, a, b, c, x : INTEGER;
//				BEGIN
//					BEGIN
//						number := 2;
//						a := number;
//						b := 10 * a + 10 * number div 4;
//						c := a - - b
//					END;
//					x := 11;
//				END.
//			`,
//			wantGlobalScope: map[string]float64{
//				"a": 2, "x": 11,
//				"c": 27, "b": 25,
//				"number": 2,
//			},
//		},
//		"case insensitivity": {
//			givenProgram: `
//				PROGRAM p;
//				VAR
//					number, a, b, c, x : INTEGER;
//				BEGIN
//					BEGIN
//						number := 2;
//						a := NumBer;
//						B := 10 * a + 10 * NUMBER Div 4;
//						c := a - - b
//					end;
//					x := 11;
//				END.
//			`,
//			wantGlobalScope: map[string]float64{
//				"a": 2, "x": 11,
//				"c": 27, "b": 25,
//				"number": 2,
//			},
//		},
//		"example from part10": {
//			givenProgram: `
//				PROGRAM Part10AST;
//				VAR
//					a, b : INTEGER;
//					y 	 : REAL;
//
//				BEGIN {Part10AST}
//					a := 2;
//					b := 10 * a + 10 * a DIV 4;
//					y := 20 / 7 + 3.14
//				END. {Part10AST}
//			`,
//			wantGlobalScope: map[string]float64{
//				"A": 2, "B": 25,
//				"Y": 20.0/7 + 3.14,
//			},
//		},
//	}
//
//	normalizeScope := func(m map[string]float64) map[string]float64 {
//		mm := make(map[string]float64, len(m))
//		for k, v := range m {
//			mm[normalizeKeyword(k)] = v
//		}
//		return mm
//	}
//
//	for name, tc := range tests {
//		t.Run(name, func(t *testing.T) {
//			it := NewInterpreter()
//			it.Interpret(tc.givenProgram)
//			assert.Equal(t,
//				normalizeScope(tc.wantGlobalScope),
//				normalizeScope(it.callStack))
//		})
//	}
//}
