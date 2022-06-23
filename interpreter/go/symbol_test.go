package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScopedSymbolTable_String(t *testing.T) {
	t.Run("global scope", func(t *testing.T) {
		globalScope := NewScopedSymbolTable("global", 1, nil)
		globalScope.Define(NewBuiltinTypeSymbol("INTEGER"))
		globalScope.Define(NewBuiltinTypeSymbol("REAL"))
		globalScope.Define(NewVarSymbol("a", NewBuiltinTypeSymbol("INTEGER")))
		wantOutput := `
SCOPE (SCOPED SYMBOL TABLE)
===========================
Scope name     : global
Scope level    : 1

Scope (Scoped symbol table) contents
------------------------------------
      a: <a:INTEGER>
   REAL: REAL
INTEGER: INTEGER
`
		assert.Equal(t, strings.TrimSpace(wantOutput), strings.TrimSpace(globalScope.String()))
	})

	t.Run("with enclosing scope", func(t *testing.T) {
		globalScope := NewScopedSymbolTable("global", 1, nil)
		globalScope.Define(NewBuiltinTypeSymbol("INTEGER"))
		globalScope.Define(NewBuiltinTypeSymbol("REAL"))

		innerScope := NewScopedSymbolTable("inner", 2, globalScope)
		innerScope.Define(NewVarSymbol("x", NewBuiltinTypeSymbol("INTEGER")))

		wantOutput := `
SCOPE (SCOPED SYMBOL TABLE)
===========================
Scope name     : inner
Scope level    : 2
Enclosing scope: global

Scope (Scoped symbol table) contents
------------------------------------
      x: <x:INTEGER>
`
		assert.Equal(t, strings.TrimSpace(wantOutput), strings.TrimSpace(innerScope.String()))
	})
}
