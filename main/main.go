package main

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	VERSION = "v1.0.0"
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
}

func main() {
	log.Println("Starting smoothie-runner...")

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		util.Info("Received host " + sig.String())
		done <- true
	}()

	startApiServer()
	go listenInput()

	<-done // listen for sigint and sigterm
	shutdown()
}

func shutdown() {
	util.Info("Shutting down smoothie-runner...")
	os.Exit(0)
}