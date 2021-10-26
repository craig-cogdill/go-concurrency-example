package threadworker

import (
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/grindlemire/log"
	"github.com/textileio/go-threads/broadcast"
	"github.com/vrecan/life"
)

type ThreadWorker interface {
	Start()
	Close() error
}

type threadworker struct {
	*life.Life
    listener *broadcast.Listener
}

func New(b *broadcast.Broadcaster) ThreadWorker {
	worker := threadworker{
		Life: life.NewLife(),
        listener: b.Listen(),
	}
	worker.SetRun(worker.run)
	return worker 
}

func (w *threadworker) calculateHash() {
    var sb strings.Builder
    for i := 0; i < 1000; i++ {
        sb.WriteString(uuid.NewString())
    }
    _, err := bcrypt.GenerateFromPassword([]byte(sb.String()), bcrypt.DefaultCost)
    if err != nil {
        log.Error("There was a problem generating a hash")
    }
}

func (w *threadworker) run() {
    defer w.listener.Discard()
	for {
		select {
		case <-w.Done:
			return
		case ping := <-w.listener.Channel():
            // Convert the message received into a WaitGroup for signalling that this thread is done
            wg, ok := ping.(*sync.WaitGroup)
            if !ok {
                log.Error("Unable to convert channel message to waitgroup... shit has hit the fan")
            }
            w.calculateHash()
            wg.Done() 
        default:
            continue
	    }
    }
}
