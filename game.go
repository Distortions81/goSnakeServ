package main

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"runtime"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

const FrameSpeed = 250

func processLobbies() {
	var numBytes int
	var gameTicks int

	go func() {
		wg := sizedwaitgroup.New(runtime.NumCPU())

		for {
			loopStart := time.Now()
			gameTicks++

			var numPlayers int

			lobbyLock.Lock()
			for l := range lobbyList {

				wg.Add()
				go func(l int) {
					lobbyList[l].lock.Lock()
					defer lobbyList[l].lock.Unlock()

					lobby := lobbyList[l]

					if !lobby.showApple {
						spawnApple(lobby)
					}

					playersAlive := 0
					totalPlayers := 0

					for _, player := range lobby.Players {
						totalPlayers++

						/* skip anyone not in this lobby */
						if !player.isBot {
							if player.inLobby != nil {
								if player.inLobby.ID != lobby.ID {
									continue
								}
							}
						}

						/* Ignore, dead or not init */
						if player.length < 1 {
							continue
						}
						if player.DeadFor > 0 {
							if player.DeadFor > 4 {
								player.tiles = []XY{{}}
								player.length = 0
								continue
							}
							player.DeadFor++
							continue
						} else if player.DeadFor < 0 {
							player.DeadFor++
							continue
						}
						playersAlive++

						ptype := "Player"
						if player.isBot {
							ptype = "Bot"
						}

						/* Test basic AI */
						if player.isBot {
							aiMove(player)
						}

						head := player.tiles[player.length-1]
						if player.numDirs > 0 {
							newDir := player.dirs[0]

							if player.numDirs > 1 {
								player.dirs = append(player.dirs[:0], player.dirs[1:]...)
							} else {
								player.dirs = nil
							}
							player.numDirs--

							if newDir != reverseDir(player.oldDir) {
								player.direction = newDir
							}
							player.oldDir = player.direction
						}

						newHead := goDir(player.direction, head)
						if newHead.X > lobby.boardSize || newHead.X < 1 ||
							newHead.Y > lobby.boardSize || newHead.Y < 1 || willCollidePlayer(player.inLobby, player, player.direction) {
							player.DeadFor = 1
							if !player.isBot {
								doLog(true, "%v %v #%v died at %v,%v in lobby %v.", ptype, player.Name, player.id, player.head.X, player.head.Y, player.inLobby.Name)
							}
							continue
						}
						player.head = newHead

						if lobby.showApple && didCollideApple(player) {
							lobbyList[l].showApple = false
							player.tiles = append(player.tiles, XY{X: newHead.X, Y: newHead.Y})
							player.length++
							if !player.isBot {
								doLog(true, "%v %v ate an apple at %v,%v in lobby %v.", ptype, player.Name, player.head.X, player.head.Y, player.inLobby.Name)
							}
						} else {
							player.tiles = append(player.tiles[1:], XY{X: newHead.X, Y: newHead.Y})
						}

					}

					/* If needed, send keyframe */
					if lobby.dirty {
						outBuf := serializeLobbyBinary(lobby)

						for _, player := range lobby.Players {

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

						for _, player := range lobby.Players {
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
					/* Respawn players in dead lobbies */
					if playersAlive <= 2 && totalPlayers > 2 {
						//doLog(true, "Reviving AIs in lobby: %v", lobby.Name)
						for _, testP := range lobby.Players {
							if testP.isBot && testP.length == 0 && maxRespawn > 0 {
								testP.length = 3
								testP.DeadFor = -8
								testP.isBot = true

								maxRespawn--

								var randx, randy uint8
								for x := 0; x < 10000; x++ {
									randx = uint8(rand.Intn(defaultBoardSize))
									randy = uint8(rand.Intn(defaultBoardSize))
									if !didCollidePlayer(testP.inLobby, testP) {
										break
									}
								}

								tiles := []XY{}
								for x := 0; x < int(testP.length); x++ {
									tiles = append(tiles, XY{X: randx, Y: randy})
								}
								testP.tiles = tiles
							}
						}
					}
					wg.Done()
				}(l)
			}
			wg.Wait()

			lobbyLock.Unlock()

			if gameTicks%240 == 0 && numBytes > 0 && numPlayers > 0 {
				doLog(true, "Wrote %0.2fkb/sec for %v players.", (float32(numBytes)/1024.0/240.0)*4, numPlayers)
				numBytes = 0
			}

			took := time.Since(loopStart)
			remaining := (time.Millisecond * FrameSpeed) - took

			if remaining > 0 { //Kill remaining time
				time.Sleep(remaining)

			} else { //We are lagging behind realtime
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
		for _, player := range lobby.Players {
			for _, tile := range player.tiles {
				if tile.X != rx && tile.Y != ry {
					lobby.showApple = true
					lobby.apple = XY{X: rx, Y: ry}
					//doLog(true, "Spawned apple at %v,%v for lobby %v", rx, ry, lobby.Name)
					return true
				}
			}
		}

	}

	return false
}

func didCollideApple(player *playerData) bool {
	if player.inLobby == nil {
		return false
	}
	for _, tile := range player.tiles {
		if tile.X == player.inLobby.apple.X &&
			tile.Y == player.inLobby.apple.Y {
			return true
		}
	}
	return false
}

/* Quick and dirty, optimize later */
func didCollidePlayer(lobby *lobbyData, playerA *playerData) bool {
	if playerA.DeadFor < 0 {
		playerA.DeadFor = -8
		return false
	}
	for _, playerB := range lobby.Players {
		if playerB.DeadFor < 0 {
			playerA.DeadFor = -8
			return false
		}
		for _, tileA := range playerA.tiles {
			for _, tileB := range playerB.tiles {
				if tileA.X == tileB.X && tileA.Y == tileB.Y {
					return true
				}
			}
		}
	}

	return false
}

/* Quick and dirty, optimize later */
func willCollidePlayer(lobby *lobbyData, playerA *playerData, dir uint8) bool {
	if playerA.DeadFor != 0 {
		return false
	}

	head := playerA.tiles[playerA.length-1]
	newHead := goDir(dir, head)

	for _, playerB := range lobby.Players {
		if playerB.DeadFor > 0 {
			continue
		}
		for b, tileB := range playerB.tiles {
			if tileB.X == newHead.X && tileB.Y == newHead.Y {
				if !playerA.isBot && playerA.id == playerB.id {
					doLog(true, "%v hit themself at %v,%v, going %v, position: %v, tiles: %v", playerA.Name, newHead.X, newHead.Y, dirToString(playerA.direction), b, playerA.tiles)
				}
				return true
			}
		}
	}

	return false
}

func aiMove(ai *playerData) {
	if !ai.isBot || ai.inLobby == nil || ai.length < 1 {
		return
	}

	dir := ai.direction
	if rand.Intn(15) == 0 {
		dir = uint8(rand.Intn(DIR_WEST + 1)) /* New test */
		ai.direction = dir
	}

	head := ai.tiles[ai.length-1]
	newHead := goDir(dir, head)

	/* If we keep going, will we collide with edge or another player? */
	if newHead.X > ai.inLobby.boardSize || newHead.X < 1 ||
		newHead.Y > ai.inLobby.boardSize || newHead.Y < 1 ||
		willCollidePlayer(ai.inLobby, ai, dir) {

		/* Try another direction */
		for x := 0; x < 256; x++ {
			if x == int(ai.direction) {
				continue
			}

			/* Rotate */
			dir = uint8(rand.Intn(DIR_WEST + 1)) /* New test */
			newHead = goDir(dir, head)

			/* Check if we are good */
			if newHead.X > ai.inLobby.boardSize || newHead.X < 1 ||
				newHead.Y > ai.inLobby.boardSize || newHead.Y < 1 ||
				willCollidePlayer(ai.inLobby, ai, dir) {

				/* Nope, try again */
				continue

			} else {

				/* Good, proceed */
				ai.direction = dir
				break
			}

		}
	}
}

func PosIntMod(d, m int) int {
	var res int = d % m
	if res < 0 && m > 0 {
		return res + m
	}
	return res
}

// This can be further optimized, once game logic is put into a module both use.
func binaryGameUpdate(lobby *lobbyData) []byte {
	var outBuf = new(bytes.Buffer)

	//Apple position
	binary.Write(outBuf, binary.BigEndian, lobby.showApple)
	binary.Write(outBuf, binary.BigEndian, lobby.apple.X)
	binary.Write(outBuf, binary.BigEndian, lobby.apple.Y)

	//Number of players
	binary.Write(outBuf, binary.BigEndian, uint16(len(lobby.Players)))
	for _, player := range lobby.Players {
		//Player Dead For
		binary.Write(outBuf, binary.BigEndian, player.DeadFor)

		//Player Length
		binary.Write(outBuf, binary.BigEndian, player.length)

		for x := uint16(0); x < player.length; x++ {
			//Tile X/Y
			binary.Write(outBuf, binary.BigEndian, player.tiles[x].X)
			binary.Write(outBuf, binary.BigEndian, player.tiles[x].Y)
		}
	}

	//fmt.Printf("data: '%03v'\n", (outBuf.Bytes()))
	return outBuf.Bytes()
}

func serializeLobbyBinary(lobby *lobbyData) []byte {
	var outBuf = new(bytes.Buffer)

	nameLen := uint8(len(lobby.Name))
	//Lobby Name Len
	binary.Write(outBuf, binary.BigEndian, nameLen)
	for x := uint8(0); x < nameLen; x++ {
		//Lobby Name Character
		binary.Write(outBuf, binary.BigEndian, byte(lobby.Name[x]))
	}

	//Lobby data
	binary.Write(outBuf, binary.BigEndian, lobby.ID)
	binary.Write(outBuf, binary.BigEndian, lobby.showApple)
	binary.Write(outBuf, binary.BigEndian, lobby.apple.X)
	binary.Write(outBuf, binary.BigEndian, lobby.apple.Y)

	//Number of players
	binary.Write(outBuf, binary.BigEndian, uint16(len(lobby.Players)))
	for _, player := range lobby.Players {
		//Player ID
		binary.Write(outBuf, binary.BigEndian, player.id)

		nameLen := uint16(len(player.Name))
		//Player Name Length
		binary.Write(outBuf, binary.BigEndian, nameLen)
		for x := uint16(0); x < nameLen; x++ {
			//Player Name Character
			binary.Write(outBuf, binary.BigEndian, byte(player.Name[x]))
		}

		//Player Dead For
		binary.Write(outBuf, binary.BigEndian, player.DeadFor)

		//Player Length
		binary.Write(outBuf, binary.BigEndian, player.length)
		for x := uint16(0); x < player.length; x++ {
			//Tile position
			binary.Write(outBuf, binary.BigEndian, player.tiles[x].X)
			binary.Write(outBuf, binary.BigEndian, player.tiles[x].Y)
		}
	}

	//fmt.Printf("data: '%03v'\n", (outBuf.Bytes()))
	return outBuf.Bytes()
}
