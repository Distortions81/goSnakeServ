package main

func init() {
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

var lobbyList = []lobbyData{
	{
		ID:      0,
		Name:    "Test Lobby",
		Players: testPlayerList,
	},
	{
		ID:      0,
		Name:    "super long name n00b room with sprinkles and spam",
		Players: testPlayerListTwo,
	},
}

var players map[uint64]*playerData
