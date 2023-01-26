package main

import (
	"fmt"
	"sort"
	"strings"
)

type Symbol interface {
	fmt.Stringer
	GetName() string
	SetScopeLevel(int)
}

type baseSymbol struct {
	scopeLevel int
}

func (sb *baseSymbol) SetScopeLevel(scopeLevel int) {
	sb.scopeLevel = scopeLevel
}

var _ Symbol = (*BuiltinTypeSymbol)(nil)
var _ Symbol = (*VarSymbol)(nil)
var _ Symbol = (*ProcedureSymbol)(nil)

func NewBuiltinTypeSymbol(name string) *BuiltinTypeSymbol {
	return &BuiltinTypeSymbol{
		name: name,
	}
}

type BuiltinTypeSymbol struct {
	baseSymbol
	name string
}

func (bs *BuiltinTypeSymbol) GetName() string {
	return bs.name
}

func (bs *BuiltinTypeSymbol) String() string {
	return bs.name
}

func NewVarSymbol(name string, typ Symbol) *VarSymbol {
	return &VarSymbol{
		name: name,
		typ:  typ,
	}
}

type VarSymbol struct {
	baseSymbol
	name string
	typ  Symbol
}

func (vs *VarSymbol) GetName() string {
	return vs.name
}

func (vs *VarSymbol) String() string {
	return fmt.Sprintf("<%s:%s>", vs.name, vs.typ)
}

func NewProcedureSymbol(name string, formalParams []*VarSymbol) *ProcedureSymbol {
	return &ProcedureSymbol{
		name:         name,
		formalParams: formalParams,
	}
}

type ProcedureSymbol struct {
	baseSymbol
	name         string
	formalParams []*VarSymbol
	blockNode    *BlockNode
}

func (ps *ProcedureSymbol) String() string {
	var params []string
	for _, p := range ps.formalParams {
		params = append(params, p.String())
	}
	return fmt.Sprintf("<%s:%s>", ps.name, strings.Join(params, ";"))
}

func (ps *ProcedureSymbol) GetName() string {
	return ps.name
}

func (ps *ProcedureSymbol) SetBlockNode(node *BlockNode) {
	ps.blockNode = node
}

func NewScopedSymbolTable(
	name string,
	level int,
	enclosingScope *ScopedSymbolTable,
) *ScopedSymbolTable {
	return &ScopedSymbolTable{
		name:           name,
		level:          level,
		enclosingScope: enclosingScope,
		symbols:        make(map[string]Symbol),
	}
}

type ScopedSymbolTable struct {
	name           string
	level          int
	enclosingScope *ScopedSymbolTable
	symbols        map[string]Symbol
}

func (st *ScopedSymbolTable) String() string {
	output := strings.Builder{}
	h1 := "SCOPE (SCOPED SYMBOL TABLE)"
	output.WriteString(h1)
	output.WriteRune('\n')
	for i := 0; i < len(h1); i++ {
		output.WriteRune('=')
	}
	output.WriteRune('\n')
	output.WriteString(fmt.Sprintf("%-15s: %s\n", "Scope name", st.name))
	output.WriteString(fmt.Sprintf("%-15s: %d\n", "Scope level", st.level))
	if st.enclosingScope != nil {
		output.WriteString(fmt.Sprintf("%-15s: %s\n", "Enclosing scope", st.enclosingScope.name))
	}
	output.WriteRune('\n')
	h2 := "Scope (Scoped symbol table) contents"
	output.WriteString(h2)
	output.WriteRune('\n')
	for i := 0; i < len(h2); i++ {
		output.WriteRune('-')
	}
	output.WriteRune('\n')
	var symbolRows []string
	for k, v := range st.symbols {
		symbolRows = append(symbolRows, fmt.Sprintf("%7s: %s", k, v.String()))
	}
	sort.Strings(symbolRows)
	output.WriteString(strings.Join(symbolRows, "\n"))
	output.WriteRune('\n')
	return output.String()
}

func (st *ScopedSymbolTable) Define(s Symbol) {
	s.SetScopeLevel(st.level)
	st.symbols[s.GetName()] = s
	return
}

func (st *ScopedSymbolTable) Lookup(name string, currentOnly bool) (s Symbol, ok bool) {
	if s, ok = st.symbols[name]; ok {
		return
	}

	if currentOnly || st.enclosingScope == nil {
		return
	}

	s, ok = st.enclosingScope.Lookup(name, false)
	return
}
