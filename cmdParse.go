package main

import (
	"encoding/json"
	"fmt"
	"goSnakeServ/cwlog"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func commandParser(input string, w http.ResponseWriter) {

	/* Before ID check */
	if input == "init" {
		id := makeUID()
		newPlayer := playerData{Name: genName(), ID: id, LastActive: time.Now(), Direction: DIR_SOUTH}

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

	//cwlog.DoLog(true, "%v: %v: '%v'", cmdPart[0], cmdPart[1], cmdPart[2])
	useridstr, command, data := cmdPart[0], cmdPart[1], cmdPart[2]
	userid, _ := strconv.ParseUint(useridstr, 10, 64)

	/* Find player, if invalid exit */
	pListLock.RLock()
	player := pList[userid]
	pListLock.RUnlock()

	if player == nil {
		cwlog.DoLog(true, "Invalid userid: %v", useridstr)
		return
	}

	if command == "go" {
		val, err := strconv.ParseUint(data, 10, 8)
		if err != nil {
			cwlog.DoLog(true, "commandParser: go: ParseUint: Error: %v", err)
			return
		}
		if player.inLobby == nil {
			return
		}
		lobbyLock.Lock()
		player.inLobby.lock.Lock()
		player.Direction = uint8(val)
		writeByte(w, player.inLobby.outBuf)
		player.inLobby.lock.Unlock()
		lobbyLock.Unlock()

	} else if command == "keyframe" {

		player.inLobby.lock.Lock()
		dir, _ := strconv.ParseUint(data, 10, 8)
		player.Direction = uint8(dir)
		buf, err := json.Marshal(player.inLobby)
		player.inLobby.lock.Unlock()

		if err != nil {
			cwlog.DoLog(true, "commandParser: keyframe: Marshal: Error: %v", err)
			return
		}
		lobbyLock.Lock()
		writeByte(w, buf)
		lobbyLock.Unlock()
	} else if command == "ping" { /* Keep alive, and check latency */
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
		length := 10
		for l, lobby := range lobbyList {
			if lobby.ID == inputID {
				player.Direction = DIR_SOUTH
				lobby.Players = append(lobby.Players, player)
				player.inLobby = lobbyList[l]

				var randx, randy uint16
				for x := 0; x < 10000; x++ {
					randx = uint16(rand.Intn(defaultBoardSize))
					randy = uint16(rand.Intn(defaultBoardSize))
					if !didCollidePlayer(player.inLobby, player) {
						break
					}
				}

				tiles := []XY{}
				for x := 0; x < length; x++ {
					tiles = append(tiles, XY{X: randx, Y: randy})
				}
				player.Tiles = tiles
				player.Length = uint32(length)

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
