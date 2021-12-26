package main

import (
	"fmt"
)

func NewSemanticAnalyzer() *SemanticAnalyzer {
	globalScope := NewScopedSymbolTable("global", 0, nil)
	globalScope.Define(NewBuiltinTypeSymbol("INTEGER"))
	globalScope.Define(NewBuiltinTypeSymbol("REAL"))

	return &SemanticAnalyzer{currentScope: globalScope}
}

var _ Visitor = (*SemanticAnalyzer)(nil)

type SemanticAnalyzer struct {
	currentScope *ScopedSymbolTable
}

func (s *SemanticAnalyzer) Before(node ASTNode) (shouldStepIn bool, err error) {
	switch n := node.(type) {
	case *ProgramNode:
		s.currentScope.Define(NewProcedureSymbol(n.name, nil))
		s.currentScope = NewScopedSymbolTable(n.name, s.currentScope.level+1, s.currentScope)
	case *ProcedureDeclNode:
		var procedureParamsSymbols []*VarSymbol
		procedureScope := NewScopedSymbolTable(n.name, s.currentScope.level+1, s.currentScope)
		for _, param := range n.params {
			typSymbol, ok := s.currentScope.Lookup(param.typNode.value, false)
			if !ok {
				err = s.error(ErrorCodeUnknownDataType, param.typNode.token)
				return
			}
			varSymbol := NewVarSymbol(param.varNode.value, typSymbol)
			procedureScope.Define(varSymbol)
			procedureParamsSymbols = append(procedureParamsSymbols, varSymbol)
		}
		// NOTE: inject block sub-AST into procedure symbol for interpretation usage.
		procedureSymbol := NewProcedureSymbol(n.name, procedureParamsSymbols)
		procedureSymbol.SetBlockNode(n.block)
		s.currentScope.Define(procedureSymbol)
		s.currentScope = procedureScope
	case *VarDeclNode:
		// check type exists
		typSymbol, ok := s.currentScope.Lookup(n.typNode.value, false)
		if !ok {
			err = s.error(ErrorCodeUnknownDataType, n.typNode.token)
			return
		}
		// check duplicate definitions
		if _, ok = s.currentScope.Lookup(n.varNode.value, true); ok {
			err = s.error(ErrorCodeDuplicateId, n.varNode.token)
			return
		}
		s.currentScope.Define(NewVarSymbol(n.varNode.value, typSymbol))
	case *VarNode:
		if _, ok := s.currentScope.Lookup(n.value, false); !ok {
			err = s.error(ErrorCodeIdNotFound, n.token)
			return
		}
	case *ProcedureCallNode:
		symbol, ok := s.currentScope.Lookup(n.name, false)
		if !ok {
			err = s.error(ErrorCodeIdNotFound, n.token)
			return
		}
		procedureSymbol := symbol.(*ProcedureSymbol)
		if len(procedureSymbol.formalParams) != len(n.actualParams) {
			err = s.error(ErrorCodeArgumentsMismatch, n.token)
			return
		}
		// NOTE: inject symbol info (signature) to AST.
		n.procSymbol = procedureSymbol
		return false, nil
	}
	return true, nil
}

func (s *SemanticAnalyzer) After(node ASTNode) (err error) {
	switch node.(type) {
	case *ProgramNode:
		s.currentScope = s.currentScope.enclosingScope
	case *ProcedureDeclNode:
		s.currentScope = s.currentScope.enclosingScope
	default:
	}
	return
}

func (s *SemanticAnalyzer) error(code ErrorCode, token *Token) error {
	return Error{
		Code:    code,
		Module:  ModuleSemanticAnalyzer,
		Message: fmt.Sprintf("token:%s", token),
	}
}
