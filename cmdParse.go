package main

import (
	"encoding/json"
	"fmt"
	"goSnakeServ/cwlog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func commandParser(input string, w http.ResponseWriter) bool {

	/* Before ID check */
	if input == "init" {
		id := makeUID()
		newPlayer := playerData{Name: genName(), ID: id, lastActive: time.Now()}
		players[id] = &newPlayer
		cwlog.DoLog(true, "Created player %v (%v).", newPlayer.Name, newPlayer.ID)

		b, err := json.Marshal(newPlayer)
		if err != nil {
			cwlog.DoLog(true, "commandParser: init: err: %v", err)
			return false
		}

		return writeByte(w, b)
	}

	cmdPart := strings.Split(input, ":")

	cwlog.DoLog(true, "%v: %v: '%v'", cmdPart[0], cmdPart[1], cmdPart[2])
	useridstr, command, data := cmdPart[0], cmdPart[1], cmdPart[2]
	userid, _ := strconv.ParseUint(useridstr, 10, 64)

	if len(cmdPart) != 3 {
		cwlog.DoLog(true, "Malformed request: %v", input)
		return false
	}

	/* Find player, if invalid exit*/
	player := players[userid]
	if player == nil {
		cwlog.DoLog(true, "Invalid userid: %v", useridstr)
		return false
	}

	if command == "ping" {
		cwlog.DoLog(true, "Client: %v (PING)", player.ID)
		playerActivity(player)
		return writeByte(w, []byte("PONG"))
	} else if command == "list" { /* List lobbies */
		b, _ := json.Marshal(lobbyList)
		playerActivity(player)
		return writeByte(w, b)

	} else if command == "join" { /* Join a lobby */
		inputID, err := strconv.ParseUint(data, 10, 64)
		if err != nil {
			cwlog.DoLog(true, "commandParser: Join: ParseUint: Error: %v", err)
			return false
		}
		if player.inLobby != nil {
			cwlog.DoLog(true, "commandParser: Join: player %v already in a lobby: %v,", player.ID, player.inLobby.ID)
			return false
		}
		for l, lobby := range lobbyList {
			if lobby.ID == inputID {
				lobby.Players = append(lobby.Players, player)
				player.inLobby = lobbyList[l]
				cwlog.DoLog(true, "Player: %v joined lobby: %v", player.ID, inputID)
				playerActivity(player)
				return writeTo(w, "joined", "%v", inputID)
			}
		}
		cwlog.DoLog(true, "Could not find lobby: %v for player: %v", inputID, player.ID)
		return false

	} else if command == "name" { /* Change player name */
		newName := filterName(data)
		if playerNameUnique(newName) {
			cwlog.DoLog(true, "Changed player '%v' (%v) name to '%v'", player.Name, player.ID, newName)
			player.Name = newName
			playerActivity(player)
		} else {
			cwlog.DoLog(true, "Player (%v) tried to rename to a non-unique name: '%v'", player.ID, newName)
		}
		return writeTo(w, "name", "%v", player.Name)
	} else if command == "createLobby" {
		newLobby := makePersonalLobby(player)
		if newLobby != nil {
			playerActivity(player)
			return writeTo(w, "createdLobby", "%v", newLobby.ID)
		}
		return false
	} else {
		cwlog.DoLog(true, "Unknown Command.")
		return false
	}
}

func writeByte(w http.ResponseWriter, input []byte) bool {
	_, err := w.Write(input)
	if err != nil {
		cwlog.DoLog(true, "Error writing response: %v", err)
		return false
	}
	return true
}

func writeByteTo(w http.ResponseWriter, command string, input []byte) bool {
	buf := []byte(command + ":")
	buf = append(buf[:], input[:]...)

	_, err := w.Write(buf)
	if err != nil {
		cwlog.DoLog(true, "Error writing response: %v", err)
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
