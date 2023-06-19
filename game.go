package main

import (
	"goSnakeServ/cwlog"
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
			time.Sleep(time.Millisecond)
			lobbyLock.Lock()
			for l, _ := range lobbyList {

				wg.Add()
				go func(l int) {
					var deletePlayer = -1
					lobby := lobbyList[l]
					lobby.lock.Lock()
					defer lobby.lock.Unlock()

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

			took := time.Since(start)
			remaining := (time.Millisecond * 250) - took

			if remaining > 0 { //Kill remaining time
				time.Sleep(remaining)
				//cwlog.DoLog(true, "Frame took %v, %v left.", took, remaining)

			} else { //We are lagging behind realtime
				cwlog.DoLog(true, "Unable to keep up: took: %v", took)
			}

			lobbyLock.Unlock()
		}
	}()
}
