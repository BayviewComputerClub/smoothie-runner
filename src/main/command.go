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

func listenInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		input , err := reader.ReadString('\n')
		if err == io.EOF {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if err != nil {
			println(err.Error())
		}

		input = strings.TrimRight(input,  "\n")
		cFound := false
		if strings.Split(input, " ")[0] == "ec" {
			for k, v := range commands {
				if k == strings.Split(input, " ")[1] {
					in := ""
					for i, str := range strings.Split(input, " ") {
						for i != 0 && i != 1 {
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
}

func substring(s string, start int, end int) string {
	start_str_idx := 0
	i := 0
	for j := range s {
		if i == start {
			start_str_idx = j
		}
		if i == end {
			return s[start_str_idx:j]
		}
		i++
	}
	return s[start_str_idx:]
}