package main

import (
	"fmt"
	"strings"
)

type ErrorCode int32

const (
	ErrorCodeUnspecified ErrorCode = iota
	// Parser.
	ErrorCodeUnexpectedToken
	// SemanticAnalyzer.
	ErrorCodeIdNotFound
	ErrorCodeDuplicateId
	ErrorCodeUnknownDataType
	ErrorCodeArgumentsMismatch
	// Lexer.
	ErrorCodeUnknownRune
	ErrorCodeUnclosedComment
)

func (ec ErrorCode) String() string {
	switch ec {
	case ErrorCodeUnspecified:
		return "Unspecified"
	case ErrorCodeUnexpectedToken:
		return "UnexpectedToken"
	case ErrorCodeIdNotFound:
		return "IdNotFound"
	case ErrorCodeDuplicateId:
		return "DuplicateId"
	case ErrorCodeUnknownDataType:
		return "UnknownDataType"
	case ErrorCodeArgumentsMismatch:
		return "ArgumentsMismatch"
	case ErrorCodeUnknownRune:
		return "UnknownRune"
	case ErrorCodeUnclosedComment:
		return "UnclosedComment"
	default:
		return "Unknown"
	}
}

type Module int32

const (
	ModuleUnspecified Module = iota
	ModuleLexer
	ModuleParser
	ModuleSemanticAnalyzer
	ModuleInterpreter
)

func (m Module) String() string {
	switch m {
	case ModuleUnspecified:
		return "Unspecified"
	case ModuleLexer:
		return "Lexer"
	case ModuleParser:
		return "Parser"
	case ModuleSemanticAnalyzer:
		return "SemanticAnalyzer"
	case ModuleInterpreter:
		return "Interpreter"
	default:
		return "Unknown"
	}
}

var _ error = Error{}

type Error struct {
	Code    ErrorCode
	Module  Module
	Message string
}

func (e Error) Error() string {
	sb := strings.Builder{}
	sb.WriteString("<Error: ")
	var fields []string
	if e.Module != ModuleUnspecified {
		fields = append(fields, fmt.Sprintf("module=%s", e.Module))
	}
	if e.Code != ErrorCodeUnspecified {
		fields = append(fields, fmt.Sprintf("code=%s", e.Code))
	}
	if e.Message != "" {
		fields = append(fields, fmt.Sprintf("message=\"%s\"", e.Message))
	}
	sb.WriteString(strings.Join(fields, ","))
	sb.WriteRune('>')
	return sb.String()
}
