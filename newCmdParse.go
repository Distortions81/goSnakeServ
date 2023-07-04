package main

import (
	"encoding/json"
	"goSnakeServ/cwlog"
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
	if cmdName == "" {
		cwlog.DoLog(true, "Received: 0x%02X, %v\n", d, string(data))
	} else {
		cwlog.DoLog(true, "Received: %v, %v\n", cmdName, string(data))
	}
	switch d {
	case CMD_INIT: //INIT
		if !checkSecret(data) {
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
			cwlog.DoLog(true, "newParser: init: err: %v", err)
			return
		}

		writeToPlayer(player, byte(RECV_LOCALPLAYER), b)
	case CMD_PINGPONG: //PING
		if checkSecret(data) {
			//cwlog.DoLog(true, "PING")
			writeToPlayer(player, byte(CMD_PINGPONG), generateSecret())
		} else {
			cwlog.DoLog(true, "malformed PING")
			player.conn.Close()
			return
		}
	case CMD_LOGIN:
		//Login
	case CMD_NAME:
		//Set Name

	case CMD_GETLOBBIES:
		data, err := json.Marshal(&lobbyList)
		if err != nil || data == nil {
			return
		}
		writeToPlayer(player, RECV_LOBBYLIST, data)
	case CMD_JOINLOBBY:
		//join lobby
	case CMD_CREATELOBBY:
		//create lobby
	case CMD_SETLOBBY:
		//set lobby settings

	case CMD_SPAWN:
		//spawn
	case CMD_GODIR:
		//go dir

	default:
		cwlog.DoLog(true, "Received invalid: 0x%02X, %v\n", d, string(data))
		player.conn.Close()
		return
	}
}

func writeToPlayer(player *playerData, header byte, input []byte) bool {

	if player.conn == nil {
		return false
	}

	var err error
	if input == nil {
		err = player.conn.WriteMessage(websocket.BinaryMessage, []byte{header})
	} else {
		err = player.conn.WriteMessage(websocket.BinaryMessage, append([]byte{header}, input...))
	}
	if err != nil {
		cwlog.DoLog(true, "Error writing response: %v", err)
		player.conn.Close()
		return false
	}
	return true
}
