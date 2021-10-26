package server

import (
	"context"
	"net/http"
	"time"
	"sync"

	"github.com/gorilla/mux"
	"github.com/grindlemire/log"
	"github.com/vrecan/life"
    "github.com/textileio/go-threads/broadcast"
)

type Server interface {
	Start()
	Close()
}

// This function is blocking and should only be called as a goroutine
func launchHttpServer(server *http.Server) {
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe() Error: %v", err)
	}
}

func shutdownHttpServer(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

type server struct {
	*life.Life
	router *mux.Router
	broadcast *broadcast.Broadcaster
	numSubscribers int
}

func New(b *broadcast.Broadcaster, listeners int) *server {
	s := &server{
		Life:   life.NewLife(),
        router: mux.NewRouter(),
        broadcast: b,
        numSubscribers: listeners, 
	}

    s.router.HandleFunc("/hash", s.runHashWithThreadWorkers)
	s.SetRun(s.run)
	return s 
}

func (s *server) runHashWithThreadWorkers(w http.ResponseWriter, request *http.Request) {

    var ping sync.WaitGroup
    ping.Add(s.numSubscribers)
    s.broadcast.Send(&ping) // signal the workers to do their thing
    // Wait for all threads to complete their task
    ping.Wait()

    // Respond
    w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
    w.Write([]byte("success"))
}

func (s *server) run() {
    log.Info("Starting server...")
    httpServer := http.Server{
		Addr:    ":8181",
		Handler: s.router,
	}

    go launchHttpServer(&httpServer)

    // wait for a shutdown signal
	for {
		select {
		case <-s.Done:
            log.Info("Shutting down server...")
            shutdownHttpServer(&httpServer)
			return
		}
	}
}
