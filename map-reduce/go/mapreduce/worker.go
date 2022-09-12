package mr

import (
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

// DoMapTaskArgs represents arguments passed to Worker.DoMapTask.
type DoMapTaskArgs struct {
	MapTask *MapTask
}

// DoMapTask executes the given map task synchronously.
func (w *Worker) DoMapTask(args *DoMapTaskArgs, _ *struct{}) error {
	log.Printf("worker %s start doing map task %s\n", w.address, args.MapTask.Id)
	return args.MapTask.Do()
}

// DoReduceTaskArgs represents arguments passed to Worker.DoReduceTask.
type DoReduceTaskArgs struct {
	ReduceTask *ReduceTask
}

// DoReduceTask executes the given reduce task synchronously.
func (w *Worker) DoReduceTask(args *DoReduceTaskArgs, _ *struct{}) error {
	log.Printf("worker %s start doing reduce task %s\n", w.address, args.ReduceTask.Id)
	return args.ReduceTask.Do()
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
