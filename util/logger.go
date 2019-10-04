package util

import (
	"log"
	"time"
)

func Info(output string) {
	println(time.Now().Format("2006-01-02 15:04:05") + " [INFO] " + output)
}

func Warn(output string) {
	println(time.Now().Format("2006-01-02 15:04:05") + " [WARN] " + output)
}

func Fatal(output string) {
	log.Fatal(output)
}