package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

func NewInterpreter() *Interpreter {
	return &Interpreter{
		callStack: new(CallStack),
	}
}

var _ Visitor = (*Interpreter)(nil)

type Interpreter struct {
	callStack *CallStack
}

func (it *Interpreter) Interpret(source string) (err error) {
	lexer := NewLexer(source)
	parser, err := NewParser(lexer)
	if err != nil {
		return
	}
	root, err := parser.Parse()
	if err != nil {
		return
	}
	if err = Walk(NewSemanticAnalyzer(), root); err != nil {
		return
	}
	return Walk(it, root)
}

func (it *Interpreter) Before(node ASTNode) (shouldStepIn bool, err error) {
	switch n := node.(type) {
	case *ProgramNode:
		ar := NewActivationRecord(n.name, ARKindProgram, 1)
		it.callStack.Push(ar)
		log.Debugf("ENTER: PROGRAM %s\n", n.name)
		log.Debugln(it.callStack)
	case *BlockNode:
	case *CompoundNode:
	case *VarDeclNode:
		return false, nil
	case *ProcedureDeclNode:
		return false, nil
	case *ProcedureCallNode:
		ar := NewActivationRecord(n.name, ARKindProcedure, n.procSymbol.scopeLevel+1)

		formalParams, actualParams := n.procSymbol.formalParams, n.actualParams
		if len(formalParams) != len(actualParams) {
			panic("len(formalParams) != len(actualParams)")
		}
		for i := 0; i < len(formalParams); i++ {
			ar.Set(formalParams[i].GetName(), it.expr(actualParams[i]))
		}

		it.callStack.Push(ar)
		log.Debugf("ENTER: PROCEDURE %s", n.name)
		log.Debugln(it.callStack)
	case *AssignNode:
		lhs := n.left
		rhs := it.expr(n.right)
		it.callStack.Peek().Set(lhs.value, rhs)
		return false, nil
	case *NoopNode:
	default:
		log.WithField("node", n).Panicln("unreachable", n)
	}
	return true, nil
}

func (it *Interpreter) After(node ASTNode) (err error) {
	switch n := node.(type) {
	case *ProgramNode:
		log.Debugf("LEAVE: PROGRAM %s", n.name)
		log.Debugln(it.callStack)
		it.callStack.Pop()
	case *ProcedureCallNode:
		log.Debugf("LEAVE: PROCEDURE %s", n.name)
		log.Debugln(it.callStack)
		it.callStack.Pop()
	}
	return
}

func (it *Interpreter) expr(node ASTNode) float64 {
	switch n := node.(type) {
	case *NumNode:
		if n.token.Kind == IntegerConst {
			return float64(n.intValue)
		} else if n.token.Kind == RealConst {
			return n.floatValue
		}
		log.Panicln("invalid token in NumNode")
	case *VarNode:
		if v, ok := it.callStack.Peek().Get(n.value); ok {
			return v.(float64)
		}
		log.Panicln("invalid symbol")
	case *UnaryOpNode:
		ret := it.expr(n.operand)
		if n.op == Minus {
			return -ret
		}
		return ret
	case *BinOpNode:
		lhs := it.expr(n.left)
		rhs := it.expr(n.right)
		switch n.op {
		case Plus:
			return lhs + rhs
		case Minus:
			return lhs - rhs
		case Mul:
			return lhs * rhs
		case IntegerDiv:
			return float64(int64(lhs) / int64(rhs))
		case FloatDiv:
			return lhs / rhs
		}
		log.Panicln("unreachable")
	}
	log.Panicln("unreachable")
	return 0
}

func main() {
	sourceFile := flag.String("f", "", "A Pascal source file")
	logLevel := flag.String("v", "INFO", "log level, debug, info, warn")

	flag.Parse()

	if *sourceFile == "" {
		log.Infoln("a source file path is required")
		return
	}

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.WithField("level", logLevel).Info("invalid log level")
		return
	}
	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{
		PadLevelText: true,
	})

	source, err := ioutil.ReadFile(*sourceFile)
	if err != nil {
		log.WithField("err", err.Error()).Info("read source file")
		return
	}

	interpreter := NewInterpreter()
	if err = interpreter.Interpret(string(source)); err != nil {
		log.Errorln(err)
	}
}
