package mr

import (
	"net/rpc"
	"testing"
)

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
				Id:            randString(10),
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
	})

	t.Run("word count", func(t *testing.T) {
		operation := new(Operation)
		err = client.Call("Master.SubmitJob", &SubmitArgs{
			Job: &Job{
				Id:            randString(10),
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
	})

	t.Run("average", func(t *testing.T) {
		operation := new(Operation)
		err = client.Call("Master.SubmitJob", &SubmitArgs{
			Job: &Job{
				Id:            randString(10),
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
	})

	t.Run("join", func(t *testing.T) {
		operation := new(Operation)
		err = client.Call("Master.SubmitJob", &SubmitArgs{
			Job: &Job{
				Id:            randString(10),
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
	})
}
