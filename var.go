package main

import (
	"math/rand"
	"sync"
)

const testLobbys = 15
const testPlayers = 150

var lobbyList []*lobbyData
var lobbyLock sync.RWMutex

var pList map[uint64]*playerData
var pListLock sync.RWMutex

func init() {
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	for x := 0; x < testLobbys; x++ {
		newLobby := &lobbyData{ID: makeUID(), Name: genName(), boardSize: 32}
		lobbyList = append(lobbyList, newLobby)
	}

	pList = make(map[uint64]*playerData)

	for x := 0; x < testPlayers; x++ {
		id := makeUID()
		pList[id] = &playerData{Name: genName(), ID: id, Tiles: []XY{}}
	}

	for p := range pList {
		rVal := rand.Intn(testLobbys)
		lobbyList[rVal].Players = append(lobbyList[rVal].Players, pList[p])
	}
}
