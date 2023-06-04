package main

func init() {
	players = make(map[uint64]*playerData)
}

var testPlayerList = []*playerData{
	{Name: "IUseArchBTW1337hax0rluserftw"},
	{Name: "SteamDeckUser"},
	{Name: "SnakeyMcSnakeFace"},
	{Name: "ITHINKIAMCLEVERANDVERYFUNNYHAHAHAHA12345679ROFLMAOBBQ"},
}

var lobbyList = []lobbyData{
	{
		ID:      0,
		Name:    "People sometimes use really long names for stuff and things",
		Players: testPlayerList,
	},
}

var players map[uint64]*playerData
