package main

import "sync"

func init() {
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	players = make(map[uint64]*playerData)
}

var testPlayerList = []*playerData{
	{Name: "foxvenusnoodles"},
	{Name: "cereseggleopard"},
	{Name: "foxvealoldeuboi"},
	{Name: "swanednamodedog"},
	{Name: "pigvegashooting"},
}

var testPlayerListTwo = []*playerData{
	{Name: "lamerburgermeisterwithbutter"},
	{Name: "idiotsandwich"},
	{Name: "inquisitiveidiot"},
}

var lobbyList = []*lobbyData{
	{
		ID:      0,
		Name:    "Test Lobby",
		Players: testPlayerList,
	},
	{
		ID:      1,
		Name:    "super long name n00b room with sprinkles and spam",
		Players: testPlayerListTwo,
	},
	{
		ID:      2,
		Name:    "lobby lobby",
		Players: nil,
	},
	{
		ID:      3,
		Name:    "something somthing lobby",
		Players: testPlayerList,
	},
	{
		ID:      4,
		Name:    "hork bork",
		Players: nil,
	},
}

var players map[uint64]*playerData
var lobbyLock sync.Mutex
