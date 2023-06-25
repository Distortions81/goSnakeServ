package main

import (
	"sync"
	"time"
)

type lobbyData struct {
	ID   uint64 `json:"i,omitempty"`
	Name string `json:"n,omitempty"`

	Players   []*playerData `json:"p,omitempty"`
	Ticks     uint64        `json:"t,omitempty"`
	Level     uint16        `json:"l,omitempty"`
	ShowApple bool          `json:"s,omitempty"`
	Apple     XY            `json:"a,omitempty"`
	boardSize uint16

	//outBuf []byte

	lock sync.Mutex
}

type playerData struct {
	ID   uint64 `json:"i,omitempty"`
	Name string `json:"n,omitempty"`

	DeadFor   uint8  `json:"x,omitempty"`
	Length    uint32 `json:"l,omitempty"`
	Tiles     []XY   `json:"t,omitempty"`
	Head      XY     `json:"h,omitempty"`
	Direction uint8  `json:"d,omitempty"`

	lastActive time.Time

	inLobby *lobbyData
	myLobby *lobbyData
	isBot   bool
}

type XY struct {
	X uint16
	Y uint16
}
