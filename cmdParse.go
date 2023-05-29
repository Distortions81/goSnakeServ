package main

import (
	"goSnakeServ/cwlog"
	"net/http"
	"strings"
)

func commandParser(input string, w http.ResponseWriter) {

	/* Remove newlines */
	input = strings.ReplaceAll(input, "\n", "")
	input = strings.ReplaceAll(input, "\r", "")

	cmdPart := strings.Split(input, ":")
	pLen := len(cmdPart)
	if pLen < 2 {
		cwlog.DoLog(true, "Invalid number of arguments.")
		return
	}

	if cmdPart[0] == "test" { /* Check for updates */
		cwlog.DoLog(true, "test: '%v'", cmdPart[1])

		_, err := w.Write([]byte("test"))
		if err != nil {
			cwlog.DoLog(true, "Error writing response:", err)
			return
		}

		return
	} else {
		cwlog.DoLog(true, "Unknown Command.")
	}
}
