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

const port = "8181"

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
	log.Debug("Notifying workers of new API request")
	var ping sync.WaitGroup
	ping.Add(s.numSubscribers)
	s.broadcast.Send(&ping)
	ping.Wait()
	log.Debug("Workers finished with request")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

func (s *server) run() {
    log.Debugf("Starting server on port %d...", port)
    httpServer := http.Server{
		Addr:    (":"+port),
		Handler: s.router,
	}

    go launchHttpServer(&httpServer)

	for range(s.Done) {
		log.Debug("Shutting down server...")
		shutdownHttpServer(&httpServer)
		return
	}
}
