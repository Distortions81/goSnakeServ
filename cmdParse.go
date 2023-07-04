package main

import (
	"encoding/json"
	"fmt"
	"goSnakeServ/cwlog"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func commandParser(input string, c *websocket.Conn) {

	/* Before ID check */
	if input == "init" {
		id := makeUID()
		newPlayer := playerData{Name: genName(), ID: id, lastActive: time.Now(), Direction: DIR_SOUTH}

		cwlog.DoLog(true, "Created player %v (%v).", newPlayer.Name, newPlayer.ID)

		b, err := json.Marshal(newPlayer)
		if err != nil {
			cwlog.DoLog(true, "commandParser: init: err: %v", err)
			return
		}

		pListLock.Lock()
		pList[id] = &newPlayer
		writeByte(c, b)
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

		dir := uint8(val)
		if reverseDir(dir) != player.Direction {
			player.Direction = dir
		}

		player.inLobby.lock.Unlock()
		lobbyLock.Unlock()

	} else if command == "keyframe" {
		player.inLobby.lock.Lock()
		/* Prevent reversing into self */
		d, _ := strconv.ParseUint(data, 10, 8)
		dir := uint8(d)
		if reverseDir(dir) != player.Direction {
			player.Direction = dir
		}

		buf, _ := json.Marshal(player.inLobby)
		player.inLobby.lock.Unlock()

		writeByte(c, (buf))

	} else if command == "ping" { /* Keep alive, and check latency */
		cwlog.DoLog(true, "Client: %v (PING)", player.ID)
		playerActivity(player)
		writeByte(c, []byte("PONG"))
		return

	} else if command == "list" { /* List lobbies */
		b, _ := json.Marshal(lobbyList)
		playerActivity(player)
		writeByte(c, b)
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
		length := 3
		for l, lobby := range lobbyList {
			if lobby.ID == inputID {
				player.Direction = DIR_SOUTH

				/* Reuse dead slots */
				var makeNew bool = true
				for f, find := range lobby.Players {
					if find.DeadFor > 4 {
						lobby.Players[f] = player
						makeNew = false
						cwlog.DoLog(true, "Reused old player slot.")
						break
					}
				}
				if makeNew {
					lobby.Players = append(lobby.Players, player)
				}
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

				cwlog.DoLog(true, "Player: %v joined lobby: %v at %v,%v", player.ID, inputID, randx, randy)
				playerActivity(player)
				writeTo(c, "joined", "%v", inputID)
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
		writeTo(c, "name", "%v", player.Name)
		return
	} else if command == "createLobby" {
		newName := filterName(data)
		newLobby := makePersonalLobby(player, newName)
		if newLobby != nil {
			playerActivity(player)
			writeTo(c, "createdLobby", "%v", newLobby.ID)
			return
		}
		return
	} else {
		cwlog.DoLog(true, "Unknown Command.")
		return
	}
}

func writeByte(c *websocket.Conn, input []byte) bool {
	err := c.WriteMessage(websocket.BinaryMessage, input)
	if err != nil {
		cwlog.DoLog(true, "Error writing response: %v", err)
		c.Close()
		return false
	}
	return true
}

func writeByteTo(c *websocket.Conn, command string, input []byte) bool {
	buf := []byte(command + ":")
	buf = append(buf[:], input[:]...)

	err := c.WriteMessage(websocket.BinaryMessage, buf)
	if err != nil {
		cwlog.DoLog(true, "Error writing response: %v", err)
		c.Close()
		return false
	}

	cwlog.DoLog(true, "WroteTo %v:%v", command, string(input))
	return true
}

func writeStringTo(c *websocket.Conn, command string, input string) bool {
	return writeByteTo(c, command, []byte(input))
}

func writeTo(c *websocket.Conn, command string, inputFormat string, args ...interface{}) bool {
	input := fmt.Sprintf(inputFormat, args...)
	return writeStringTo(c, command, input)
}
