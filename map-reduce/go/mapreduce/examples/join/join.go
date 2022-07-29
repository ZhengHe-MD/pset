package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mr"
)

var _ mr.MapReducer = Join{}

type Join struct {
	On map[string]string // specify what fields to join on, map<table, field>
}

func (join Join) Map(data []byte) (kvs []mr.KeyValue, err error) {
	bytLines := bytes.Split(data, []byte{'\n'})
	if len(bytLines) <= 2 {
		err = errors.New("empty table")
		return
	}

	table := string(bytLines[0])
	on, ok := join.On[table]
	if !ok {
		err = errors.New(
			fmt.Sprintf("join column of table %s not found", table))
		return
	}

	onIdx := -1
	var colNames []string
	for i, bytCol := range bytes.Split(bytLines[1], []byte{','}) {
		if string(bytCol) == on {
			onIdx = i
		}
		colNames = append(colNames, fmt.Sprintf("%s.%s", table, string(bytCol)))
	}

	if onIdx < 0 {
		err = errors.New(
			fmt.Sprintf("join column %s is not in table %s", on, table))
		return
	}

	var byt []byte
	for _, bytLine := range bytLines[2:] {
		kv := mr.KeyValue{}
		cols := make(map[string]string, len(colNames))
		cols["_table"] = table
		for j, bytCol := range bytes.Split(bytLine, []byte{','}) {
			if j == onIdx {
				kv.Key = string(bytCol)
				continue
			}
			cols[colNames[j]] = string(bytCol)
		}
		byt, err = json.Marshal(cols)
		if err != nil {
			return
		}
		kv.Value = string(byt)
		kvs = append(kvs, kv)
	}

	return
}

func (join Join) Reduce(key string, values []string) (v string, err error) {
	rows := make(map[string][]map[string]string)
	for _, value := range values {
		kv := make(map[string]string)
		if err = json.Unmarshal([]byte(value), &kv); err != nil {
			return
		}
		tableName := kv["_table"]
		delete(kv, "_table")
		rows[tableName] = append(rows[tableName], kv)
	}

	var tableNames []string
	for tableName, _ := range rows {
		tableNames = append(tableNames, tableName)
	}
	if len(tableNames) != 2 {
		err = errors.New("only join between two rows is allowed")
		return
	}

	rowsA, rowsB := rows[tableNames[0]], rows[tableNames[1]]
	var jRows []map[string]string
	for _, rowA := range rowsA {
		for _, rowB := range rowsB {
			jRow := rowA
			for colName, colValue := range rowB {
				jRow[colName] = colValue
			}
			jRows = append(jRows, jRow)
		}
	}

	byt, err := json.Marshal(jRows)
	if err != nil {
		return
	}

	v = string(byt)
	return
}

func main() {
	job := mr.Job{
		Id:        "join",
		InputDir:  "./mixtures/input",
		OutputDir: "./mixtures/output",
		Processor: Join{
			On: map[string]string{
				"students":    "id",
				"enrollments": "student_id",
			},
		},
		R: 3,
	}

	if err := job.Sequential(); err != nil {
		log.Panic(err)
	}

	log.Println("done")
}
