package main

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type lobbyData struct {
	ID   uint16 `json:"i"`
	Name string `json:"n"`

	Players []*playerData `json:"p"`

	showApple bool
	apple     XY
	boardSize uint8

	dirty bool
	lock  sync.Mutex
}

type playerData struct {
	conn     *websocket.Conn
	connLock sync.Mutex

	id   uint32
	Name string `json:"n"`

	DeadFor   int8 `json:"x"`
	length    uint16
	tiles     []XY
	head      XY
	oldDir    uint8
	direction uint8
	dirs      []uint8
	numDirs   uint8

	lastActive time.Time
	lastPing   time.Time

	inLobby *lobbyData

	isBot bool
}

type XY struct {
	X uint8
	Y uint8
}
