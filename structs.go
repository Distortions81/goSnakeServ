package main

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
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

	lock sync.Mutex
}

type playerData struct {
	conn     *websocket.Conn
	connLock sync.Mutex

	ID   uint64 `json:"i"`
	Name string `json:"n"`

	DeadFor   uint8  `json:"x"`
	Length    uint32 `json:"l"`
	Tiles     []XY   `json:"t"`
	Head      XY     `json:"h"`
	oldDir    uint8
	Direction uint8 `json:"d"`

	lastActive time.Time
	lastPing   time.Time

	inLobby *lobbyData
	myLobby *lobbyData
	isBot   bool
}

type XY struct {
	X uint16
	Y uint16
}
