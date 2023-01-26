package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActivationRecord_String(t *testing.T) {
	tests := map[string]struct {
		givenActivationRecord *ActivationRecord
		wantString            string
	}{
		"empty call stack": {
			givenActivationRecord: &ActivationRecord{
				Name:         "Main",
				Kind:         ARKindProgram,
				NestingLevel: 1,
				Members:      nil,
			},
			wantString: `1: PROGRAM Main`,
		},
		"has members": {
			givenActivationRecord: &ActivationRecord{
				Name:         "Main",
				Kind:         ARKindProgram,
				NestingLevel: 1,
				Members:      map[string]interface{}{"y": 7},
			},
			wantString: `
				1: PROGRAM Main
					y                   : 7
				`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, strings.Fields(tc.wantString), strings.Fields(tc.givenActivationRecord.String()))
		})
	}
}

func TestCallStack_String(t *testing.T) {
	callStack := new(CallStack)
	callStack.Push(&ActivationRecord{
		Name:         "Main",
		Kind:         ARKindProgram,
		NestingLevel: 1,
		Members:      map[string]interface{}{"y": 7},
	})

	wantString := `
		CALL STACK
		1: PROGRAM Main
			y                   : 7
	`
	assert.Equal(t, strings.Fields(wantString), strings.Fields(callStack.String()))
}
