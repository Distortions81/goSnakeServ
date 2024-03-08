package main

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type lobbyData struct {
	ID   uint16 `json:"i"`
	Name string `json:"n"`

	players     []*playerData
	PlayerNames string `json:"p"`
	NumPlayers  uint16 `json:"c"`
	numConn     uint16

	showApple  bool
	apple      XY
	boardSize  uint8
	grid       map[XY]bool
	createdFor uint32

	dirty bool
	lock  sync.Mutex
}

type playerData struct {
	conn *websocket.Conn
	lock sync.Mutex

	id   uint32
	name string

	deadFor   int8
	length    uint16
	tiles     []XY
	head      XY
	oldDir    DIR
	direction DIR
	dirs      []DIR
	numDirs   uint8
	dirToggle bool

	lastActive time.Time
	lastPing   time.Time

	inLobby *lobbyData

	isBot bool
}

type XY struct {
	X uint8
	Y uint8
}
