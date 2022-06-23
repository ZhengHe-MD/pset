package main

import (
	"fmt"
	"strconv"
)

type Visitor interface {
	Before(node ASTNode) (bool, error)
	After(node ASTNode) error
}

type ASTNode interface{}

var (
	_ ASTNode = (*ProgramNode)(nil)
	_ ASTNode = (*BlockNode)(nil)
	_ ASTNode = (*ProcedureDeclNode)(nil)
	_ ASTNode = (*ParamNode)(nil)
	_ ASTNode = (*VarDeclNode)(nil)
	_ ASTNode = (*CompoundNode)(nil)
	_ ASTNode = (*AssignNode)(nil)
	_ ASTNode = (*ProcedureCallNode)(nil)
	_ ASTNode = (*VarNode)(nil)
	_ ASTNode = (*TypNode)(nil)
	_ ASTNode = (*NumNode)(nil)
	_ ASTNode = (*UnaryOpNode)(nil)
	_ ASTNode = (*BinOpNode)(nil)
	_ ASTNode = (*NoopNode)(nil)
)

type ProgramNode struct {
	name  string
	block *BlockNode
}

func NewProcedureDeclNode(name string, params []*ParamNode, block *BlockNode) *ProcedureDeclNode {
	return &ProcedureDeclNode{
		name:   name,
		params: params,
		block:  block,
	}
}

type ProcedureDeclNode struct {
	name   string
	params []*ParamNode
	block  *BlockNode
}

func NewParamNode(varNode *VarNode, typNode *TypNode) *ParamNode {
	return &ParamNode{
		varNode: varNode,
		typNode: typNode,
	}
}

type ParamNode struct {
	varNode *VarNode
	typNode *TypNode
}

type BlockNode struct {
	declarations []ASTNode
	compoundStmt *CompoundNode
}

func NewVarDeclNode(varNode *VarNode, typNode *TypNode) *VarDeclNode {
	return &VarDeclNode{varNode: varNode, typNode: typNode}
}

type VarDeclNode struct {
	varNode *VarNode
	typNode *TypNode
}

func NewTypNode(token *Token) *TypNode {
	return &TypNode{
		token: token,
		value: token.Value,
	}
}

type TypNode struct {
	token *Token
	value string
}

type CompoundNode struct {
	children []ASTNode
}

type AssignNode struct {
	token *Token
	left  *VarNode
	right ASTNode
}

func NewProcedureCallNode(token *Token, actualParams []ASTNode) *ProcedureCallNode {
	return &ProcedureCallNode{
		token:        token,
		name:         token.Value,
		actualParams: actualParams,
	}
}

type ProcedureCallNode struct {
	token        *Token
	name         string
	actualParams []ASTNode
	procSymbol   *ProcedureSymbol
}

func NewVarNode(token *Token) *VarNode {
	return &VarNode{
		token: token,
		value: token.Value,
	}
}

type VarNode struct {
	token *Token
	value string
}

func NewIntegerNumNode(token *Token) *NumNode {
	intValue, err := strconv.Atoi(token.Value)
	if err != nil {
		panic(fmt.Sprintf("invalid integer token: %s", token))
	}
	return &NumNode{
		token:    token,
		intValue: intValue,
	}
}

func NewRealNumNode(token *Token) *NumNode {
	floatValue, err := strconv.ParseFloat(token.Value, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid real token: %s", token))
	}
	return &NumNode{
		token:      token,
		floatValue: floatValue,
	}
}

type NumNode struct {
	token      *Token
	intValue   int
	floatValue float64
}

type UnaryOpNode struct {
	token   *Token
	operand ASTNode
	op      TokenKind
}

func NewBinOpNode(token *Token, left, right ASTNode) *BinOpNode {
	return &BinOpNode{
		token: token,
		left:  left,
		right: right,
		op:    token.Kind,
	}
}

type BinOpNode struct {
	token *Token
	left  ASTNode
	right ASTNode
	op    TokenKind
}

type NoopNode struct{}

var noop = &NoopNode{}

func NewParser(lexer *Lexer) (parser *Parser, err error) {
	currToken, err := lexer.GetNextToken()
	if err != nil {
		return
	}
	parser = &Parser{
		lexer:     lexer,
		currToken: currToken,
	}
	return
}

// Parser implements the following Pascal CFG:
//
// program : PROGRAM variable SEMI block DOT
// block : declarations compound_statement
// declarations: (declaration)*
// declaration: VAR (variable_declaration SEMI)+
// 			   | (PROCEDURE ID (LPAREN formal_parameter_list RPAREN)? SEMI block SEMI)*
//             | empty
// variable_declaration : ID (COMMA ID)* COLON type_spec
// formal_parameter_list: formal_parameters
// 						| formal_parameters SEMI formal_parameter_list
// formal_parameters: ID (COMMA ID)* COLON type_spec
// type_spec : INTEGER | REAL
// compound_statement : BEGIN statement_list END
// statement_list : statement
// 				  | statement SEMI statement_list
// statement : compound_statement
// 			 | procedure_call_statement
// 			 | assignment_statement
// 			 | empty
// procedure_call_statement : ID LPAREN (expr (COMMA expr)*)? RPAREN
// assignment_statement : variable ASSIGN expr
// empty :
// expr: term ((PLUS | MINUS) term)*
// term: factor ((MUL | INTEGER_DIV | FLOAT_DIV) factor)*
// factor : PLUS factor
// 		  | MINUS factor
// 		  | INTEGER_CONST
// 		  | REAL_CONST
// 		  | LPAREN expr RPAREN
// 		  | variable
// variable: ID
type Parser struct {
	lexer     *Lexer
	currToken *Token
}

func (p *Parser) Parse() (node ASTNode, err error) {
	if node, err = p.program(); err != nil {
		return
	}

	if p.currToken.Kind != EOF {
		err = p.error(ErrorCodeUnexpectedToken)
		return
	}
	return
}

// program : PROGRAM variable SEMI block DOT
func (p *Parser) program() (node *ProgramNode, err error) {
	if err = p.eat(Program); err != nil {
		return
	}
	var varNode *VarNode
	if varNode, err = p.variable(); err != nil {
		return
	}
	if err = p.eat(Semi); err != nil {
		return
	}
	var blockNode *BlockNode
	if blockNode, err = p.block(); err != nil {
		return
	}
	if err = p.eat(Dot); err != nil {
		return
	}
	node = &ProgramNode{
		name:  varNode.value,
		block: blockNode,
	}
	return
}

// block : declarations compound_statement
func (p *Parser) block() (node *BlockNode, err error) {
	declarations, err := p.declarations()
	if err != nil {
		return
	}
	compoundStmt, err := p.compoundStmt()
	if err != nil {
		return
	}
	node = &BlockNode{
		declarations: declarations,
		compoundStmt: compoundStmt,
	}
	return
}

// declarations: (declaration)*
func (p *Parser) declarations() (nodes []ASTNode, err error) {
	for p.currToken.Kind == Var || p.currToken.Kind == Procedure {
		var declNodes []ASTNode
		if declNodes, err = p.declaration(); err != nil {
			return
		}
		nodes = append(nodes, declNodes...)
	}
	return
}

// declaration: VAR (variable_declaration SEMI)+
// 			   | (PROCEDURE ID (LPAREN formal_parameter_list RPAREN)? SEMI block SEMI)*
//             | empty
func (p *Parser) declaration() (nodes []ASTNode, err error) {
	if p.currToken.Kind == Var {
		if err = p.eat(Var); err != nil {
			return
		}

		var varDeclNodes []*VarDeclNode
		if varDeclNodes, err = p.variableDeclaration(); err != nil {
			return
		}
		for _, varDeclNode := range varDeclNodes {
			nodes = append(nodes, varDeclNode)
		}
		if err = p.eat(Semi); err != nil {
			return
		}
		for p.currToken.Kind == ID {
			if varDeclNodes, err = p.variableDeclaration(); err != nil {
				return
			}
			for _, varDeclNode := range varDeclNodes {
				nodes = append(nodes, varDeclNode)
			}
			if err = p.eat(Semi); err != nil {
				return
			}
		}
	}

	for p.currToken.Kind == Procedure {
		procedureDeclNode := &ProcedureDeclNode{}

		if err = p.eat(Procedure); err != nil {
			return
		}
		procedureDeclNode.name = p.currToken.Value
		if err = p.eat(ID); err != nil {
			return
		}
		if p.currToken.Kind == LParen {
			if err = p.eat(LParen); err != nil {
				return
			}
			if procedureDeclNode.params, err = p.formalParameterList(); err != nil {
				return
			}
			if err = p.eat(RParen); err != nil {
				return
			}
		}
		if err = p.eat(Semi); err != nil {
			return
		}
		if procedureDeclNode.block, err = p.block(); err != nil {
			return
		}
		if err = p.eat(Semi); err != nil {
			return
		}
		nodes = append(nodes, procedureDeclNode)
	}
	return
}

// variable_declaration : ID (COMMA ID)* COLON type_spec
func (p *Parser) variableDeclaration() (nodes []*VarDeclNode, err error) {
	varNodes := []*VarNode{NewVarNode(p.currToken)}
	if err = p.eat(ID); err != nil {
		return
	}
	for p.currToken.Kind == Comma {
		if err = p.eat(Comma); err != nil {
			return
		}
		varNodes = append(varNodes, NewVarNode(p.currToken))
		if err = p.eat(ID); err != nil {
			return
		}
	}
	if err = p.eat(Colon); err != nil {
		return
	}

	var typNode *TypNode
	if typNode, err = p.typeSpec(); err != nil {
		return
	}

	nodes = make([]*VarDeclNode, len(varNodes))
	for i, varNode := range varNodes {
		nodes[i] = NewVarDeclNode(varNode, typNode)
	}
	return
}

// formal_parameter_list: formal_parameters
// 						| formal_parameters SEMI formal_parameter_list
func (p *Parser) formalParameterList() (nodes []*ParamNode, err error) {
	paramNodes, err := p.formalParameters()
	if err != nil {
		return
	}
	nodes = append(nodes, paramNodes...)
	for p.currToken.Kind == Semi {
		if err = p.eat(Semi); err != nil {
			return
		}
		if paramNodes, err = p.formalParameterList(); err != nil {
			return
		}
		nodes = append(nodes, paramNodes...)
	}
	return
}

// formal_parameters: ID (COMMA ID)* COLON type_spec
func (p *Parser) formalParameters() (nodes []*ParamNode, err error) {
	nodes = append(nodes, NewParamNode(NewVarNode(p.currToken), nil))
	if err = p.eat(ID); err != nil {
		return
	}
	for p.currToken.Kind == Comma {
		if err = p.eat(Comma); err != nil {
			return
		}
		nodes = append(nodes, NewParamNode(NewVarNode(p.currToken), nil))
		if err = p.eat(ID); err != nil {
			return
		}
	}
	if err = p.eat(Colon); err != nil {
		return
	}

	var typNode *TypNode
	if typNode, err = p.typeSpec(); err != nil {
		return
	}
	for _, node := range nodes {
		node.typNode = typNode
	}
	return
}

// type_spec : INTEGER | REAL
func (p *Parser) typeSpec() (node *TypNode, err error) {
	if p.currToken.Kind != Integer && p.currToken.Kind != Real {
		err = p.error(ErrorCodeUnexpectedToken)
		return
	}
	node = NewTypNode(p.currToken)
	err = p.eat(p.currToken.Kind)
	return
}

// compound_statement : BEGIN statement_list END
func (p *Parser) compoundStmt() (node *CompoundNode, err error) {
	if err = p.eat(Begin); err != nil {
		return
	}
	nodes, err := p.stmtList()
	if err != nil {
		return
	}

	if err = p.eat(End); err != nil {
		return
	}

	node = &CompoundNode{children: nodes}
	return
}

// statement_list : statement
// 					| statement SEMI statement_list
func (p *Parser) stmtList() (nodes []ASTNode, err error) {
	var node ASTNode
	if node, err = p.stmt(); err != nil {
		return
	}
	nodes = append(nodes, node)
	var neighboringNodes []ASTNode
	for p.currToken.Kind == Semi {
		if err = p.eat(Semi); err != nil {
			return
		}

		if neighboringNodes, err = p.stmtList(); err != nil {
			return
		}
		nodes = append(nodes, neighboringNodes...)
	}

	return
}

// statement : compound_statement
// 			 | procedure_call_statement
// 			 | assignment_statement
// 			 | empty
func (p *Parser) stmt() (node ASTNode, err error) {
	if p.currToken.Kind == Begin {
		return p.compoundStmt()
	}

	var nextToken *Token
	if p.currToken.Kind == ID {
		if nextToken, err = PeekNextToken(p.lexer); err != nil {
			return
		}
		if nextToken.Kind == LParen {
			return p.procedureCallStmt()
		}
		return p.assignStmt()
	}

	node = p.empty()
	return
}

// procedure_call_statement : ID LPAREN (expr (COMMA expr)*)? RPAREN
func (p *Parser) procedureCallStmt() (node ASTNode, err error) {
	token := p.currToken
	if err = p.eats(ID, LParen); err != nil {
		return
	}

	argument, err := p.expr()
	if err != nil {
		return
	}
	arguments := []ASTNode{argument}
	for p.currToken.Kind == Comma {
		if err = p.eat(Comma); err != nil {
			return
		}

		if argument, err = p.expr(); err != nil {
			return
		}

		arguments = append(arguments, argument)
	}

	if err = p.eat(RParen); err != nil {
		return
	}

	return NewProcedureCallNode(token, arguments), nil
}

// assignment_statement : variable ASSIGN expr
func (p *Parser) assignStmt() (node ASTNode, err error) {
	id, err := p.variable()
	if err != nil {
		return
	}

	token := p.currToken
	if err = p.eat(Assign); err != nil {
		return
	}
	expr, err := p.expr()
	if err != nil {
		return
	}

	node = &AssignNode{
		token: token,
		left:  id,
		right: expr,
	}
	return
}

func (p *Parser) empty() ASTNode {
	return noop
}

// expr : term ((PLUS|MINUS) term)*
func (p *Parser) expr() (node ASTNode, err error) {
	if node, err = p.term(); err != nil {
		return
	}
	for p.currToken.Kind == Plus || p.currToken.Kind == Minus {
		token := p.currToken
		if err = p.eat(token.Kind); err != nil {
			return
		}
		var term ASTNode
		if term, err = p.term(); err != nil {
			return
		}
		node = &BinOpNode{
			left:  node,
			right: term,
			token: token,
			op:    token.Kind,
		}
	}
	return
}

// term: factor ((MUL | INTEGER_DIV | FLOAT_DIV) factor)*
func (p *Parser) term() (node ASTNode, err error) {
	if node, err = p.factor(); err != nil {
		return
	}

	for p.currToken.Kind == Mul || p.currToken.Kind == IntegerDiv || p.currToken.Kind == FloatDiv {
		token := p.currToken
		if err = p.eat(p.currToken.Kind); err != nil {
			return
		}
		var factorNode ASTNode
		if factorNode, err = p.factor(); err != nil {
			return
		}
		node = &BinOpNode{
			left:  node,
			right: factorNode,
			token: token,
			op:    token.Kind,
		}
	}
	return
}

// factor : PLUS factor
// 		  | MINUS factor
// 		  | INTEGER_CONST
// 		  | REAL_CONST
// 		  | LPAREN expr RPAREN
// 		  | variable
func (p *Parser) factor() (node ASTNode, err error) {
	token := p.currToken
	switch token.Kind {
	case Plus:
		if err = p.eat(Plus); err != nil {
			return
		}
		var operand ASTNode
		if operand, err = p.factor(); err != nil {
			return
		}
		node = &UnaryOpNode{op: Plus, token: token, operand: operand}
		return
	case Minus:
		if err = p.eat(Minus); err != nil {
			return
		}
		var operand ASTNode
		if operand, err = p.factor(); err != nil {
			return
		}
		node = &UnaryOpNode{op: Minus, token: token, operand: operand}
		return
	case IntegerConst:
		if err = p.eat(IntegerConst); err != nil {
			return
		}
		var iv int
		if iv, err = strconv.Atoi(token.Value); err != nil {
			err = p.error(ErrorCodeUnexpectedToken)
			return
		}
		node = &NumNode{token: token, intValue: iv}
		return
	case RealConst:
		if err = p.eat(RealConst); err != nil {
			return
		}
		var fv float64
		if fv, err = strconv.ParseFloat(token.Value, 64); err != nil {
			err = p.error(ErrorCodeUnexpectedToken)
			return
		}
		node = &NumNode{token: token, floatValue: fv}
		return
	case LParen:
		if err = p.eat(LParen); err != nil {
			return
		}
		if node, err = p.expr(); err != nil {
			return
		}
		if err = p.eat(RParen); err != nil {
			return
		}
		return
	case ID:
		return p.variable()
	default:
		err = p.error(ErrorCodeUnexpectedToken)
		return
	}
}

func (p *Parser) variable() (node *VarNode, err error) {
	token := p.currToken
	if err = p.eat(ID); err != nil {
		return
	}
	node = &VarNode{token: token, value: token.Value}
	return
}

func (p *Parser) eat(kind TokenKind) (err error) {
	if p.currToken.Kind != kind {
		err = p.error(ErrorCodeUnexpectedToken)
		return
	}
	p.currToken, err = p.lexer.GetNextToken()
	return
}

func (p *Parser) eats(kinds ...TokenKind) (err error) {
	for _, kind := range kinds {
		if err = p.eat(kind); err != nil {
			return
		}
	}
	return
}

func (p *Parser) error(code ErrorCode) error {
	return Error{
		Code:    code,
		Module:  ModuleParser,
		Message: fmt.Sprintf("code=%s,token=%s", code, p.currToken),
	}
}
