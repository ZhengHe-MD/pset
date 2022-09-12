package mr

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
)

// Job keeps all related information including input parameters,
// map/reduce logic, runtime status, etc.
type Job struct {
	Id            string
	InputDir      string // the directory where input files reside
	OutputDir     string // the directory where output files reside
	ProcessorName string // the name of a MapReducer defined in mapreducers.go
	R             int    // number of reduce tasks

	operation *Operation // job status
}

// MapTask provides all the information needed to run a map task.
type MapTask struct {
	Id        string
	InputFile string // the input file to map phase.
	Job       *Job
}

// Do physically runs the map task.
func (mt *MapTask) Do() (err error) {
	mapReducer, ok := MapReducers[mt.Job.ProcessorName]
	if !ok {
		return errors.New("processor " + mt.Job.ProcessorName + " not found")
	}

	byt, err := ioutil.ReadFile(mt.InputFile)
	if err != nil {
		return
	}

	kvs, err := mapReducer.Map(byt)
	if err != nil {
		return
	}

	mappedFiles := make([]*os.File, 0, mt.Job.R)
	for i := 0; i < mt.Job.R; i++ {
		var mf *os.File
		mf, err = os.OpenFile(mappedFile(mt.Job.Id, mt.Id, strconv.Itoa(i)), os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return
		}
		defer func() {
			closeErr := mf.Close()
			if err == nil {
				err = closeErr
			}
		}()
		mappedFiles = append(mappedFiles, mf)
	}

	encoders := make([]Encoder, 0, mt.Job.R)
	for _, mf := range mappedFiles {
		encoders = append(encoders, json.NewEncoder(mf))
	}

	var hsh int
	for _, kv := range kvs {
		if hsh, err = hash(kv.Key); err != nil {
			return
		}

		if err = encoders[hsh%mt.Job.R].Encode(&kv); err != nil {
			return
		}
	}

	return
}

func hash(key string) (hsh int, err error) {
	h := fnv.New32a()
	if _, err = h.Write([]byte(key)); err != nil {
		return
	}
	hsh = int(h.Sum32() & 0x7fffffff)
	return
}

// ReduceTask provides all the information needed to run a reduce task.
type ReduceTask struct {
	Id  string
	M   int // number of map tasks
	Job *Job
}

// Do physically runs the reduce task.
func (rt *ReduceTask) Do() (err error) {
	mapReducer, ok := MapReducers[rt.Job.ProcessorName]
	if !ok {
		return errors.New("processor " + rt.Job.ProcessorName + " not found")
	}

	var kvs []KeyValue

	// shuffle
	var mf *os.File
	for i := 0; i < rt.M; i++ {
		mf, err = os.Open(mappedFile(rt.Job.Id, strconv.Itoa(i), rt.Id))
		if err != nil {
			return
		}

		var shard []KeyValue
		decoder := json.NewDecoder(mf)
		var kv KeyValue
		for {
			err = decoder.Decode(&kv)
			if err == io.EOF {
				break
			}
			shard = append(shard, kv)
		}

		kvs = append(kvs, shard...)
		if closeErr := mf.Close(); closeErr != nil {
			return closeErr
		}
	}

	// sort by key
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Key < kvs[j].Key
	})

	// reduce
	var rkvs []KeyValue
	var k, v string
	var vs []string
	var i int
	for i < len(kvs) {
		k, vs = kvs[i].Key, append(vs, kvs[i].Value)

		for i+1 < len(kvs) && kvs[i+1].Key == k {
			i += 1
			vs = append(vs, kvs[i].Value)
		}

		if v, err = mapReducer.Reduce(k, vs); err != nil {
			return
		}

		rkvs = append(rkvs, KeyValue{Key: k, Value: v})

		vs = vs[:0]
		i += 1
	}

	outputFile, err := os.OpenFile(
		path.Join(rt.Job.OutputDir, reducedFile(rt.Job.Id, rt.Id)),
		os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer func() {
		closeErr := outputFile.Close()
		if err == nil {
			err = closeErr
		}
	}()

	encoder := json.NewEncoder(outputFile)
	for _, kv := range rkvs {
		if err = encoder.Encode(&kv); err != nil {
			return
		}
	}
	return
}

// mappedFile constructs the name of the mapped file which a MapTask generates
// for the corresponding ReduceTask.
func mappedFile(jobId string, mapTask string, reduceTask string) string {
	return fmt.Sprintf("mrtmp.%s-%s-%s", jobId, mapTask, reduceTask)
}

// reducedFile constructs the name of the reduced file which a ReduceTask generates.
func reducedFile(jobId string, reduceTask string) string {
	return fmt.Sprintf("mr.reduce.%s-%s", jobId, reduceTask)
}
