package mr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

var MapReducers = map[string]MapReducer{
	"wc":  WordCount{},
	"avg": Avg{},
	"select": Select{
		Where: map[string]string{"name": "Jill"},
	},
	"join": Join{
		// NOTE: for the sake of simplicity, hard-code the join columns here.
		On: map[string]string{
			"students":    "id",
			"enrollments": "student_id",
		},
	},
}

type WordCount struct{}

func (wc WordCount) Map(data []byte) (kvs []KeyValue, err error) {
	for _, byt := range bytes.Fields(data) {
		kvs = append(kvs, KeyValue{
			Key:   string(byt),
			Value: "1",
		})
	}
	return
}

func (wc WordCount) Reduce(key string, values []string) (string, error) {
	return strconv.Itoa(len(values)), nil
}

type Avg struct{}

func (a Avg) Map(data []byte) (kvs []KeyValue, err error) {
	var cnt, sum, num int
	for _, byt := range bytes.Fields(data) {
		num, err = strconv.Atoi(string(byt))
		if err != nil {
			return
		}
		sum += num
		cnt += 1
	}

	kvs = append(kvs,
		KeyValue{Key: "sum", Value: strconv.Itoa(sum)},
		KeyValue{Key: "cnt", Value: strconv.Itoa(cnt)})
	return
}

func (a Avg) Reduce(key string, values []string) (value string, err error) {
	var sum, num int
	for _, v := range values {
		num, err = strconv.Atoi(v)
		if err != nil {
			return
		}
		sum += num
	}
	value = strconv.Itoa(sum)
	return
}

type Select struct {
	// Where represents the conditions that selected rows must satisfy.
	// for simplicity, only one single condition with equals sign is
	// implicitly supported. like,
	// {
	//   "name": "Jill"
	// }
	// more than one conditions or ops other than equals sign are not supported
	// for the present, but it's not hard to implement that.
	Where map[string]string
}

func (s Select) Map(data []byte) (kvs []KeyValue, err error) {
	bytLines := bytes.Split(data, []byte{'\n'})
	var whereIdx int
	var colNames []string
	for i, bytCol := range bytes.Split(bytLines[0], []byte{','}) {
		if _, ok := s.Where[string(bytCol)]; ok {
			whereIdx = i
		}
		colNames = append(colNames, string(bytCol))
	}

	var byt []byte
	for _, bytLine := range bytLines[2:] {
		kv := KeyValue{}
		cols := make(map[string]string, len(colNames))
		for j, bytCol := range bytes.Split(bytLine, []byte{','}) {
			cols[colNames[j]] = string(bytCol)
		}
		if cols[colNames[whereIdx]] != s.Where[colNames[whereIdx]] {
			continue
		}
		kv.Key = cols[colNames[0]] // hard-code the index of id field for simplicity
		byt, err = json.Marshal(cols)
		if err != nil {
			return
		}
		kv.Value = string(byt)
		kvs = append(kvs, kv)
	}
	return
}

func (s Select) Reduce(key string, values []string) (value string, err error) {
	return values[0], nil
}

type Join struct {
	On map[string]string // specify what fields to join on, map<table, field>
}

func (join Join) Map(data []byte) (kvs []KeyValue, err error) {
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
		kv := KeyValue{}
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
