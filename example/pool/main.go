package main

import (
	"fmt"
	"time"
	"github.com/ioswarm/golik"
)

type Worker struct{

}

func (w *Worker) HandleString(s string, c golik.CloveRef) {
	time.Sleep(1 * time.Second)
	c.Info("Receive: %v", s)
}

func NewWorker() interface{} {
	return &Worker{}
}

// TODO explain in more detail
func main() {
	system := golik.GolikSystem()

	ref := system.Of(golik.WorkerPool(10, golik.CloveDefinition {
		Name: "worker",
		Receiver: golik.Minion(&Worker{}),
	}))

	for i := 0; i < 100; i++ {
		//time.Sleep(250 * time.Millisecond)
		ref.Tell(fmt.Sprintf("STRING-VALUE -> %v", i))
	}

	<-system.Terminated()
}