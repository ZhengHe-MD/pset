package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	tests := map[string]struct {
		givenCode    ErrorCode
		givenModule  Module
		givenMessage string
		wantString   string
	}{
		"without module": {
			givenCode:    ErrorCodeIdNotFound,
			givenModule:  0,
			givenMessage: "hello world",
			wantString:   `<Error: code=IdNotFound,message="hello world">`,
		},
		"without code": {
			givenCode:    0,
			givenModule:  ModuleLexer,
			givenMessage: "hello world",
			wantString:   `<Error: module=Lexer,message="hello world">`,
		},
		"without message": {
			givenCode:    ErrorCodeDuplicateId,
			givenModule:  ModuleParser,
			givenMessage: "",
			wantString:   "<Error: module=Parser,code=DuplicateId>",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			e := Error{
				Code:    tc.givenCode,
				Module:  tc.givenModule,
				Message: tc.givenMessage,
			}
			assert.Equal(t, tc.wantString, e.Error())
		})
	}
}
