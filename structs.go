package main

import (
	"sync"
	"time"
)

type lobbyData struct {
	ID   uint64
	Name string

	Players []*playerData
	Ticks   uint64
	Level   uint16
	tiles   map[XY]bool

	lock sync.Mutex
}

type playerData struct {
	ID         uint64
	Name       string
	lastActive time.Time

	inLobby   *lobbyData
	myLobby   *lobbyData
	deadFor   uint8
	Length    uint32
	Tiles     []XY
	Head      XY
	Direction uint8
}

type XY struct {
	X uint16
	Y uint16
}
