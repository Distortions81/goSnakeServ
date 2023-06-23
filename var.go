package main

import (
	"math/rand"
	"sync"
)

const (
	testLobbys       = 15
	testPlayers      = 150
	defaultBoardSize = 32
)

var (
	lobbyList []*lobbyData
	lobbyLock sync.RWMutex

	pList     map[uint64]*playerData
	pListLock sync.RWMutex
)

func init() {
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	for x := 0; x < testLobbys; x++ {
		newLobby := &lobbyData{ID: makeUID(), Name: genName(), boardSize: defaultBoardSize}
		lobbyList = append(lobbyList, newLobby)
	}

	pList = make(map[uint64]*playerData)

	length := 10
	for x := 0; x < testPlayers; x++ {
		id := makeUID()
		pList[id] = &playerData{Name: genName(), ID: id, Length: uint32(length), Direction: uint8(rand.Intn(DIR_WEST)), isBot: true}
	}

	for p := range pList {
		rVal := rand.Intn(testLobbys)
		lobbyList[rVal].Players = append(lobbyList[rVal].Players, pList[p])
		pList[p].inLobby = lobbyList[rVal]

		var randx, randy uint16
		for x := 0; x < 10000; x++ {
			randx = uint16(rand.Intn(defaultBoardSize))
			randy = uint16(rand.Intn(defaultBoardSize))
			if !didCollidePlayer(pList[p].inLobby, pList[p]) {
				break
			}
		}

		tiles := []XY{}
		for x := 0; x < length; x++ {
			tiles = append(tiles, XY{X: randx, Y: randy})
		}
		pList[p].Tiles = tiles
	}

}
