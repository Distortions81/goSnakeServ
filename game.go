package main

import (
	"encoding/json"
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

					if !lobby.ShowApple {
						spawnApple(lobby)
					}

					lobby.Ticks++
					playersAlive := 0

					for _, player := range lobby.Players {
						/* Ignore, dead or not init */
						if player.Length < 1 {
							continue
						}
						if player.DeadFor > 0 {
							if player.DeadFor > 4 {
								player.Tiles = []XY{{}}
								player.Length = 0
								continue
							}
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

						head := player.Tiles[player.Length-1]
						newHead := goDir(player.Direction, head)
						if newHead.X > lobby.boardSize || newHead.X < 1 ||
							newHead.Y > lobby.boardSize || newHead.Y < 1 || willCollidePlayer(player.inLobby, player, player.Direction) {
							player.DeadFor = 1
							if !player.isBot {
								doLog(true, "%v %v #%v died at %v,%v in lobby %v.\n", ptype, player.Name, player.ID, player.Head.X, player.Head.Y, player.inLobby.Name)
							}
							continue
						}

						if lobby.ShowApple && didCollideApple(player) {
							lobby.ShowApple = false
							player.Tiles = append(player.Tiles, XY{X: newHead.X, Y: newHead.Y})
							player.Length++
							if !player.isBot {
								doLog(true, "%v %v ate an apple at %v,%v in lobby %v.", ptype, player.Name, player.Head.X, player.Head.Y, player.inLobby.Name)
							}
						} else {
							player.Tiles = append(player.Tiles[1:], XY{X: newHead.X, Y: newHead.Y})
						}
						player.Head = head

					}
					outBuf, _ := json.Marshal(&lobby)

					for _, player := range lobby.Players {
						if player.isBot || player.conn == nil {
							continue
						}
						if !writeToPlayer(player, RECV_KEYFRAME, outBuf) {
							player.conn = nil
							doLog(true, "Player.conn write failed, invalidated conn.")
							continue
						}
						numPlayers++
						numBytes += len(outBuf)
					}

					maxRespawn := 10
					/* Respawn players in dead lobbies */
					if playersAlive < 5 {
						doLog(true, "Reviving AIs in lobby: %v", lobby.Name)
						for _, testP := range lobby.Players {
							if testP.isBot && testP.Length == 0 && maxRespawn > 0 {
								testP.Length = 3
								testP.DeadFor = 0
								testP.isBot = true

								maxRespawn--

								var randx, randy uint16
								for x := 0; x < 10000; x++ {
									randx = uint16(rand.Intn(defaultBoardSize))
									randy = uint16(rand.Intn(defaultBoardSize))
									if !didCollidePlayer(testP.inLobby, testP) {
										break
									}
								}

								tiles := []XY{}
								for x := 0; x < int(testP.Length); x++ {
									tiles = append(tiles, XY{X: randx, Y: randy})
								}
								testP.Tiles = tiles
							}
						}
					}
					wg.Done()
				}(l)
			}
			wg.Wait()

			lobbyLock.Unlock()

			if gameTicks%240 == 0 && numBytes > 0 && numPlayers > 0 {
				doLog(true, "Wrote %0.2fkb/sec for %v players.", float32(numBytes)/1024.0/240.0, numPlayers)
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

	limit := int(lobby.boardSize*lobby.boardSize) * 100

	for c := 0; c < limit; c++ {
		rx, ry := uint16(rand.Intn(int(lobby.boardSize-1)))+1, uint16(rand.Intn(int(lobby.boardSize-1))+1)
		for _, player := range lobby.Players {
			for _, tile := range player.Tiles {
				if tile.X != rx && tile.Y != ry {
					lobby.ShowApple = true
					lobby.Apple = XY{X: rx, Y: ry}
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
	for _, tile := range player.Tiles {
		if tile.X == player.inLobby.Apple.X &&
			tile.Y == player.inLobby.Apple.Y {
			return true
		}
	}
	return false
}

/* Quick and dirty, optimize later */
func didCollidePlayer(lobby *lobbyData, playerA *playerData) bool {
	for _, playerB := range lobby.Players {
		for _, tileA := range playerA.Tiles {
			for _, tileB := range playerB.Tiles {
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

	head := playerA.Tiles[playerA.Length-1]
	newHead := goDir(dir, head)

	for _, playerB := range lobby.Players {
		if playerB.DeadFor != 0 {
			continue
		}
		for _, tileA := range playerB.Tiles {
			if tileA.X == newHead.X && tileA.Y == newHead.Y {
				return true
			}
		}
	}

	return false
}

func aiMove(ai *playerData) {
	if !ai.isBot || ai.inLobby == nil || ai.Length < 1 {
		return
	}

	dir := ai.Direction
	if rand.Intn(15) == 0 {
		dir = uint8(rand.Intn(DIR_WEST + 1)) /* New test */
		ai.Direction = dir
	}

	head := ai.Tiles[ai.Length-1]
	newHead := goDir(dir, head)

	/* If we keep going, will we collide with edge or another player? */
	if newHead.X > ai.inLobby.boardSize || newHead.X < 1 ||
		newHead.Y > ai.inLobby.boardSize || newHead.Y < 1 ||
		willCollidePlayer(ai.inLobby, ai, dir) {

		/* Try another direction */
		for x := 0; x < 256; x++ {
			if x == int(ai.Direction) {
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
				ai.Direction = dir
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
