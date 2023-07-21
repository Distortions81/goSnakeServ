package main

import (
	"bytes"
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
		doLog(true, "Received: %v", cmdName)
	}
	switch d {
	case CMD_INIT: //INIT
		if !checkSecret(nil, data) {
			player.conn.Close()
			return
		}

		player.id = makePlayerUID()
		player.lastActive = time.Now()
		player.Name = genName()

		pListLock.Lock()
		pList[player.id] = player
		var outBuf = new(bytes.Buffer)
		binary.Write(outBuf, binary.BigEndian, player.id)
		pListLock.Unlock()

		writeToPlayer(player, byte(RECV_LOCALPLAYER), outBuf.Bytes())
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

		inputID := binary.BigEndian.Uint16(data)

		if inputID == 0 {
			deleteFromLobby(player)
			return
		}

		if player.inLobby != nil {
			doLog(true, "commandParser: Join: player %v already in a lobby: %v,", player.id, player.inLobby.ID)
			return
		}
		length := 3
		player.DeadFor = -8

		/* OPTIMIZE */
		for l, lobby := range lobbyList {
			if lobby.ID == inputID {
				lobby.lock.Lock()
				defer lobby.lock.Unlock()

				player.direction = DIR_SOUTH
				player.oldDir = DIR_SOUTH
				player.tiles = []XY{}

				/* Reuse dead slots */
				var makeNew bool = true
				for f, find := range lobby.Players {
					if find.DeadFor > 4 {
						if find.inLobby != nil {
							if find.inLobby.ID == lobby.ID {
								continue
							}
						}
						lobby.Players[f] = player
						makeNew = false
						doLog(true, "Reused old player slot.")
						break
					}
				}
				if makeNew {
					lobby.Players = append(lobby.Players, player)
					lobby.dirty = true
				}
				player.inLobby = lobbyList[l]

				var randx, randy uint8
				for x := 0; x < 10000; x++ {
					randx = uint8(rand.Intn(defaultBoardSize))
					randy = uint8(rand.Intn(defaultBoardSize))
					if !didCollidePlayer(player.inLobby, player) {
						break
					}
				}

				tiles := []XY{}
				for x := 0; x <= length; x++ {
					tiles = append(tiles, XY{X: randx, Y: randy})
				}
				player.tiles = tiles
				player.length = uint16(length - 1)

				doLog(true, "Player: %v joined lobby: %v at %v,%v", player.id, inputID, randx, randy)
				playerActivity(player)
				writeToPlayer(player, CMD_JOINLOBBY, serializeLobbyBinary(lobby))
				return
			}
		}
		doLog(true, "Could not find lobby: %v for player: %v", inputID, player.id)
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
		doLog(true, "Received invalid: 0x%02X, %v", d, string(data))
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

	player.inLobby.dirty = true
	player.DeadFor = 1
	player.inLobby = nil

	doLog(true, "Deleted %v from lobby", player.Name)
}
