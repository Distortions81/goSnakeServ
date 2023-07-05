package main

import (
	"encoding/binary"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

func newParser(input []byte, player *playerData) {
	inputLen := len(input)

	if inputLen <= 0 {
		return
	}

	d := input[0]
	data := input[1:]

	cmdName := cmdNames[d]
	if cmdName != "" {
		doLog(true, "Received: %v\n", cmdName)
	}
	switch d {
	case CMD_INIT: //INIT
		if !checkSecret(nil, data) {
			player.conn.Close()
			return
		}

		player.ID = makeUID()
		player.lastActive = time.Now()
		player.Name = genName()

		pListLock.Lock()
		pList[player.ID] = player
		pListLock.Unlock()

		b, err := json.Marshal(player)
		if err != nil {
			doLog(true, "newParser: init: err: %v", err)
			return
		}

		writeToPlayer(player, byte(RECV_LOCALPLAYER), b)
	case CMD_PINGPONG: //PING
		if checkSecret(player, data) {
			//doLog(true, "PING")
			player.lastPing = time.Now()
			writeToPlayer(player, byte(CMD_PINGPONG), generateSecret(player))
		} else {
			doLog(true, "malformed PING")
			player.conn.Close()
			return
		}
	case CMD_LOGIN:
		//Login
	case CMD_NAME:
		//Set Name

	case CMD_GETLOBBIES:
		lobbyLock.Lock()
		data, err := json.Marshal(&lobbyList)
		lobbyLock.Unlock()
		if err != nil || data == nil {
			return
		}
		writeToPlayer(player, RECV_LOBBYLIST, data)
	case CMD_JOINLOBBY:

		inputID := binary.BigEndian.Uint64(data)

		if inputID == 0 {
			deleteFromLobby(player)
			return
		}

		if player.inLobby != nil {
			doLog(true, "commandParser: Join: player %v already in a lobby: %v,", player.ID, player.inLobby.ID)
			return
		}
		length := 3

		/* OPTIMIZE */
		for l, lobby := range lobbyList {
			if lobby.ID == inputID {
				lobby.lock.Lock()
				defer lobby.lock.Unlock()

				player.Direction = DIR_SOUTH
				player.oldDir = DIR_SOUTH

				/* Reuse dead slots */
				var makeNew bool = true
				for f, find := range lobby.Players {
					if find.DeadFor > 4 {
						lobby.Players[f] = player
						makeNew = false
						doLog(true, "Reused old player slot.")
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

				doLog(true, "Player: %v joined lobby: %v at %v,%v", player.ID, inputID, randx, randy)
				playerActivity(player)
				writeToPlayer(player, CMD_JOINLOBBY, data)
				return
			}
		}
		doLog(true, "Could not find lobby: %v for player: %v", inputID, player.ID)
		return
	case CMD_CREATELOBBY:
		//create lobby
	case CMD_SETLOBBY:
		//set lobby settings

	case CMD_SPAWN:
		//spawn
	case CMD_GODIR:
		if player.inLobby != nil {
			player.inLobby.lock.Lock()
			if player.numDirs < 3 {
				player.numDirs++
				player.dirs = append(player.dirs, uint8(data[0]))
			}
			player.inLobby.lock.Unlock()
		}
	default:
		doLog(true, "Received invalid: 0x%02X, %v\n", d, string(data))
		player.conn.Close()
		return
	}
}

func writeToPlayer(player *playerData, header byte, input []byte) bool {

	if player.conn == nil {
		return false
	}

	player.connLock.Lock()
	defer player.connLock.Unlock()

	var err error
	if input == nil {
		err = player.conn.WriteMessage(websocket.BinaryMessage, []byte{header})
	} else {
		err = player.conn.WriteMessage(websocket.BinaryMessage, append([]byte{header}, input...))
	}
	if err != nil {
		doLog(true, "Error writing response: %v", err)
		player.conn.Close()
		return false
	}
	return true
}

func deleteFromLobby(player *playerData) {
	if player.inLobby == nil {
		return
	}

	player.inLobby.lock.Lock()
	defer player.inLobby.lock.Unlock()

	deleteme := -1
	count := 0
	for p, target := range player.inLobby.Players {
		count++
		if target.ID == player.ID {
			deleteme = p
		}
	}

	if deleteme != -1 {
		if count > 0 {
			player.inLobby.Players = append(player.inLobby.Players[:deleteme], player.inLobby.Players[deleteme+1:]...)
		} else {
			player.inLobby.Players = nil
		}
	}

	player.DeadFor = 1
	player.inLobby = nil

}
