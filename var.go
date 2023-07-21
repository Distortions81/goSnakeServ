package main

import (
	"math/rand"
	"sync"
)

const (
	testlobbies      = 4
	testPlayers      = testlobbies * 8
	defaultBoardSize = 32
)

var (
	lobbyList []*lobbyData
	lobbyLock sync.RWMutex

	pList     map[uint32]*playerData
	pListLock sync.RWMutex
)

func init() {
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	makeTestLobbies()

	pList = make(map[uint32]*playerData)

	makeAIs()
}

func makeTestLobbies() {
	for x := 0; x < testlobbies; x++ {
		newLobby := &lobbyData{ID: makeLobbyUID(), Name: genName(), boardSize: defaultBoardSize}
		lobbyList = append(lobbyList, newLobby)
	}
}

func makeAIs() {
	length := 3
	for x := 0; x < testPlayers; x++ {
		id := makePlayerUID()
		pList[id] = &playerData{Name: genName(), id: id, length: uint16(length), direction: uint8(rand.Intn(DIR_WEST)), isBot: true, DeadFor: -8}
	}

	for p := range pList {
		rVal := rand.Intn(testlobbies)
		lobbyList[rVal].Players = append(lobbyList[rVal].Players, pList[p])
		lobbyList[rVal].dirty = true

		pList[p].inLobby = lobbyList[rVal]

		var randx, randy uint8
		for x := 0; x < 10000; x++ {
			randx = uint8(rand.Intn(defaultBoardSize))
			randy = uint8(rand.Intn(defaultBoardSize))
			if !didCollidePlayer(pList[p].inLobby, pList[p]) {
				break
			}
		}

		tiles := []XY{}
		for x := 0; x < length; x++ {
			tiles = append(tiles, XY{X: randx, Y: randy})
		}
		pList[p].tiles = tiles
	}
}
