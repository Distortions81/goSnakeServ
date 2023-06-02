package main

import (
	"sync"
)

type lobbyData struct {
	ID   uint64
	Name string

	Players []playerData
	Ticks   uint64
	Level   uint16
	tiles   map[XY]bool

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
