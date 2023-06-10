package main

import (
	"math/rand"
	"sync"
)

const testLobbys = 15
const testPlayers = 150

var lobbyList []*lobbyData
var lobbyLock sync.RWMutex

var players map[uint64]*playerData
var playersLock sync.RWMutex

func init() {
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	for x := 0; x < testLobbys; x++ {
		newLobby := &lobbyData{ID: makeUID(), Name: genName()}
		lobbyList = append(lobbyList, newLobby)
	}

	players = make(map[uint64]*playerData)

	for x := 0; x < testPlayers; x++ {
		id := makeUID()
		players[id] = &playerData{Name: genName(), ID: id}
	}

	for p := range players {
		rVal := rand.Intn(testLobbys)
		lobbyList[rVal].Players = append(lobbyList[rVal].Players, players[p])
	}
}
