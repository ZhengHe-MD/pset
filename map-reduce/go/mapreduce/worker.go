package mr

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
)

// Worker is the concrete type for Worker node described in the original paper.
type Worker struct {
	address string
	master  string

	lsn      net.Listener
	shutdown chan struct{}
}

// NewWorker creates an instance of Worker node.
func NewWorker(address string, master string) *Worker {
	return &Worker{
		address:  address,
		master:   master,
		shutdown: make(chan struct{}),
	}
}

// DoTaskArgs represents arguments passed when the DoTask method of a worker node is called.
type DoTaskArgs struct {
	Type   TaskType
	Map    *MapTask
	Reduce *ReduceTask
}

// TaskType denotes the type of task, only map or reduce is allowed.
type TaskType int32

const (
	_ TaskType = iota
	TaskTypeMap
	TaskTypeReduce
)

func (tt TaskType) String() string {
	switch tt {
	case TaskTypeMap:
		return "Map"
	case TaskTypeReduce:
		return "Reduce"
	default:
		return "Unknown"
	}
}

// DoTask execute the given map/reduce task synchronously.
func (w *Worker) DoTask(args *DoTaskArgs, _ *struct{}) error {
	switch args.Type {
	case TaskTypeMap:
		log.Printf("worker %s start doing %s task %s\n", w.address, args.Type, args.Map.Id)
		return args.Map.DoMap()
	case TaskTypeReduce:
		log.Printf("worker %s start doing %s task %s\n", w.address, args.Type, args.Reduce.Id)
		return args.Reduce.DoReduce()
	default:
		return errors.New(fmt.Sprintf("unsupported task type %d", args.Type))
	}
}

func (w *Worker) register() error {
	client, err := rpc.Dial("unix", w.master)
	if err != nil {
		return err
	}

	return client.Call("Master.Register", &RegisterArgs{
		Worker: w.address,
	}, nil)
}

func (w *Worker) serveRPC() (err error) {
	srv := rpc.NewServer()
	if err = srv.Register(w); err != nil {
		return
	}

	var lad *net.UnixAddr
	if lad, err = net.ResolveUnixAddr("unix", w.address); err != nil {
		return
	}
	var lsn net.Listener
	if lsn, err = net.ListenUnix("unix", lad); err != nil {
		return
	}
	w.lsn = lsn

	go func() {
		log.Println("worker node starts listening")

		var conn net.Conn
	rpcLoop:
		for {
			select {
			case <-w.shutdown:
				log.Println("break rpcLoop due to shutdown signal.")
				break rpcLoop
			default:
			}

			if conn, err = lsn.Accept(); err != nil {
				break
			}

			go srv.ServeConn(conn)
		}
		log.Println("worker node exits")
	}()

	return w.register()
}

func (w *Worker) shutdownRPC() (err error) {
	close(w.shutdown)
	return w.lsn.Close()
}
