package main

import (
	"fmt"
	"strings"
)

type ARKind int

const (
	ARKindProgram   ARKind = 1
	ARKindProcedure ARKind = 2
)

func (k ARKind) String() string {
	switch k {
	case ARKindProgram:
		return "PROGRAM"
	case ARKindProcedure:
		return "PROCEDURE"
	default:
		return "UNKNOWN"
	}
}

type CallStack struct {
	records []*ActivationRecord
}

func (cs *CallStack) Push(ar *ActivationRecord) {
	cs.records = append(cs.records, ar)
}

func (cs *CallStack) Pop() (ar *ActivationRecord) {
	ar = cs.records[len(cs.records)-1]
	cs.records = cs.records[:len(cs.records)-1]
	return
}

func (cs *CallStack) Peek() *ActivationRecord {
	return cs.records[len(cs.records)-1]
}

func (cs *CallStack) String() string {
	var records = make([]string, len(cs.records))
	for i, record := range cs.records {
		records[i] = record.String()
	}
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}
	sb := strings.Builder{}
	sb.WriteString("CALL STACK\n")
	sb.WriteString(strings.Join(records, "\n"))
	sb.WriteRune('\n')
	return sb.String()
}

func NewActivationRecord(name string, kind ARKind, nestingLevel int) *ActivationRecord {
	return &ActivationRecord{
		Name:         name,
		Kind:         kind,
		NestingLevel: nestingLevel,
		Members:      make(map[string]interface{}),
	}
}

type ActivationRecord struct {
	Name         string
	Kind         ARKind
	NestingLevel int
	Members      map[string]interface{}
}

func (cs *ActivationRecord) Get(key string) (value interface{}, ok bool) {
	value, ok = cs.Members[key]
	return
}

func (cs *ActivationRecord) Set(key string, value interface{}) {
	cs.Members[key] = value
}

func (cs *ActivationRecord) String() string {
	lines := []string{
		fmt.Sprintf("%d: %s %s",
			cs.NestingLevel, cs.Kind.String(), cs.Name),
	}

	for key, value := range cs.Members {
		lines = append(lines, fmt.Sprintf("\t%-20s: %v", key, value))
	}
	return strings.Join(lines, "\n")
}
