package main

import (
	"encoding/json"
	"goSnakeServ/cwlog"
	"time"

	"github.com/gorilla/websocket"
)

func newParser(data []byte, player *playerData) {
	dataLen := len(data)

	if dataLen <= 0 {
		return
	}

	d := data[0]

	cmdName := cmdNames[d]
	if cmdName == "" {
		cwlog.DoLog(true, "Header: 0x%02X\n", d)
	} else {
		cwlog.DoLog(true, "Header: %v\n", cmdName)
	}

	switch d {
	case CMD_INIT: //INIT
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
		writeToPlayer(player, byte(CMD_PINGPONG), nil)
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
		cwlog.DoLog(true, "Invalid header: 0x%02X", d)
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
