package main

import (
	"goSnakeServ/cwlog"
	"math/rand"
	"runtime"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

var wg sizedwaitgroup.SizedWaitGroup

func processLobbies() {
	wg = sizedwaitgroup.New(runtime.NumCPU())

	go func() {
		for {
			start := time.Now()
			lobbyLock.Lock()
			for l, _ := range lobbyList {

				wg.Add()
				go func(l int) {
					var deletePlayer = -1
					lobby := lobbyList[l]
					lobby.lock.Lock()
					defer lobby.lock.Unlock()

					if !lobby.ShowApple {
						spawnApple(lobby)
					}

					lobby.Ticks++

					lobbyList[l].outBuf = nil
					for p, player := range lobby.Players {
						defer func() { lobbyList[l].outBuf = append(lobbyList[l].outBuf, byte(player.Direction)) }()

						/* Ignore, dead or not init */
						if player.Length < 1 {
							continue
						}
						if player.DeadFor > 0 {
							if player.DeadFor > 4 {
								deletePlayer = p
							}
							if player.DeadFor == 1 {
								cwlog.DoLog(true, "Player %v died.", player.ID)
							}
							player.DeadFor++
							continue
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
							cwlog.DoLog(true, "Player %v #%v died.\n", player.Name, player.ID)
							continue
						}

						if lobby.ShowApple && didCollideApple(player) {
							lobby.ShowApple = false
							player.Tiles = append(player.Tiles, XY{X: newHead.X, Y: newHead.Y})
							player.Length++
							cwlog.DoLog(true, "Player %v ate an apple.", player.Name)
						} else {
							player.Tiles = append(player.Tiles[1:], XY{X: newHead.X, Y: newHead.Y})
						}
						player.Head = head
					}
					if deletePlayer > -1 {
						cwlog.DoLog(true, "Player %v #%v deleted.\n", lobby.Players[deletePlayer].Name, lobby.Players[deletePlayer].ID)
						lobby.Players = append(lobby.Players[:deletePlayer], lobby.Players[deletePlayer+1:]...)
					}
					wg.Done()
				}(l)
			}
			wg.Wait()

			lobbyLock.Unlock()

			took := time.Since(start)
			remaining := (time.Millisecond * 250) - took

			if remaining > 0 { //Kill remaining time
				time.Sleep(remaining)
				//cwlog.DoLog(true, "Frame took %v, %v left.", took, remaining)

			} else { //We are lagging behind realtime
				cwlog.DoLog(true, "Unable to keep up: took: %v", took)
			}

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

	limit := int(lobby.boardSize*lobby.boardSize) * 10

	for c := 0; c < limit; c++ {
		rx, ry := uint16(rand.Intn(int(lobby.boardSize))), uint16(rand.Intn(int(lobby.boardSize)))
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
		//Skip self
		if playerA.ID == playerB.ID {
			continue
		}
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
	if !ai.isBot || ai.inLobby == nil {
		return
	}

	dir := ai.Direction
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
