package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	VERSION = "v1.0.0"
)

var (

)

func info(output string) {
	log.Println(time.Now().Format("2006-01-02 15:04:05") + " [INFO] " + output)
}

func warn(output string) {
	log.Println(time.Now().Format("2006-01-02 15:04:05") + " [WARN] " + output)
}

func main() {
	log.Println("Starting smoothie-runner...")

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		info("Received host " + sig.String())
		done <- true
	}()

	go listenInput()

	<-done // listen for sigint and sigterm
	shutdown()
}

func shutdown() {
	info("Shutting down smoothie-runner...")
	os.Exit(0)
}