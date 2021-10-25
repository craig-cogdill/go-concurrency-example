package threadworker 

import (
	"github.com/vrecan/life"
	// "github.com/grindlemire/log"
)

type ThreadWorker interface {
	Start()
	Close()
}

type threadworker struct {
	*life.Life
}

func New() *threadworker {
	worker := &threadworker{
		Life: life.NewLife(),
	}
	worker.SetRun(worker.run)
	return worker 
}

func (w *threadworker) run() {
	for {
		select {
		case <-w.Done:
			return
		}
	}
}
