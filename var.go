package main

import (
	"math/rand"
	"sync"
)

const (
	testlobbies      = 2
	testPlayers      = testlobbies * 4
	defaultBoardSize = 32
)

var (
	lobbyList []*lobbyData
	lobbyLock sync.RWMutex

	playerList map[uint32]*playerData
	connPList  map[uint32]*playerData
	pListLock  sync.RWMutex
)

func init() {
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	makeTestLobbies()

	playerList = make(map[uint32]*playerData)
	connPList = make(map[uint32]*playerData)

	makeAIs()
}

func makeTestLobbies() {
	for x := 0; x < testlobbies; x++ {
		newLobby := &lobbyData{dirty: true, ID: makeLobbyUID(), Name: genName(), boardSize: defaultBoardSize, grid: make(map[XY]bool, defaultBoardSize*defaultBoardSize)}
		lobbyList = append(lobbyList, newLobby)
	}
}

func makeAIs() {
	length := 3
	for x := 0; x < testPlayers; x++ {
		id := makePlayerUID()
		playerList[id] = &playerData{name: genName(), id: id, length: uint16(length), direction: DIR(rand.Intn(int(DIR_WEST))), isBot: true, deadFor: -8}
	}

	for p := range playerList {
		rVal := rand.Intn(testlobbies)
		lobbyList[rVal].players = append(lobbyList[rVal].players, playerList[p])
		lobbyList[rVal].dirty = true

		playerList[p].inLobby = lobbyList[rVal]

		var randx, randy uint8
		for x := 0; x < 10000; x++ {
			randx = uint8(rand.Intn(defaultBoardSize))
			randy = uint8(rand.Intn(defaultBoardSize))
			if !willCollidePlayer(playerList[p].inLobby, playerList[p], playerList[p].direction) {
				break
			}
		}

		tiles := []XY{}
		for x := 0; x < length; x++ {
			tiles = append(tiles, XY{X: randx, Y: randy})
		}
		playerList[p].tiles = tiles
	}
}
