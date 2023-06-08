package main

import (
	"math/rand"
	"sync"
)

const testLobbys = 3
const testPlayers = 7

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
		players[p].inLobby = lobbyList[rand.Intn(testLobbys)]
		players[p].inLobby.Players = append(players[p].inLobby.Players, players[p])
	}
}

var lobbyList = []*lobbyData{}

var players map[uint64]*playerData
var lobbyLock sync.Mutex
