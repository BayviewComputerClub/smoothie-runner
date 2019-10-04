package util

import (
	"log"
	"time"
)

func info(output string) {
	println(time.Now().Format("2006-01-02 15:04:05") + " [INFO] " + output)
}

func warn(output string) {
	println(time.Now().Format("2006-01-02 15:04:05") + " [WARN] " + output)
}

func fatal(output string) {
	log.Fatal(output)
}