package main

import (
	"sync"
)

type lobbyData struct {
	ID      uint64
	Name    string
	Players []playerData
	Ticks   uint64
	Level   uint16
	tiles   [MAX_BOARD_SIZE][MAX_BOARD_SIZE]bool

	lock sync.Mutex
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
