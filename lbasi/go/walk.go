package main

func Walk(visitor Visitor, node ASTNode) (err error) {
	ok, err := visitor.Before(node)
	if err != nil || !ok {
		return
	}

	switch n := node.(type) {
	case *ProgramNode:
		if err = Walk(visitor, n.block); err != nil {
			return
		}
	case *BlockNode:
		for _, decl := range n.declarations {
			if err = Walk(visitor, decl); err != nil {
				return
			}
		}
		if err = Walk(visitor, n.compoundStmt); err != nil {
			return
		}
	case *VarDeclNode:
		if err = Walk(visitor, n.varNode); err != nil {
			return
		}
		if err = Walk(visitor, n.typNode); err != nil {
			return
		}
	case *ProcedureDeclNode:
		for _, param := range n.params {
			if err = Walk(visitor, param); err != nil {
				return
			}
		}
		if err = Walk(visitor, n.block); err != nil {
			return
		}
	case *ParamNode:
		if err = Walk(visitor, n.varNode); err != nil {
			return
		}
		if err = Walk(visitor, n.typNode); err != nil {
			return
		}
	case *CompoundNode:
		for _, child := range n.children {
			if err = Walk(visitor, child); err != nil {
				return
			}
		}
	case *ProcedureCallNode:
		if err = Walk(visitor, n.procSymbol.blockNode); err != nil {
			return
		}
	case *AssignNode:
		if err = Walk(visitor, n.left); err != nil {
			return
		}
		if err = Walk(visitor, n.right); err != nil {
			return
		}
	case *NoopNode:
		// do nothing
	}

	return visitor.After(node)
}
