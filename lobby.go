package main

import "sync"

const MAX_BOARD_SIZE = 0xFFFF

type lobbyData struct {
	ID      uint64
	Name    string
	Players []playerData
	Tiles   [MAX_BOARD_SIZE][MAX_BOARD_SIZE]bool

	Lock sync.Mutex
}

type playerData struct {
	ID   uint64
	Name string

	Dead      bool
	Length    uint32
	Tiles     []XY
	Head      XY
	Direction uint8
}

type XY struct {
	X uint16
	Y uint16
}
