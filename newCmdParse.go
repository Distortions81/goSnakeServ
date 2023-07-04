package main

import "goSnakeServ/cwlog"

func newParser(data []byte, player *playerData) {
	dataLen := len(data)

	if dataLen <= 0 {
		return
	}

	d := data[0]
	cwlog.DoLog(true, "Header: 0x%x\n", d)

	if d < 0x10 {
		switch d {
		case 0x01:
			//init
		case 0x02:
			//ping
		}
	} else if d < 0x20 {
		switch d {
		case 0x10:
			//Login
		case 0x11:
			//Set Name
		}
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
	} else if d < 0x40 {
		switch d {
		case 0x30:
			//spawn
		case 0x31:
			//go dir
		}
	} else if d < 0x50 {
		//Reserved
	}
}
