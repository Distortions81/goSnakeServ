package main

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"runtime"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

func processLobbies() {
	var numBytes int
	var gameTicks int

	/* New thread*/
	go func() {

		/*Sized wait group*/
		wg := sizedwaitgroup.New(runtime.NumCPU())

		/*Infinite loop*/
		for {
			loopStart := time.Now()

			gameTicks++

			var numPlayers int

			/*Lock the lobbies*/
			lobbyLock.Lock()
			for l := range lobbyList {
				if lobbyList[l].numConn == 0 && gameTicks > 10 {
					continue
				}
				/*TODO: Optimize: process (lobbies / threads) each, to lower overhead*/

				wg.Add()
				go func(l int) {
					/* Lock the individual lobby */
					lobbyList[l].lock.Lock()
					defer lobbyList[l].lock.Unlock()

					lobby := lobbyList[l]

					/* If needed, spawn apple */
					if !lobby.showApple {
						spawnApple(lobby)
					}

					playersAlive := 0
					totalPlayers := 0

					lobby.PlayerNames = ""
					for p, player := range lobby.players {

						if player.deadFor != 0 {
							continue
						}
						if p <= lobbyMaxNames {
							if p > 0 {
								lobby.PlayerNames = lobby.PlayerNames + ", "
							}
							lobby.PlayerNames = lobby.PlayerNames + player.name
						}

						/*Found players instead of using len()*/
						totalPlayers++

						/* Ignore, dead or not init */
						if player.length < 1 || player.id == 0 {
							continue
						}
						/* Player is now dead */
						if player.deadFor > 0 {

							/* Flashing animation done, erase tiles */
							if player.deadFor > 4 {
								for _, tile := range player.tiles {
									lobby.grid[tile] = false
								}
								player.tiles = nil
								player.length = 0
								continue
							}
							player.deadFor++
							continue
						} else if player.deadFor < 0 {
							player.deadFor++
							continue
						}
						playersAlive++

						/* Test basic AI */
						if player.isBot {
							aiMove(player)
						} else {

							/* Grab a queued command, if there is one */
							if player.numDirs > 0 {
								newDir := player.dirs[0]

								if player.numDirs > 1 {
									player.dirs = append(player.dirs[:0], player.dirs[1:]...)
								} else {
									player.dirs = nil
								}
								player.numDirs--

								/* Don't reverse into self! */
								if newDir != reverseDir(player.oldDir) {
									player.direction = newDir
								}
								player.oldDir = player.direction
							}
						}
					}
					/* Another loop, so we don't have a rolling effect */
					for _, player := range lobby.players {

						/* Ignore, dead or not init */
						if player.length < 1 || player.id == 0 || player.deadFor != 0 {
							continue
						}

						/* go in direction */
						newHead := goDir(player.direction, player.tiles[player.length-1])
						if newHead.X > lobby.boardSize || newHead.X < 1 ||
							newHead.Y > lobby.boardSize || newHead.Y < 1 || willCollidePlayer(player.inLobby, player, player.direction) {
							player.deadFor = 1

							ptype := "Player"
							if player.isBot {
								ptype = "Bot"
							}

							if !player.isBot {
								doLog(true, "%v %v (%v) died at %v,%v in lobby %v.", ptype, player.name, player.id, player.head.X, player.head.Y, player.inLobby.Name)
							}
							continue
						}
						player.head = newHead

						/* New player head */
						newTile := XY{X: newHead.X, Y: newHead.Y}

						/* Check for apple */
						if lobby.showApple && willCollideApple(player) {
							lobbyList[l].showApple = false
							/* Ate an apple, extend one tile */
							player.tiles = append(player.tiles, newHead)
							lobby.grid[newHead] = true

							player.length++
							addScore(player)

							ptype := "Player"
							if player.isBot {
								ptype = "Bot"
							}

							if !player.isBot {
								doLog(true, "%v %v (%v) ate an apple at %v,%v in lobby %v.", ptype, player.name, player.id, player.head.X, player.head.Y, player.inLobby.Name)
							}
						} else {
							/* Otherwise, delete old tail and add new head */
							lobby.grid[player.tiles[0]] = false
							player.tiles = append(player.tiles[1:], newHead)
							lobby.grid[newTile] = true
						}

					}
					lobby.NumPlayers = uint16(playersAlive)

					/* If needed, send keyframe */
					if lobby.dirty {
						outBuf := serializeLobbyBinary(lobby)
						for _, player := range lobby.players {

							if player.isBot || player.conn == nil {
								continue
							}
							if player.inLobby != nil {
								if player.inLobby.ID != lobby.ID {
									continue
								}
							}
							if !writeToPlayer(player, RECV_KEYFRAME, outBuf) {
								player.conn = nil
								doLog(true, "Player.conn write failed, invalidated conn.")
								continue
							}
							numPlayers++
							numBytes += len(outBuf)
						}

						lobby.dirty = false
					} else {
						/* otherwise, just send relevant data*/
						outBuf := binaryGameUpdate(lobby)

						for _, player := range lobby.players {
							if player.isBot || player.conn == nil {
								continue
							}
							if player.inLobby != nil {
								if player.inLobby.ID != lobby.ID {
									continue
								}
							}
							if !writeToPlayer(player, RECV_PLAYERUPDATE, outBuf) {
								player.conn = nil
								doLog(true, "Player.conn write failed, invalidated conn.")
								continue
							}
							numPlayers++
							numBytes += len(outBuf)
						}
					}

					maxRespawn := 1
					/* Respawn ai in dead lobbies */
					if playersAlive < (testPlayers/testlobbies) && totalPlayers > 2 {

						for _, testP := range lobby.players {
							if testP.isBot && testP.length == 0 && maxRespawn > 0 {
								testP.length = 3
								testP.deadFor = -8
								testP.isBot = true
								testP.tiles = nil

								maxRespawn--

								var randx, randy uint8
								for x := 0; x < 10000; x++ {
									randx = uint8(rand.Intn(defaultBoardSize))
									randy = uint8(rand.Intn(defaultBoardSize))
									if !willCollidePlayer(testP.inLobby, testP, testP.direction) {
										break
									}
								}

								tiles := []XY{}
								for x := 0; x < int(testP.length); x++ {
									tiles = append(tiles, XY{X: randx, Y: randy})
								}
								testP.tiles = tiles

								//doLog(true, "Reviving %v (%v) in lobby: %v", testP.name, testP.id, lobby.Name)
							}
						}

					}
					wg.Done()
				}(l)
			}
			wg.Wait()

			if gameTicks%240 == 0 && numBytes > 0 && numPlayers > 0 {
				doLog(true, "Wrote %0.2fkb/sec for %v players.", (float32(numBytes)/1024.0/FrameSpeedMS)*4, numPlayers)
				numBytes = 0
			}

			lobbyLock.Unlock()

			took := time.Since(loopStart)
			remaining := (time.Millisecond * FrameSpeedMS) - took

			if remaining > 0 { /*Kill remaining time*/
				time.Sleep(remaining)

			} else { /*We are lagging behind realtime*/
				time.Sleep(time.Millisecond)
				doLog(true, "Unable to keep up: took: %v", took)
			}
		}
	}()
}

/* Quick and dirty, optimize later */
func spawnApple(lobby *lobbyData) bool {

	limit := int(lobby.boardSize) * int(lobby.boardSize) * 100

	for c := 0; c < limit; c++ {
		rx, ry := uint8(rand.Intn(int(lobby.boardSize-1)))+1, uint8(rand.Intn(int(lobby.boardSize-1))+1)
		for _, player := range lobby.players {
			for _, tile := range player.tiles {
				if tile.X != rx && tile.Y != ry {
					lobby.showApple = true
					lobby.apple = XY{X: rx, Y: ry}
					/*doLog(true, "Spawned apple at %v,%v for lobby %v", rx, ry, lobby.Name)*/
					return true
				}
			}
		}

	}

	return false
}

func willCollideApple(player *playerData) bool {
	if player.head.X == player.inLobby.apple.X &&
		player.head.Y == player.inLobby.apple.Y {
		return true
	}

	return false
}

/* Quick and dirty, optimize later */
func willCollidePlayer(lobby *lobbyData, playerA *playerData, dir DIR) bool {
	/* Player is dead or spawning, ignore */
	if playerA.deadFor != 0 {
		return false
	}

	head := playerA.tiles[playerA.length-1]
	newHead := goDir(dir, head)

	/* Hit someone? */
	return lobby.grid[newHead]
}

func goToApple(ai *playerData) DIR {
	startPos := ai.tiles[ai.length-1]
	endPos := ai.inLobby.apple

	dir := ai.direction

	if ai.dirToggle {
		if startPos.X > endPos.X {
			dir = DIR_WEST
		} else if startPos.X < endPos.X {
			dir = DIR_EAST
		}
		ai.dirToggle = false
	} else {
		if startPos.Y > endPos.Y {
			dir = DIR_NORTH
		} else if startPos.Y < endPos.Y {
			dir = DIR_SOUTH
		}
		ai.dirToggle = true
	}

	if dir == reverseDir(ai.direction) ||
		willCollidePlayer(ai.inLobby, ai, dir) {
		return ai.direction
	}
	return dir
}

func aiMove(ai *playerData) {
	dir := goToApple(ai)

	seed := PosIntMod(int(ai.id), 2)

	head := ai.tiles[ai.length-1]
	newHead := goDir(dir, head)

	/* If we keep going, will we collide with edge or another player? */
	if newHead.X > ai.inLobby.boardSize || newHead.X < 1 ||
		newHead.Y > ai.inLobby.boardSize || newHead.Y < 1 ||
		willCollidePlayer(ai.inLobby, ai, dir) {

		/* Try another direction */
		for x := 0; x <= 5; x++ {

			if seed == 0 {
				dir = DIR(PosIntMod(int(dir+1), 3))
			} else {
				dir = DIR(PosIntMod(int(dir-1), 3))
			}
			if dir == reverseDir(ai.direction) {
				/* don't reverse into self */
				continue
			}
			newHead = goDir(dir, head)

			/* Check if we are good */
			if newHead.X > ai.inLobby.boardSize || newHead.X < 1 ||
				newHead.Y > ai.inLobby.boardSize || newHead.Y < 1 ||
				willCollidePlayer(ai.inLobby, ai, dir) {
				/* Nope, try again */
				continue
			} else {

				/* Good, proceed */
				break
			}

		}
	}

	ai.direction = dir
}

/* Rotate DIR value clockwise */
func RotCW(dir DIR) uint8 {
	return uint8(PosIntMod(int(dir+1), 3))
}

/* Rotate DIR value counter-clockwise */
func RotCCW(dir uint8) uint8 {

	return uint8(PosIntMod(int(dir-1), 3))
}

func PosIntMod(d, m int) int {
	var res int = d % m
	if res < 0 && m > 0 {
		return res + m
	}
	return res
}

/* This can be further optimized, once game logic is put into a module both use.*/
func binaryGameUpdate(lobby *lobbyData) []byte {
	var outBuf = new(bytes.Buffer)

	/*Apple position*/
	binary.Write(outBuf, binary.LittleEndian, lobby.showApple)
	binary.Write(outBuf, binary.LittleEndian, lobby.apple.X)
	binary.Write(outBuf, binary.LittleEndian, lobby.apple.Y)

	/*Number of players*/
	binary.Write(outBuf, binary.LittleEndian, uint16(len(lobby.players)))
	for _, player := range lobby.players {
		/*Player Dead For*/
		binary.Write(outBuf, binary.LittleEndian, player.deadFor)

		/*Player Length*/
		binary.Write(outBuf, binary.LittleEndian, player.length)

		for x := uint16(0); x < player.length; x++ {
			/*Tile X/Y*/
			binary.Write(outBuf, binary.LittleEndian, player.tiles[x].X)
			binary.Write(outBuf, binary.LittleEndian, player.tiles[x].Y)
		}
	}

	/*fmt.Printf("data: '%03v'\n", (outBuf.Bytes()))*/
	return outBuf.Bytes()
}

func serializeLobbyBinary(lobby *lobbyData) []byte {
	var outBuf = new(bytes.Buffer)

	nameLen := uint8(len(lobby.Name))
	/*Lobby Name Len*/
	binary.Write(outBuf, binary.LittleEndian, nameLen)
	for x := uint8(0); x < nameLen; x++ {
		/*Lobby Name Character*/
		binary.Write(outBuf, binary.LittleEndian, byte(lobby.Name[x]))
	}

	/*Lobby data*/
	binary.Write(outBuf, binary.LittleEndian, lobby.ID)
	binary.Write(outBuf, binary.LittleEndian, lobby.showApple)
	binary.Write(outBuf, binary.LittleEndian, lobby.apple.X)
	binary.Write(outBuf, binary.LittleEndian, lobby.apple.Y)

	/*Number of players*/
	binary.Write(outBuf, binary.LittleEndian, uint16(len(lobby.players)))
	for _, player := range lobby.players {
		/*Player ID*/
		binary.Write(outBuf, binary.LittleEndian, player.id)

		nameLen := uint16(len(player.name))
		/*Player Name Length*/
		binary.Write(outBuf, binary.LittleEndian, nameLen)
		for x := uint16(0); x < nameLen; x++ {
			/*Player Name Character*/
			binary.Write(outBuf, binary.LittleEndian, byte(player.name[x]))
		}

		/*Player Dead For*/
		binary.Write(outBuf, binary.LittleEndian, player.deadFor)

		/*Player Length*/
		binary.Write(outBuf, binary.LittleEndian, player.length)
		for x := uint16(0); x < player.length; x++ {
			/*Tile position*/
			binary.Write(outBuf, binary.LittleEndian, player.tiles[x].X)
			binary.Write(outBuf, binary.LittleEndian, player.tiles[x].Y)
		}
	}

	/*fmt.Printf("data: '%03v'\n", (outBuf.Bytes()))*/
	return outBuf.Bytes()
}
