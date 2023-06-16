package main

import (
	"goSnakeServ/cwlog"
	"time"
)

func processLobbies() {
	go func() {
		for {
			start := time.Now()

			lobbyLock.Lock()
			for _, lobby := range lobbyList {
				lobby.Ticks++
				for _, player := range lobby.Players {
					player.lock.Lock()

					/* Ignore, dead or not init */
					if player.Length < 1 {
						player.lock.Unlock()
						continue
					}
					if player.deadFor > 0 {
						cwlog.DoLog(true, "Player %v died.", player.ID)
						player.deadFor++
						player.lock.Unlock()
						continue
					}
					head := player.Tiles[player.Length-1]
					newHead := goDir(player.Direction, head)
					if newHead.X > lobby.boardSize || newHead.X < 1 ||
						newHead.Y > lobby.boardSize || newHead.Y < 1 {
						player.deadFor = 1
						cwlog.DoLog(true, "Player %v #%v died.\n", player.Name, player.ID)
						continue
					}

					player.Tiles = append(player.Tiles[1:], XY{X: newHead.X, Y: newHead.Y})
					player.Head = head
					player.lock.Unlock()
				}
				//network here
			}
			lobbyLock.Unlock()

			took := time.Since(start)
			remaining := (time.Millisecond * 250) - took

			if remaining > 0 { //Kill remaining time
				time.Sleep(remaining)
				cwlog.DoLog(true, "Frame took %v, %v left.", took, remaining)

			} else { //We are lagging behind realtime
				cwlog.DoLog(true, "Unable to keep up: took: %v", took)
			}
		}
	}()
}
