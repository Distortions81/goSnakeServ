package main

import (
	"goSnakeServ/cwlog"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

var wg sizedwaitgroup.SizedWaitGroup
var netLock sync.Mutex

const FrameSpeed = 249
const NetTime = FrameSpeed / 10
const FrameTime = FrameSpeed - NetTime

func processLobbies() {
	wg = sizedwaitgroup.New(runtime.NumCPU())

	go func() {
		for {
			loopStart := time.Now()
			time.Sleep(time.Millisecond)

			lobbyLock.Lock()
			netLock.Lock()
			for l, _ := range lobbyList {

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

					//lobbyList[l].outBuf = nil
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
								cwlog.DoLog(true, "%v %v #%v died at %v,%v in lobby %v.\n", ptype, player.Name, player.ID, player.Head.X, player.Head.Y, player.inLobby.Name)
							}
							continue
						}

						if lobby.ShowApple && didCollideApple(player) {
							lobby.ShowApple = false
							player.Tiles = append(player.Tiles, XY{X: newHead.X, Y: newHead.Y})
							player.Length++
							if !player.isBot {
								cwlog.DoLog(true, "%v %v ate an apple at %v,%v in lobby %v.", ptype, player.Name, player.Head.X, player.Head.Y, player.inLobby.Name)
							}
						} else {
							player.Tiles = append(player.Tiles[1:], XY{X: newHead.X, Y: newHead.Y})
						}
						player.Head = head
					}
					maxRespawn := 10
					/* Respawn players in dead lobbies */
					if playersAlive < 5 {
						cwlog.DoLog(true, "Reviving AIs in lobby: %v", lobby.Name)
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

			took := time.Since(loopStart)
			remaining := (time.Millisecond * FrameTime) - took

			if remaining > 0 { //Kill remaining time
				time.Sleep(remaining)
				//cwlog.DoLog(true, "Frame took %v, %v left.", took, remaining)

			} else { //We are lagging behind realtime
				cwlog.DoLog(true, "Unable to keep up: took: %v", took)
			}

			netLock.Unlock()
			time.Sleep(time.Millisecond * NetTime)
		}
	}()
}

func rotateCW(dir uint8) uint8 {
	return uint8(PosIntMod(int(dir+1), DIR_WEST))
}

func rotateCCW(dir uint8) uint8 {
	return uint8(PosIntMod(int(dir-1), DIR_WEST))
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
