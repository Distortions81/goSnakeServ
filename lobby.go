package main

func makeLobby(name string) *lobbyData {
	lobbyLock.Lock()
	defer lobbyLock.Unlock()

	newLobby := &lobbyData{ID: makeLobbyUID(), Name: genName(), boardSize: defaultBoardSize}
	lobbyList = append(lobbyList, newLobby)

	return newLobby
}
