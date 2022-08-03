package mr

import (
	"log"
	"net/rpc"
	"testing"
	"time"
)

func _wait(t *testing.T, client *rpc.Client, operation *Operation) {
	log.Printf("job submitted, operation id %s\n", operation.Id)
	if operation.Done {
		t.Fatal("operation.Done should be false right after SubmitJob")
	}

	var done bool
	for !done {
		op := new(Operation)
		err := client.Call("Master.GetOperation", &GetOperationArgs{Id: operation.Id}, op)
		if err != nil {
			t.Fatal(err)
		}
		log.Println(op)
		done = op.Done
		time.Sleep(1 * time.Second)
	}
	return
}

func TestExamples(t *testing.T) {
	cluster := NewCluster(2)
	if err := cluster.Start(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := cluster.Shutdown(); err != nil {
			t.Fatal(err)
		}
	}()

	// meat
	endpoint, err := cluster.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	client, err := rpc.Dial("unix", endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	t.Run("word count (sequential)", func(t *testing.T) {
		operation := new(Operation)
		err = client.Call("Master.SubmitJob", &SubmitArgs{
			Job: &Job{
				InputDir:      "./mixtures/wc/input",
				OutputDir:     "./mixtures/wc/output",
				ProcessorName: "wc",
				R:             2,
			},
			Distributed: false,
		}, operation)
		if err != nil {
			t.Fatal(err)
		}
		_wait(t, client, operation)
	})

	t.Run("word count", func(t *testing.T) {
		operation := new(Operation)
		err = client.Call("Master.SubmitJob", &SubmitArgs{
			Job: &Job{
				InputDir:      "./mixtures/wc/input",
				OutputDir:     "./mixtures/wc/output",
				ProcessorName: "wc",
				R:             2,
			},
			Distributed: true,
		}, operation)
		if err != nil {
			t.Fatal(err)
		}
		_wait(t, client, operation)
	})

	t.Run("average", func(t *testing.T) {
		operation := new(Operation)
		err = client.Call("Master.SubmitJob", &SubmitArgs{
			Job: &Job{
				InputDir:      "./mixtures/avg/input",
				OutputDir:     "./mixtures/avg/output",
				ProcessorName: "avg",
				R:             2,
			},
			Distributed: true,
		}, operation)
		if err != nil {
			t.Fatal(err)
		}
		_wait(t, client, operation)
	})

	t.Run("join", func(t *testing.T) {
		operation := new(Operation)
		err = client.Call("Master.SubmitJob", &SubmitArgs{
			Job: &Job{
				InputDir:      "./mixtures/join/input",
				OutputDir:     "./mixtures/join/output",
				ProcessorName: "join",
				R:             2,
			},
			Distributed: true,
		}, operation)
		if err != nil {
			t.Fatal(err)
		}
		_wait(t, client, operation)
	})
}
