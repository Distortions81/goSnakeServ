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

func commandParser(input string, w http.ResponseWriter, player *playerData) bool {

	/* Remove newlines */
	input = strings.TrimSuffix(input, "\n")

	cmdPart := strings.Split(input, ":")
	cwlog.DoLog(true, "%v: '%v'", cmdPart[0], cmdPart[1])

	if cmdPart[0] == "list" { /* List lobbies */
		b, _ := json.Marshal(lobbyList)
		return writeByteTo(w, "list", b, player)

	} else if cmdPart[0] == "join" { /* Join a lobby */
		inputID, err := strconv.ParseUint(cmdPart[1], 10, 64)
		if err != nil {
			cwlog.DoLog(true, "commandParser: Join: Error: %v", err)
			return false
		}
		for _, lob := range lobbyList {
			if lob.ID == uint64(inputID) {
				lob.Players = append(lob.Players, player)
				b, _ := json.Marshal(lob.ID)
				return writeByteTo(w, "joined", b, player)
			}
		}

	} else if cmdPart[0] == "name" { /* Join a lobby */

	} else {
		cwlog.DoLog(true, "Unknown Command.")
		return false
	}

	return true
}

func writeByteTo(w http.ResponseWriter, command string, input []byte, player *playerData) bool {
	buf := []byte(command + ":")
	buf = append(buf[:], input[:]...)

	_, err := w.Write(buf)
	if err != nil {
		cwlog.DoLog(true, "Error writing response:", err)
		return false
	}

	cwlog.DoLog(true, "WroteTo(%v): %v:%v", player.ID, command, string(input))
	return true
}

func writeStringTo(w http.ResponseWriter, command string, input string, player *playerData) {
	writeByteTo(w, command, []byte(input), player)
}
