package main

func makePersonalLobby(player *playerData, name string) *lobbyData {
	if player.inLobby != nil || player.myLobby != nil {
		return nil
	}
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	newLobby := &lobbyData{Name: player.Name + "'s game", ID: player.ID, Ticks: 0, tiles: make(map[XY]bool), Level: 1}
	if name != "" {
		newLobby.Name = name
	}
	lobbyList = append(lobbyList, newLobby)

	lobbyAddr := lobbyList[len(lobbyList)-1]
	player.inLobby = lobbyAddr
	player.myLobby = lobbyAddr
	return lobbyAddr
}
