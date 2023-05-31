package main

import (
	"sync"
	"time"
)

type lobbyListing struct {
	Name        string
	PlayerCount string
	PlayerNames string
	Started     string
	Remaining   string
	BoardSize   uint8
	Level       uint16
}

type lobbyData struct {
	ID      uint64
	Name    string
	Players []playerData
	Ticks   uint64
	Started time.Time
	EndTime time.Time
	Level   uint16
	Tiles   [MAX_BOARD_SIZE][MAX_BOARD_SIZE]bool

	Listing lobbyListing
	Lock    sync.Mutex
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
