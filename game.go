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

					if !lobby.ShowApple {
						spawnApple(lobby)
					}

					lobby.Ticks++
					playersAlive := 0
					totalPlayers := 0

					for _, player := range lobby.Players {
						totalPlayers++

						/* Ignore, dead or not init */
						if player.Length < 1 {
							continue
						}
						if player.DeadFor > 0 {
							if player.DeadFor > 4 {
								player.Tiles = []XY{{}}
								player.ID = 0
								player.Length = 0
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

						head := player.Tiles[player.Length-1]
						if player.numDirs > 0 {
							newDir := player.dirs[0]

							if player.numDirs > 1 {
								player.dirs = append(player.dirs[:0], player.dirs[1:]...)
							} else {
								player.dirs = nil
							}
							player.numDirs--

							if newDir != reverseDir(player.oldDir) {
								player.Direction = newDir
							}
							player.oldDir = player.Direction
						}

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

					/* If needed, send keyframe */
					if lobby.dirty {
						outBuf := serializeLobbyBinary(lobby)

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

						lobby.dirty = false
					} else {
						/* otherwise, just send relevant data*/
						outBuf := binaryGameUpdate(lobby)

						for _, player := range lobby.Players {
							if player.isBot || player.conn == nil {
								continue
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
						doLog(true, "Reviving AIs in lobby: %v", lobby.Name)
						for _, testP := range lobby.Players {
							if testP.isBot && testP.Length == 0 && maxRespawn > 0 {
								testP.Length = 3
								testP.DeadFor = -8
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
	if playerA.DeadFor < 0 {
		playerA.DeadFor = -8
		return false
	}
	for _, playerB := range lobby.Players {
		if playerB.DeadFor < 0 {
			playerA.DeadFor = -8
			return false
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
		if playerB.DeadFor > 0 {
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

// This can be further optimized, once game logic is put into a module both use.
func binaryGameUpdate(lobby *lobbyData) []byte {
	var outBuf = new(bytes.Buffer)

	binary.Write(outBuf, binary.BigEndian, lobby.Apple.X)
	binary.Write(outBuf, binary.BigEndian, lobby.Apple.Y)

	//Number of players
	binary.Write(outBuf, binary.BigEndian, uint16(len(lobby.Players)))
	for _, player := range lobby.Players {
		//Player Dead For
		binary.Write(outBuf, binary.BigEndian, player.DeadFor)

		//Player Length
		binary.Write(outBuf, binary.BigEndian, player.Length)
		tLen := uint32(len(player.Tiles))
		for x := uint32(0); x < tLen; x++ {
			//Tile X
			binary.Write(outBuf, binary.BigEndian, player.Tiles[x].X)
			//Tile Y
			binary.Write(outBuf, binary.BigEndian, player.Tiles[x].Y)
		}
	}

	//fmt.Printf("data: '%03v'\n", (outBuf.Bytes()))
	return outBuf.Bytes()
}

func serializeLobbyBinary(lobby *lobbyData) []byte {
	var outBuf = new(bytes.Buffer)

	nameLen := uint16(len(lobby.Name))
	//Lobby Name Len
	binary.Write(outBuf, binary.BigEndian, nameLen)
	for x := uint16(0); x < nameLen; x++ {
		//Lobby Name Character
		binary.Write(outBuf, binary.BigEndian, byte(lobby.Name[x]))
	}

	binary.Write(outBuf, binary.BigEndian, lobby.ID)
	binary.Write(outBuf, binary.BigEndian, lobby.Ticks)
	binary.Write(outBuf, binary.BigEndian, lobby.Level)
	binary.Write(outBuf, binary.BigEndian, lobby.ShowApple)
	binary.Write(outBuf, binary.BigEndian, lobby.Apple.X)
	binary.Write(outBuf, binary.BigEndian, lobby.Apple.Y)

	//Number of players
	binary.Write(outBuf, binary.BigEndian, uint16(len(lobby.Players)))
	for _, player := range lobby.Players {
		//Player ID
		binary.Write(outBuf, binary.BigEndian, player.ID)

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
		binary.Write(outBuf, binary.BigEndian, player.Length)
		tLen := uint32(len(player.Tiles))
		for x := uint32(0); x < tLen; x++ {
			//Tile X
			binary.Write(outBuf, binary.BigEndian, player.Tiles[x].X)
			//Tile Y
			binary.Write(outBuf, binary.BigEndian, player.Tiles[x].Y)
		}
	}

	//fmt.Printf("data: '%03v'\n", (outBuf.Bytes()))
	return outBuf.Bytes()
}
