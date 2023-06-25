package main

import (
	"sync"
	"time"
)

type lobbyData struct {
	ID   uint64 `json:"i"`
	Name string `json:"n"`

	Players   []*playerData `json:"p"`
	Ticks     uint64        `json:"t"`
	Level     uint16        `json:"l"`
	ShowApple bool          `json:"s"`
	Apple     XY            `json:"a"`
	boardSize uint16

	//outBuf []byte

	lock sync.Mutex
}

type playerData struct {
	ID   uint64 `json:"i"`
	Name string `json:"n"`

	DeadFor   uint8  `json:"x"`
	Length    uint32 `json:"l"`
	Tiles     []XY   `json:"t"`
	Head      XY     `json:"h"`
	Direction uint8  `json:"d"`

	lastActive time.Time

	inLobby *lobbyData
	myLobby *lobbyData
	isBot   bool
}

type XY struct {
	X uint16
	Y uint16
}
