package main

import (
	"fmt"
	"io"
	"runtime"
	"syscall"

	"github.com/grindlemire/log"
	"github.com/textileio/go-threads/broadcast"
	"github.com/vrecan/death"

	"go-thread-model/pkg/server"
	"go-thread-model/pkg/threadworker"
)

const minimumThreads = 4

func calculateMaxThreads() int {
    availableCpus := runtime.NumCPU() - 1
    if availableCpus > minimumThreads {
        return availableCpus
    }
    return minimumThreads 
}

func main() {
	log.Init(log.Default)
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM)
    goRoutines := []io.Closer{}

    // Report on the number of threads
    numGoroutines := calculateMaxThreads()
    log.Info(fmt.Sprintf("Spawning %d goroutines... ", numGoroutines))

    // Create the broadcaster for notifying the threadpool
    broadcast := broadcast.NewBroadcaster(1)

    // Create and start workers
    for i := 0; i < numGoroutines; i++ {
        worker := threadworker.New(broadcast)
        goRoutines = append(goRoutines, worker)
        worker.Start()
    }

    // Start the HTTP server
    server := server.New(broadcast, numGoroutines)
    server.Start()
    goRoutines = append(goRoutines, server)

	err := d.WaitForDeath(goRoutines...)
	if err != nil {
		log.Fatalf("failed to cleanly shut down all go routines: %v", err)
	}
    broadcast.Discard()

	log.Info("successfully shutdown all go routines")
}
