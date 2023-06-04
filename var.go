package main

func init() {
	players = make(map[uint64]*playerData)
}

var lobbyList = []lobbyData{
	{
		ID:   0,
		Name: "Welcome",
	},
}

var players map[uint64]*playerData
