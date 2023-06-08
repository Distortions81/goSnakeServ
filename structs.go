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
	LastActive time.Time

	InLobby   *lobbyData
	MyLobby   *lobbyData
	DeadFor   uint8
	Length    uint32
	Tiles     []XY
	Head      XY
	Direction uint8
}

type XY struct {
	X uint16
	Y uint16
}
