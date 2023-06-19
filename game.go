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

					lobby.Ticks++

					if lobby.Ticks%40 == 0 {
						//Add apple
					}

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
							newHead.Y > lobby.boardSize || newHead.Y < 1 {
							player.DeadFor = 1
							cwlog.DoLog(true, "Player %v #%v died.\n", player.Name, player.ID)
							continue
						}

						player.Tiles = append(player.Tiles[1:], XY{X: newHead.X, Y: newHead.Y})
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

func aiMove(ai *playerData) {

	dir := ai.Direction
	head := ai.Tiles[ai.Length-1]
	newHead := goDir(dir, head)

	/* If we keep going, will we collide with edge? */
	if newHead.X > ai.inLobby.boardSize || newHead.X < 1 ||
		newHead.Y > ai.inLobby.boardSize || newHead.Y < 1 {

		/* Try another direction */
		for x := 0; x < 8; x++ {

			/* Rotate */
			dir = uint8(rand.Intn(4))
			/* New test */
			newHead = goDir(dir, head)

			/* Check if we are good */
			if newHead.X > ai.inLobby.boardSize || newHead.X < 1 ||
				newHead.Y > ai.inLobby.boardSize || newHead.Y < 1 {

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
