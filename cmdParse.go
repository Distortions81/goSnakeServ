package main

import (
	"encoding/json"
	"fmt"
	"goSnakeServ/cwlog"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	lobbyList = []lobbyData{
		{
			Name: "Welcome",
		},
	}
}

func commandParser(input string, w http.ResponseWriter) bool {

	/* Remove newlines */
	input = strings.TrimSuffix(input, "\n")

	cmdPart := strings.Split(input, ":")
	if len(cmdPart) != 3 {
		return false
	}
	cwlog.DoLog(true, "%v: %v: '%v'", cmdPart[0], cmdPart[1], cmdPart[2])
	useridstr, command, data := cmdPart[0], cmdPart[1], cmdPart[2]
	userid, _ := strconv.ParseUint(useridstr, 10, 64)

	/* Before ID check */
	if command == "init" {
		id := makeUID()
		newPlayer := &playerData{Name: genName(), ID: id}
		players[id] = newPlayer
		cwlog.DoLog(true, "Created player %v (%v).", newPlayer.Name, newPlayer.ID)
		return writeTo(w, "init", "%v", id)
	}

	/* Find player, if invalid exit*/
	player := players[userid]
	if player == nil {
		cwlog.DoLog(true, "Invalid userid: %v", useridstr)
		return false
	}

	if command == "list" { /* List lobbies */
		b, _ := json.Marshal(lobbyList)
		return writeByteTo(w, "list", b)

	} else if command == "join" { /* Join a lobby */
		inputID, err := strconv.ParseUint(data, 10, 64)
		if err != nil {
			cwlog.DoLog(true, "commandParser: Join: ParseUint: Error: %v", err)
			return false
		}
		if player.InLobby != nil {
			cwlog.DoLog(true, "commandParser: Join: player %v already in a lobby: %v,", player.ID, player.InLobby.ID)
			return false
		}
		for l, lobby := range lobbyList {
			if lobby.ID == inputID {
				lobby.Players = append(lobby.Players, player)
				player.InLobby = &lobbyList[l]
				cwlog.DoLog(true, "Player: %v joined lobby: %v", player.ID, inputID)
				return writeTo(w, "joined", "%v", inputID)
			}
		}
		cwlog.DoLog(true, "Could not find lobby: %v for player: %v", inputID, player.ID)
		return false

	} else if command == "name" { /* Change player name */
		if playerNameUnique(data) {
			cwlog.DoLog(true, "Changed player '%v' (%v) name to '%v'", player.Name, player.ID, data)
			player.Name = data
		}
		return writeTo(w, "name", "%v", player.Name)
	} else {
		cwlog.DoLog(true, "Unknown Command.")
		return false
	}
}

func writeByteTo(w http.ResponseWriter, command string, input []byte) bool {
	buf := []byte(command + ":")
	buf = append(buf[:], input[:]...)

	_, err := w.Write(buf)
	if err != nil {
		cwlog.DoLog(true, "Error writing response:", err)
		return false
	}

	cwlog.DoLog(true, "WroteTo %v:%v", command, string(input))
	return true
}

func writeStringTo(w http.ResponseWriter, command string, input string) bool {
	return writeByteTo(w, command, []byte(input))
}

func writeTo(w http.ResponseWriter, command string, inputFormat string, args ...interface{}) bool {
	input := fmt.Sprintf(inputFormat, args...)
	return writeStringTo(w, command, input)
}
