package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	VERSION = "v1.0.0"
)

var (
	PORT int

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
	PORT, err = strconv.Atoi(getEnv("PORT", "6821"))
	if err != nil {
		panic(err)
	}
}

func info(output string) {
	println(time.Now().Format("2006-01-02 15:04:05") + " [INFO] " + output)
}

func warn(output string) {
	println(time.Now().Format("2006-01-02 15:04:05") + " [WARN] " + output)
}

func fatal(output string) {
	log.Fatal(output)
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

	startApiServer()
	go listenInput()

	<-done // listen for sigint and sigterm
	shutdown()
}

func shutdown() {
	info("Shutting down smoothie-runner...")
	os.Exit(0)
}