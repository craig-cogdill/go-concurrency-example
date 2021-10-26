package threadworker

import (
	"fmt"
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
	id       int
}

func New(b *broadcast.Broadcaster, threadId int) ThreadWorker {
	worker := threadworker{
		Life:     life.NewLife(),
		listener: b.Listen(),
		id:       threadId,
	}
	worker.SetRun(worker.run)
	return worker
}

func (w *threadworker) calculateHash() {
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString(uuid.NewString())
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(sb.String()), bcrypt.DefaultCost)
	if err != nil {
		log.Error("There was a problem generating a hash")
	}
	log.Debugf("Worker %d: %s", w.id, hash)
}

func (w *threadworker) run() {
	defer w.listener.Discard()
	for {
		select {
		case <-w.Done:
			return
		case ping := <-w.listener.Channel():
			// Convert the message received into a WaitGroup for signaling that this thread is done
			wg, ok := ping.(*sync.WaitGroup)
			if !ok {
				log.Errorf("Worker %d: Unable to convert channel message to waitgroup", w.id)
			}
			w.calculateHash()
			log.Debug(fmt.Sprintf("Worker %d reporting finished", w.id))
			wg.Done()
		}
	}
}
