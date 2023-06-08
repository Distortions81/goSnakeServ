package main

func makePersonalLobby(player *playerData) *lobbyData {
	if player.InLobby != nil || player.MyLobby != nil {
		return nil
	}
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	newLobby := lobbyData{Name: player.Name + "'s game", ID: player.ID, Ticks: 0, tiles: make(map[XY]bool), Level: 1}
	lobbyList = append(lobbyList, newLobby)

	lobbyAddr := &lobbyList[len(lobbyList)-1]
	player.InLobby = lobbyAddr
	player.MyLobby = lobbyAddr
	return lobbyAddr
}
