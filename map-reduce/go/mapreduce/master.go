package mr

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"path"
	"strconv"
	"sync"
)

// Master is the concrete type for Master node described in the original paper.
type Master struct {
	Address string
	lsn     net.Listener

	mu      sync.Mutex             // protects the following fields.
	wi      int                    // index of current worker, used to implement round-robin strategy.
	workers []string               // registered worker addresses.
	clients map[string]*rpc.Client // map worker (address) to it's rpc client.
	jobs    map[string]*Job        // in-memory job store, which maps operation id to job.

	shutdown chan struct{}
}

// NewMaster creates an instance of Master node.
func NewMaster(address string) *Master {
	return &Master{
		Address:  address,
		clients:  make(map[string]*rpc.Client),
		jobs:     make(map[string]*Job),
		shutdown: make(chan struct{}),
	}
}

// RegisterArgs represents arguments passed when a worker node calls Register.
type RegisterArgs struct {
	// the communication endpoint of worker process,
	// such as IPC socket or Network socket
	Worker string
}

// Register is called when a Worker node wants to register itself to the Master node.
func (m *Master) Register(args *RegisterArgs, _ *struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// ignore all sanity checks.
	m.workers = append(m.workers, args.Worker)
	log.Println("new worker joined, existing workers ", m.workers)
	return nil
}

// SubmitArgs represents arguments passed when a client calls SubmitJob.
type SubmitArgs struct {
	Job         *Job // description of the job to submit
	Distributed bool // indicates whether the job should be scheduled distributively
}

// Operation is the reply from Master node when a client calls SubmitJob.
type Operation struct {
	Id    string
	Done  bool
	Error error
}

// SubmitJob is called when a client wants to submit a new job to Master node.
func (m *Master) SubmitJob(args *SubmitArgs, operation *Operation) error {
	job := args.Job
	job.Id = randString(8)
	operation.Id = randString(32)
	job.operation = operation

	m.mu.Lock()
	m.jobs[operation.Id] = job
	m.mu.Unlock()

	go func() {
		var err error
		if args.Distributed {
			err = m.distributed(args, operation)
		} else {
			err = m.sequential(args, operation)
		}
		if err != nil {
			operation.Error = err
		} else {
			operation.Done = true
		}
	}()

	return nil
}

// getClient is expected to choose a client of the worker with the least load.
func (m *Master) getClient() (client *rpc.Client, err error) {
	worker, err := m.getWorker()
	if err != nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	client, ok := m.clients[worker]
	if !ok {
		client, err = rpc.Dial("unix", worker)
		if err != nil {
			return
		}
		m.clients[worker] = client
	}
	return
}

// getWorker is expected to return the worker node with round-robin strategy.
func (m *Master) getWorker() (worker string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.workers) == 0 {
		err = errors.New("no registered worker nodes found")
		return
	}
	worker = m.workers[m.wi]
	m.wi = (m.wi + 1) % len(m.workers)
	return
}

// sequential runs the map/reduce job sequentially on an arbitrary Worker node.
func (m *Master) sequential(args *SubmitArgs, operation *Operation) (err error) {
	job := args.Job

	files, err := ioutil.ReadDir(job.InputDir)
	if err != nil {
		return
	}

	client, err := m.getClient()
	if err != nil {
		return
	}

	// map phase
	for i, file := range files {
		doTaskArgs := &DoTaskArgs{
			Type: TaskTypeMap,
			Map: &MapTask{
				Id:      strconv.Itoa(i),
				MapFile: path.Join(job.InputDir, file.Name()),
				Job:     job,
			},
		}

		if err = client.Call("Worker.DoTask", doTaskArgs, nil); err != nil {
			return
		}
	}

	// reduce phase
	for i := 0; i < job.R; i++ {
		doTaskArgs := &DoTaskArgs{
			Type: TaskTypeReduce,
			Reduce: &ReduceTask{
				Id:         strconv.Itoa(i),
				Job:        job,
				MapTaskNum: len(files),
			},
		}

		if err = client.Call("Worker.DoTask", doTaskArgs, nil); err != nil {
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

// distributed runs the map/reduce job on available Worker nodes in a distributed manner.
func (m *Master) distributed(args *SubmitArgs, operation *Operation) (err error) {
	job := args.Job

	files, err := ioutil.ReadDir(job.InputDir)
	if err != nil {
		return
	}

	// map phase
	var mwg sync.WaitGroup
	mwg.Add(len(files))
	for i, file := range files {
		doTaskArgs := &DoTaskArgs{
			Type: TaskTypeMap,
			Map: &MapTask{
				Id:      strconv.Itoa(i),
				MapFile: path.Join(job.InputDir, file.Name()),
				Job:     job,
			},
		}

		client, err := m.getClient()
		if err != nil {
			return err
		}

		go func() {
			err := client.Call("Worker.DoTask", doTaskArgs, nil)
			if err != nil {
				operation.Error = err
			}
			mwg.Done()
		}()
	}
	mwg.Wait()
	log.Println("Map phase done.")

	// reduce phase
	var rwg sync.WaitGroup
	rwg.Add(job.R)
	for i := 0; i < job.R; i++ {
		doTaskArgs := &DoTaskArgs{
			Type: TaskTypeReduce,
			Reduce: &ReduceTask{
				Id:         strconv.Itoa(i),
				Job:        job,
				MapTaskNum: len(files),
			},
		}

		client, err := m.getClient()
		if err != nil {
			return err
		}

		go func() {
			err := client.Call("Worker.DoTask", doTaskArgs, nil)
			if err != nil {
				operation.Error = err
			}
			rwg.Done()
		}()
	}
	rwg.Wait()
	log.Println("Reduce phase done.")

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

// GetOperationArgs represents arguments passed when a client calls GetOperation.
type GetOperationArgs struct {
	Id string
}

// GetOperation requests the operation status of a map/reduce job.
func (m *Master) GetOperation(args *GetOperationArgs, operation *Operation) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, ok := m.jobs[args.Id]
	if !ok {
		return errors.New("operation not found")
	}
	// copy job.operation's content to operation
	operation.Id = job.operation.Id
	operation.Done = job.operation.Done
	operation.Error = job.operation.Error
	return nil
}

func (m *Master) serveRPC() (err error) {
	srv := rpc.NewServer()
	if err = srv.Register(m); err != nil {
		return
	}

	var lad *net.UnixAddr
	if lad, err = net.ResolveUnixAddr("unix", m.Address); err != nil {
		return
	}
	var lsn net.Listener
	if lsn, err = net.ListenUnix("unix", lad); err != nil {
		return
	}
	m.lsn = lsn

	go func() {
		log.Println("master node starts listening")
		var conn net.Conn
	rpcLoop:
		for {
			select {
			case <-m.shutdown:
				log.Println("break rpcLoop due to shutdown signal.")
				break rpcLoop
			default:
			}

			if conn, err = lsn.Accept(); err != nil {
				break
			}

			go srv.ServeConn(conn)
		}
		log.Println("master node exits")
	}()

	return
}

func (m *Master) shutdownRPC() (err error) {
	// close all client connections.
	m.mu.Lock()
	for worker, client := range m.clients {
		log.Printf("close the connection of worker %s\n", worker)
		client.Close()
	}
	m.mu.Unlock()

	close(m.shutdown)
	return m.lsn.Close()
}

func randString(size int) string {
	buf := make([]byte, size)
	rand.Read(buf)
	return fmt.Sprintf("%x", buf)
}
