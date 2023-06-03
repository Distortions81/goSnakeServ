package main

import (
	"encoding/json"
	"goSnakeServ/cwlog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func init() {
	lobbyList = []lobbyData{
		{
			Name:  "Test",
			ID:    uint64(time.Now().UnixNano()),
			Ticks: 323,
			Level: 1,
		},
	}
}

func commandParser(input string, w http.ResponseWriter, player *playerData) {

	/* Remove newlines */
	input = strings.TrimSuffix(input, "\n")

	cmdPart := strings.Split(input, ":")
	pLen := len(cmdPart)
	if pLen < 2 {
		cwlog.DoLog(true, "Invalid number of arguments.")
		return
	}

	if cmdPart[0] == "init" {

	} else if cmdPart[0] == "list" { /* List lobbies */
		b, _ := json.Marshal(lobbyList)
		writeByteTo(w, "list", b)

	} else if cmdPart[0] == "join" { /* Join a lobby */
		inputID, err := strconv.ParseUint(cmdPart[1], 10, 64)
		if err != nil {
			cwlog.DoLog(true, "commandParser: Join: Error: %v", err)
			return
		}
		for _, lob := range lobbyList {

			if lob.ID == uint64(inputID) {
				lob.Players = append(lob.Players, player)
			}
		}
		b, _ := json.Marshal(lobbyList)
		writeByteTo(w, "list", b)

	} else if cmdPart[0] == "name" { /* Join a lobby */

	} else {
		cwlog.DoLog(true, "Unknown Command.")
		return
	}
	cwlog.DoLog(true, "%v: '%v'", cmdPart[0], cmdPart[1])
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
