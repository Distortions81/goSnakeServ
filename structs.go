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

	lock sync.Mutex
}

type playerData struct {
	Desc       http.ResponseWriter
	ID         uint64
	Name       string
	lastActive time.Time

	inLobby *lobbyData
	myLobby *lobbyData

	deadFor   uint8
	Length    uint32
	Tiles     []XY
	Head      XY
	Direction uint8

	lock sync.Mutex
}

type XY struct {
	X uint16
	Y uint16
}
