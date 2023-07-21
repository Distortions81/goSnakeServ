package main

func makePersonalLobby(player *playerData, name string) *lobbyData {
	if player.inLobby != nil || player.myLobby != nil {
		return nil
	}
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	newLobby := &lobbyData{Name: player.Name + "'s game", ID: player.ID, Ticks: 0, Level: 1}
	if name != "" {
		newLobby.Name = name
	}
	lobbyList = append(lobbyList, newLobby)

	lobbyAddr := lobbyList[len(lobbyList)-1]
	player.inLobby = lobbyAddr
	player.myLobby = lobbyAddr
	return lobbyAddr
}

func makeLobby(name string) *lobbyData {
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	newLobby := &lobbyData{ID: makeLobbyUID(), Name: genName(), boardSize: defaultBoardSize}
	lobbyList = append(lobbyList, newLobby)

	return newLobby
}
