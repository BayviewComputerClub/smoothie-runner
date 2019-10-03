package main

import (
	"bufio"
	"io"
	"os"
	"strings"
	"time"
)

var (
	commands = make(map[string]interface{})
)

func init() {
	commands["help"] = commandHelp
	commands["version"] = commandVersion

}

func commandHelp(input string) {
	println("----- Help -----")
	println("version | Get the smoothie-runner version.")
}

func commandVersion(input string) {
	println("Version: " + VERSION)
}

func listenInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err == io.EOF {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if err != nil {
			println(err.Error())
		}

		input = strings.TrimRight(input, "\n")
		cFound := false
		for k, v := range commands {
			if k == strings.Split(input, " ")[0] {
				in := ""

				for i, str := range strings.Split(input, " ") {
					for i != 0 {
						in += str
					}
				}

				v.(func(string))(in)
				cFound = true
				break
			}
		}
		if !cFound {
			println("Unknown command.")
		}

	}
}
