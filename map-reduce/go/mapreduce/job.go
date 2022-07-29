package mr

import (
	"encoding/json"
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
	Id        string
	InputDir  string     // the directory where input files reside
	OutputDir string     // the directory where output files reside
	Processor MapReducer // the real map/reduce logic user provides
	R         int        // number of reduce tasks
}

func (job *Job) Sequential() (err error) {
	files, err := ioutil.ReadDir(job.InputDir)
	if err != nil {
		return
	}

	// map phase
	for i, file := range files {
		task := &MapTask{
			Id:      strconv.Itoa(i),
			MapFile: path.Join(job.InputDir, file.Name()),
			Job:     job,
		}

		if err = task.DoMap(job.Processor); err != nil {
			return
		}
	}

	// reduce phase
	for i := 0; i < job.R; i++ {
		task := &ReduceTask{Id: strconv.Itoa(i), Job: job, MapTaskNum: len(files)}

		if err = task.DoReduce(job.Processor); err != nil {
			return
		}
	}

	// remove temporary files
	for i := 0; i < len(files); i++ {
		for j := 0; j < job.R; j++ {
			err = os.Remove(intermediateName(job.Id, strconv.Itoa(i), strconv.Itoa(j)))
			if err != nil {
				return
			}
		}
	}
	return
}

type MapTask struct {
	Id      string
	MapFile string // the input file to map phase.
	Job     *Job
}

func (mt *MapTask) DoMap(mapper Mapper) (err error) {
	byt, err := ioutil.ReadFile(mt.MapFile)
	if err != nil {
		return
	}

	kvs, err := mapper.Map(byt)
	if err != nil {
		return
	}

	var intermediates []*os.File
	for i := 0; i < mt.Job.R; i++ {
		var file *os.File
		file, err = os.OpenFile(intermediateName(mt.Job.Id, mt.Id, strconv.Itoa(i)), os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return
		}
		intermediates = append(intermediates, file)
	}

	var encoders []Encoder
	for _, intermediate := range intermediates {
		encoders = append(encoders, json.NewEncoder(intermediate))
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

	for _, intermediate := range intermediates {
		intermediate.Close()
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

type ReduceTask struct {
	Id         string
	Job        *Job
	MapTaskNum int
}

func (rt *ReduceTask) DoReduce(reducer Reducer) (err error) {
	var kvs []KeyValue

	// shuffle
	var intermediate *os.File
	for i := 0; i < rt.MapTaskNum; i++ {
		intermediate, err = os.Open(intermediateName(rt.Job.Id, strconv.Itoa(i), rt.Id))
		if err != nil {
			return
		}

		var shard []KeyValue
		decoder := json.NewDecoder(intermediate)
		var kv KeyValue
		for {
			err = decoder.Decode(&kv)
			if err == io.EOF {
				break
			}
			shard = append(shard, kv)
		}

		kvs = append(kvs, shard...)
		intermediate.Close()
	}

	// sort by key
	sort.Sort(ByKey(kvs))

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

		if v, err = reducer.Reduce(k, vs); err != nil {
			return
		}

		rkvs = append(rkvs, KeyValue{Key: k, Value: v})

		vs = vs[:0]
		i += 1
	}

	outputFile, err := os.OpenFile(
		path.Join(rt.Job.OutputDir, reduceName(rt.Job.Id, rt.Id)),
		os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	for _, kv := range rkvs {
		if err = encoder.Encode(&kv); err != nil {
			return
		}
	}
	return
}

// intermediateName constructs the name of the intermediate file which a MapTask
// produces for the corresponding ReduceTask.
func intermediateName(jobId string, mapTask string, reduceTask string) string {
	return fmt.Sprintf("mrtmp.%s-%s-%s", jobId, mapTask, reduceTask)
}

func reduceName(jobId string, reduceTask string) string {
	return fmt.Sprintf("mr.reduce.%s-%s", jobId, reduceTask)
}
