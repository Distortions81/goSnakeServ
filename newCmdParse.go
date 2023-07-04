package main

import (
	"encoding/json"
	"goSnakeServ/cwlog"
	"time"
)

func newParser(data []byte, player *playerData) {
	dataLen := len(data)

	if dataLen <= 0 {
		return
	}

	d := data[0]
	cwlog.DoLog(true, "Header: 0x%x\n", d)

	/* Early connect */
	if d < 0x10 {
		switch d {
		case 0x01: //INIT
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

			writeByte(player, byte(0x50), b)
		case 0x02: //PING
			writeByte(player, byte(0x02), nil)
		}
		/* Login */
	} else if d < 0x20 {
		switch d {
		case 0x10:
			//Login
		case 0x11:
			//Set Name
		}
		/* Lobby */
	} else if d < 0x30 {
		switch d {
		case 0x20:
			//get lobbies
		case 0x21:
			//join lobby
		case 0x22:
			//create lobby
		case 0x23:
			//set lobby settings
		}
		/* Game */
	} else if d < 0x40 {
		switch d {
		case 0x30:
			//spawn
		case 0x31:
			//go dir
		}
		/* Reserved */
	} else if d < 0x50 {
		//Reserved
	}
}
