package main

import (
	"encoding/json"
	"goSnakeServ/cwlog"
	"net/http"
	"strings"
	"time"

	"golang.org/x/exp/rand"
)

func init() {
	lobbyList = []lobbyData{
		{
			Name:    "Test",
			ID:      uint64(rand.Int63()),
			Ticks:   323,
			Started: time.Now(),
			Level:   1,
		},
	}
}

func commandParser(input string, w http.ResponseWriter) {

	/* Remove newlines */
	input = strings.TrimSuffix(input, "\n")

	cmdPart := strings.Split(input, ":")
	pLen := len(cmdPart)
	if pLen < 2 {
		cwlog.DoLog(true, "Invalid number of arguments.")
		return
	}

	if cmdPart[0] == "list" { /* Check for updates */
		cwlog.DoLog(true, "list: '%v'", cmdPart[1])

		b, _ := json.Marshal(lobbyList)
		writeByteTo(w, "list", b)

		return
	} else {
		cwlog.DoLog(true, "Unknown Command.")
	}
}

func writeByteTo(w http.ResponseWriter, command string, input []byte) {
	buf := []byte(command + ":")
	buf = append(buf[:], input[:]...)

	_, err := w.Write(buf)
	if err != nil {
		cwlog.DoLog(true, "Error writing response:", err)
		return
	}
}

func writeStringTo(w http.ResponseWriter, command string, input string) {
	writeByteTo(w, command, []byte(input))
}
