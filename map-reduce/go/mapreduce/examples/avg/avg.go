package main

import (
	"bytes"
	"log"
	"mr"
	"strconv"
)

var _ mr.MapReducer = Avg{}

type Avg struct{}

func (a Avg) Map(data []byte) (kvs []mr.KeyValue, err error) {
	var cnt, sum, num int
	for _, byt := range bytes.Fields(data) {
		num, err = strconv.Atoi(string(byt))
		if err != nil {
			return
		}
		sum += num
		cnt += 1
	}

	kvs = append(kvs, mr.KeyValue{
		Key:   "sum",
		Value: strconv.Itoa(sum),
	}, mr.KeyValue{
		Key:   "cnt",
		Value: strconv.Itoa(cnt),
	})
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

func main() {
	job := mr.Job{
		Id:        "avg",
		InputDir:  "./mixtures/input",
		OutputDir: "./mixtures/output",
		Processor: Avg{},
		R:         1,
	}

	if err := job.Sequential(); err != nil {
		log.Panic(err)
	}

	log.Println("done")
}
