package mr

import (
	"errors"
	"log"
)

type Cluster struct {
	master  *Master
	workers []*Worker

	numWorker int
}

func NewCluster(numWorker int) *Cluster {
	return &Cluster{numWorker: numWorker}
}

func (c *Cluster) Endpoint() (string, error) {
	if c.master == nil {
		return "", errors.New("cluster is not running.")
	}
	return c.master.Address, nil
}

func (c *Cluster) Start() (err error) {
	masterAddr := "/tmp/mr-master-" + randString(10)
	c.master = NewMaster(masterAddr)
	if err = c.master.serveRPC(); err != nil {
		return
	}

	for i := 0; i < c.numWorker; i++ {
		w := NewWorker("/tmp/mr-worker-"+randString(10), masterAddr)
		if err = w.serveRPC(); err != nil {
			return
		}
		c.workers = append(c.workers, w)
	}
	log.Println("cluster is ready.")
	return
}

func (c *Cluster) Shutdown() (err error) {
	for _, worker := range c.workers {
		if err = worker.shutdownRPC(); err != nil {
			return
		}
	}

	if err = c.master.shutdownRPC(); err != nil {
		return
	}
	log.Println("cluster is shutdown.")
	return
}
