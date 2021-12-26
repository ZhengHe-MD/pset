package main

import (
	"fmt"
	"io"
	"strings"
)

func NewScopeMarker(writer io.Writer) *ScopeMarker {
	return &ScopeMarker{
		writer:           writer,
		semanticAnalyzer: NewSemanticAnalyzer(),
	}
}

var _ Visitor = (*ScopeMarker)(nil)

type ScopeMarker struct {
	writer           io.Writer
	semanticAnalyzer *SemanticAnalyzer
	extraIndent      int
}

func (sm *ScopeMarker) Mark(source string) (err error) {
	lexer := NewLexer(source)
	parser, err := NewParser(lexer)
	if err != nil {
		return
	}
	root, err := parser.Parse()
	if err != nil {
		return
	}
	return Walk(sm, root)
}

func (sm *ScopeMarker) Before(node ASTNode) (shouldStepIn bool, err error) {
	switch n := node.(type) {
	case *ProgramNode:
		sm.writeLine(fmt.Sprintf("%s %s;", TokenNames[Program], sm.withScope(n.name, 0)))
	case *BlockNode:
	case *VarDeclNode:
		sm.extraIndent += 1
		sm.writeLine(fmt.Sprintf("%s %s %s %s%s",
			TokenNames[Var],
			sm.withScope(n.varNode.value, 0),
			TokenNames[Colon],
			n.typNode.value,
			TokenNames[Semi]))
		sm.extraIndent -= 1
	case *ProcedureDeclNode:
		builder := strings.Builder{}
		builder.WriteString(TokenNames[Procedure])
		builder.WriteRune(' ')
		builder.WriteString(sm.withScope(n.name, 0))
		builder.WriteString(TokenNames[LParen])
		for i, param := range n.params {
			builder.WriteString(fmt.Sprintf("%s %s %s",
				sm.withScope(param.varNode.value, -1), TokenNames[Colon], param.typNode.value))
			if i+1 != len(n.params) {
				builder.WriteRune(';')
			}
		}
		builder.WriteString(TokenNames[RParen])
		builder.WriteString(TokenNames[Semi])
		sm.extraIndent += 1
		sm.writeLine(builder.String())
		sm.extraIndent -= 1
	case *CompoundNode:
		sm.writeLine(TokenNames[Begin])
	case *AssignNode:
		sm.writeLine(fmt.Sprintf("%s %s %s%s",
			sm.withLookupScope(n.left.value), n.token.Value, sm.expr(n.right), TokenNames[Semi]))
	}

	if _, err = sm.semanticAnalyzer.Before(node); err != nil {
		return
	}
	shouldStepIn = true
	return
}

func (sm *ScopeMarker) After(node ASTNode) (err error) {
	switch n := node.(type) {
	case *ProgramNode:
		builder := strings.Builder{}
		builder.WriteString(TokenNames[End])
		builder.WriteString(TokenNames[Dot])
		builder.WriteString(fmt.Sprintf(" {END OF %s}", n.name))
		sm.writeLine(builder.String())
	case *ProcedureDeclNode:
		builder := strings.Builder{}
		builder.WriteString(TokenNames[End])
		builder.WriteString(TokenNames[Semi])
		sm.writeLine(builder.String())
	case *BlockNode:
	case *CompoundNode:
	}
	return sm.semanticAnalyzer.After(node)
}

func (sm *ScopeMarker) expr(node ASTNode) string {
	switch n := node.(type) {
	case *NumNode:
		return n.token.Value
	case *UnaryOpNode:
		return "-" + sm.expr(n.operand)
	case *BinOpNode:
		return strings.Join([]string{
			sm.expr(n.left),
			n.token.Value,
			sm.expr(n.right),
		}, " ")
	case *VarNode:
		return sm.withLookupScope(n.value)
	}
	panic("unreachable")
}

func (sm *ScopeMarker) writeLine(s string) {
	indent := strings.Builder{}
	for i := 1; i < sm.semanticAnalyzer.currentScope.level+sm.extraIndent; i++ {
		indent.WriteRune('\t')
	}
	sm.writer.Write([]byte(indent.String() + s + "\n"))
}

func (sm *ScopeMarker) withScope(s string, levelup int) string {
	return fmt.Sprintf("%s%d", s, sm.semanticAnalyzer.currentScope.level-levelup)
}

func (sm *ScopeMarker) withLookupScope(s string) string {
	var levelup int
	currScope := sm.semanticAnalyzer.currentScope
	_, ok := currScope.Lookup(s, true)
	for !ok {
		levelup += 1
		currScope = currScope.enclosingScope
		if currScope == nil {
			panic(fmt.Sprintf("symbol %s not found in enclosing scope", s))
		}
		_, ok = currScope.Lookup(s, true)
	}
	return sm.withScope(s, levelup)
}
