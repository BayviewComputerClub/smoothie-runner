package main

import (
	"github.com/BayviewComputerClub/smoothie-runner/judging"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	VERSION = "v0.9-testing"
)

func getEnv(key string, def string) string {
	e := os.Getenv(key)
	if e == "" {
		return def
	}
	return e
}

func init() {
	// init commands
	commands["help"] = commandHelp
	commands["version"] = commandVersion

	// environment variables
	var err error
	shared.PORT, err = strconv.Atoi(getEnv("PORT", "6821"))
	if err != nil {
		panic(err)
	}
	shared.TESTING_DIR = getEnv("TESTING_DIR", ".")
	shared.MAX_THREADS, err = strconv.Atoi(getEnv("MAX_THREADS", "8"))
	if err != nil {
		panic(err)
	}
	shared.DEBUG = getEnv("DEBUG", "false") == "true"
	shared.SANDBOX = getEnv("SANDBOX", "true") == "true"
}

// entry point
func main() {
	log.Printf("Starting smoothie-runner %s...", VERSION)

	// sigint, sigterm listener
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		util.Info("Received host " + sig.String())
		done <- true
	}()

	// start judge workers
	for i := 0; i < shared.MAX_THREADS; i++ {
		go judging.StartQueueWorker(i+1)
	}

	// start grpc
	startApiServer()
	go listenInput()

	<-done // wait for sigint and sigterm
	shutdown()
}

func shutdown() {
	util.Info("Shutting down smoothie-runner...")
	os.Exit(0)
}