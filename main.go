package main

import (
	"fmt"
	"io"
	"os"
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

func configureLogging() {
	logConfig := log.Default
	logLevel := os.Getenv("LOG_LEVEL")
	if len(logLevel) == 0 {
		logLevel = string(log.Default.Level)
	}
	logConfig.Level = log.Level(logLevel)
	log.Init(logConfig)
}

func main() {
	configureLogging()
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM)
	goRoutines := []io.Closer{}

	// Report on the number of threads
	numGoroutines := calculateMaxThreads()
	log.Info(fmt.Sprintf("Spawning %d goroutines... ", numGoroutines))

	// Create the broadcaster for notifying the threadpool
	broadcast := broadcast.NewBroadcaster(1)

	// Create and start workers
	for i := 0; i < numGoroutines; i++ {
		worker := threadworker.New(broadcast, i)
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
