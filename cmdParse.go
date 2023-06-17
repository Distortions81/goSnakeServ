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

func commandParser(input string, w http.ResponseWriter) {

	/* Before ID check */
	if input == "init" {
		id := makeUID()
		newPlayer := playerData{Name: genName(), ID: id, lastActive: time.Now()}

		cwlog.DoLog(true, "Created player %v (%v).", newPlayer.Name, newPlayer.ID)

		b, err := json.Marshal(newPlayer)
		if err != nil {
			cwlog.DoLog(true, "commandParser: init: err: %v", err)
			return
		}

		pListLock.Lock()
		pList[id] = &newPlayer
		writeByte(w, b)
		pListLock.Unlock()

		return
	}

	cmdPart := strings.Split(input, ":")

	if len(cmdPart) != 3 {
		cwlog.DoLog(true, "Malformed request: %v", input)
		return
	}

	cwlog.DoLog(true, "%v: %v: '%v'", cmdPart[0], cmdPart[1], cmdPart[2])
	useridstr, command, data := cmdPart[0], cmdPart[1], cmdPart[2]
	userid, _ := strconv.ParseUint(useridstr, 10, 64)

	/* Find player, if invalid exit */
	pListLock.RLock()
	player := pList[userid]
	pListLock.RUnlock()

	if player == nil {
		cwlog.DoLog(true, "Invalid userid: %v", useridstr)
		return

		//Game mode
	} else if player.inLobby != nil {
		//game handlers
	}

	if command == "ping" { /* Keep alive, and check latency */
		cwlog.DoLog(true, "Client: %v (PING)", player.ID)
		playerActivity(player)
		writeByte(w, []byte("PONG"))
		return

	} else if command == "list" { /* List lobbies */
		b, _ := json.Marshal(lobbyList)
		playerActivity(player)
		writeByte(w, b)
		return

	} else if command == "join" { /* Join a lobby */
		inputID, err := strconv.ParseUint(data, 10, 64)
		if err != nil {
			cwlog.DoLog(true, "commandParser: Join: ParseUint: Error: %v", err)
			return
		}
		if player.inLobby != nil {
			cwlog.DoLog(true, "commandParser: Join: player %v already in a lobby: %v,", player.ID, player.inLobby.ID)
			return
		}
		for l, lobby := range lobbyList {
			if lobby.ID == inputID {
				player.Length = 1
				player.Tiles = []XY{{X: 1, Y: 1}}
				player.Direction = DIR_SOUTH
				lobby.Players = append(lobby.Players, player)
				player.inLobby = lobbyList[l]
				cwlog.DoLog(true, "Player: %v joined lobby: %v", player.ID, inputID)
				playerActivity(player)
				writeTo(w, "joined", "%v", inputID)
				return
			}
		}
		cwlog.DoLog(true, "Could not find lobby: %v for player: %v", inputID, player.ID)
		return

	} else if command == "name" { /* Change player name */
		newName := filterName(data)
		if playerNameUnique(newName) {
			cwlog.DoLog(true, "Changed player '%v' (%v) name to '%v'", player.Name, player.ID, newName)
			player.Name = newName
			playerActivity(player)
		} else {
			cwlog.DoLog(true, "Player (%v) tried to rename to a non-unique name: '%v'", player.ID, newName)
		}
		writeTo(w, "name", "%v", player.Name)
		return
	} else if command == "createLobby" {
		newName := filterName(data)
		newLobby := makePersonalLobby(player, newName)
		if newLobby != nil {
			playerActivity(player)
			writeTo(w, "createdLobby", "%v", newLobby.ID)
			return
		}
		return
	} else {
		cwlog.DoLog(true, "Unknown Command.")
		return
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
