package server 

import (
	"github.com/vrecan/life"
	"github.com/grindlemire/log"
)

type Server interface {
	Start()
	Close()
}

type server struct {
	*life.Life
}

func New() *server {
	s := &server{
		Life:           life.NewLife(),
	}
	s.SetRun(s.run)
	return s 
}

func (s *server) run() {
    log.Info("Starting server...")

	for {
		select {
		case <-s.Done:
            log.Info("Shutting down server...")
			return
		}
	}
}
