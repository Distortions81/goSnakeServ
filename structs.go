package main

import (
	"net/http"
	"sync"
	"time"
)

type lobbyData struct {
	ID   uint64
	Name string

	Players   []*playerData
	Ticks     uint64
	Level     uint16
	tiles     map[XY]bool
	boardSize uint16

	outBuf []byte

	lock sync.Mutex
}

type playerData struct {
	desc       http.ResponseWriter
	ID         uint64
	Name       string
	LastActive time.Time

	inLobby *lobbyData
	myLobby *lobbyData

	DeadFor   uint8
	Length    uint32
	Tiles     []XY
	Head      XY
	Direction uint8
	gameTick  uint64
	isBot     bool

	lock sync.Mutex
}

type XY struct {
	X uint16
	Y uint16
}
