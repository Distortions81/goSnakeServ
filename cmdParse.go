package main

import (
	"goSnakeServ/cwlog"
	"net/http"
	"strings"
)

func commandParser(input string, w http.ResponseWriter) {

	/* Remove newlines */
	input = strings.TrimSuffix(input, "\n")

	cmdPart := strings.Split(input, ":")
	pLen := len(cmdPart)
	if pLen < 2 {
		cwlog.DoLog(true, "Invalid number of arguments.")
		return
	}

	if cmdPart[0] == "Hello" { /* Check for updates */
		cwlog.DoLog(true, "hello: '%v'", cmdPart[1])

		_, err := w.Write([]byte("Greetings\n"))
		if err != nil {
			cwlog.DoLog(true, "Error writing response:", err)
			return
		}

		return
	} else {
		cwlog.DoLog(true, "Unknown Command.")
	}
}
