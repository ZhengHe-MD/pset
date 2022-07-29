package main

import (
	"bytes"
	"log"
	"mr"
	"strconv"
)

var _ mr.MapReducer = WordCount{}

type WordCount struct{}

func (wc WordCount) Map(data []byte) (kvs []mr.KeyValue, err error) {
	for _, byt := range bytes.Fields(data) {
		kvs = append(kvs, mr.KeyValue{
			Key:   string(byt),
			Value: "1",
		})
	}
	return
}

func (wc WordCount) Reduce(key string, values []string) (string, error) {
	return strconv.Itoa(len(values)), nil
}

func main() {
	job := mr.Job{
		Id:        "wc",
		InputDir:  "./mixtures/input",
		OutputDir: "./mixtures/output",
		Processor: WordCount{},
		R:         3,
	}

	if err := job.Sequential(); err != nil {
		log.Panic(err)
	}

	log.Println("done")
}
