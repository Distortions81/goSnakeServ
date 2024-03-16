package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

func newParser(input []byte, player *playerData) {

	inputLen := len(input)

	if inputLen <= 0 {
		return
	}

	d := CMD(input[0])
	data := input[1:]

	cmdName := cmdNames[d]
	if cmdName != "" && d != CMD_PINGPONG {
		doLog(true, "ID: %v, Received: %v", player.id, cmdName)
	}
	switch d {
	case CMD_INIT: /*INIT*/
		cmd_init(player, &data)
	case CMD_PINGPONG: /*PING*/
		cmd_pingpong(player, &data)
	case CMD_GETLOBBIES:
		cmd_getlobbies(player)
	case CMD_JOINLOBBY:
		cmd_joinlobby(player, &data)
	case CMD_CREATELOBBY:
		cmd_createlobby(player)
	case CMD_GODIR:
		cmd_godir(player, data)
	default:
		doLog(true, "Received invalid command: 0x%02X, %v", d, string(data))
		killConnection(player.conn, false)
		deleteFromLobby(player)
		player.conn = nil
		delete(playerList, player.id)
		delete(connPList, player.id)
		return
	}
}

func cmd_createlobby(player *playerData) {
	/* Already in a lobby */
	if player.inLobby != nil {
		return
	}

	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	for _, l := range lobbyList {

		/* If we already have a custom lobby, join it instead */
		if l.createdFor == player.id {
			lobbyID := uint16ToByteArray(l.ID)
			go cmd_joinlobby(player, &lobbyID)
			return
		}
	}

	/* Otherwise make a new lobby */
	newLobby := &lobbyData{
		dirty: true, ID: makeLobbyUID(), Name: player.name + "'s Lobby",
		boardSize: defaultBoardSize, grid: make(map[XY]bool,
			defaultBoardSize*defaultBoardSize), createdFor: player.id}
	lobbyList = append(lobbyList, newLobby)

	lobbyID := uint16ToByteArray(newLobby.ID)
	go cmd_joinlobby(player, &lobbyID)

}

func cmd_init(player *playerData, data *[]byte) {
	/* Players cannot re-init */
	if player.id != 0 {
		return
	}

	var outBuf = new(bytes.Buffer)

	if !checkSecret(nil, *data) {
		killConnection(player.conn, true)
		deleteFromLobby(player)
		player.conn = nil
		delete(playerList, player.id)
		delete(connPList, player.id)
		return
	}

	player.id = makePlayerUID()
	player.lastActive = time.Now()
	player.name = fmt.Sprintf("Player-%v", player.id)

	pListLock.Lock()
	playerList[player.id] = player
	connPList[player.id] = player
	binary.Write(outBuf, binary.LittleEndian, player.id)
	pListLock.Unlock()

	writeToPlayer(player, RECV_LOCALPLAYER, outBuf.Bytes())
}

func cmd_pingpong(player *playerData, data *[]byte) {
	if checkSecret(player, *data) {
		player.lastPing = time.Now()
		writeToPlayer(player, CMD_PINGPONG, generateSecret(player))
	} else {
		doLog(true, "malformed PING")
		killConnection(player.conn, true)
		deleteFromLobby(player)
		player.conn = nil
		delete(playerList, player.id)
		delete(connPList, player.id)
		return
	}
}

/* TODO convert to binary for speed */
func cmd_getlobbies(player *playerData) {
	lobbyLock.Lock()
	data, err := json.Marshal(&lobbyList)
	lobbyLock.Unlock()
	if err != nil || data == nil {
		return
	}
	writeToPlayer(player, RECV_LOBBYLIST, CompressZip(data))
}

func cmd_joinlobby(player *playerData, data *[]byte) {
	inputID := binary.LittleEndian.Uint16(*data)

	if inputID == 0 {
		deleteFromLobby(player)
		return
	}

	if player.inLobby != nil {
		doLog(true, "commandParser: Join: player %v already in a lobby: %v,", player.id, player.inLobby.ID)
		return
	}
	length := 3
	player.deadFor = -8
	player.lastActive = time.Now()

	/* OPTIMIZE */
	lobbyLock.Lock()
	defer lobbyLock.Unlock()
	for l, lobby := range lobbyList {
		if lobby.ID == inputID {
			lobby.lock.Lock()
			defer lobby.lock.Unlock()

			player.direction = DIR_SOUTH
			player.oldDir = DIR_SOUTH
			player.tiles = []XY{}
			player.length = 0
			lobby.numConn++

			/* Reuse dead slots */
			var makeNew bool = true
			for f, find := range lobby.players {
				if find.deadFor > 4 || find.length <= 0 || find.id == player.id {
					if find.inLobby != nil {
						if find.inLobby.ID == lobby.ID {
							continue
						}
					}
					lobby.players[f] = player
					makeNew = false
					doLog(true, "Reused old player slot.")
					break
				}
			}
			if makeNew {
				lobby.players = append(lobby.players, player)
				lobby.dirty = true
			}
			player.inLobby = lobbyList[l]

			var randx, randy uint8
			for x := 0; x < 1000; x++ {
				randx = uint8(rand.Intn(defaultBoardSize))
				randy = uint8(rand.Intn(defaultBoardSize))
				if !willCollidePlayer(player.inLobby, player, player.direction) {
					break
				}
			}

			tiles := []XY{}
			for x := 0; x < length; x++ {
				tiles = append(tiles, XY{X: randx, Y: randy})
			}
			player.tiles = tiles
			player.length = uint16(length)

			doLog(true, "Player: %v joined lobby: %v at %v,%v", player.id, inputID, randx, randy)
			playerActivity(player)

			lobbyList[l].dirty = true
			writeToPlayer(player, CMD_JOINLOBBY, nil)
			return
		}
	}
	doLog(true, "Could not find lobby: %v for player: %v", inputID, player.id)
}

func cmd_godir(player *playerData, data []byte) {
	if player.inLobby != nil {
		player.inLobby.lock.Lock()
		if player.numDirs < 3 {
			player.numDirs++
			player.dirs = append(player.dirs, DIR(data[0]))
			player.lastActive = time.Now()
		}
		player.inLobby.lock.Unlock()
	}
}

func writeToPlayer(player *playerData, header CMD, input []byte) bool {

	if player == nil {
		return false
	}
	if player.conn == nil {

		return false
	}

	player.lock.Lock()
	defer player.lock.Unlock()

	var err error
	if input == nil {
		err = player.conn.WriteMessage(websocket.BinaryMessage, []byte{byte(header)})
	} else {
		err = player.conn.WriteMessage(websocket.BinaryMessage, append([]byte{byte(header)}, input...))
	}
	if err != nil {
		doLog(true, "Error writing response: %v", err)
		killConnection(player.conn, false)
		deleteFromLobby(player)
		player.conn = nil
		delete(playerList, player.id)
		delete(connPList, player.id)
		return false
	}
	return true
}

var nullPlayer = &playerData{id: 0, deadFor: 1, length: 0, tiles: []XY{}}

func deleteFromLobby(player *playerData) {

	if player.inLobby == nil {
		return
	}

	player.inLobby.lock.Lock()
	defer player.inLobby.lock.Unlock()

	player.inLobby.numConn--
	player.inLobby.dirty = true
	player.deadFor = 1

	for t, test := range player.inLobby.players {
		if test.id == player.id {
			player.inLobby.players[t] = nullPlayer
			break
		}
	}
	player.inLobby = nil

	doLog(true, "Deleted %v from lobby", player.name)
}
