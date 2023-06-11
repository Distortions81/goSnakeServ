package main

import (
	"fmt"
	"goSnakeServ/cwlog"
	"math/rand"
	"sync"
	"time"

	"github.com/Distortions81/namegenerator"
)

var gen namegenerator.Generator

func init() {
	gen = namegenerator.NewNameGenerator(rand.Int63())
}

/* Prune and write DB if dirty */
func backgroundTasks() {

	go func() {
		for {
			pListLock.Lock()
			for _, player := range pList {
				if time.Since(player.lastActive) > MAX_IDLE {
					killPlayer(player.ID)
				}
			}
			pListLock.Unlock()
			time.Sleep(time.Second * 5)
		}
	}()
}

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
						//cwlog.DoLog(true, "Player %v length under 1.", player.ID)
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

func goDir(dir uint8, pos XY) XY {
	switch dir {
	case DIR_NORTH:
		pos.Y--
	case DIR_EAST:
		pos.X++
	case DIR_SOUTH:
		pos.Y++
	case DIR_WEST:
		pos.X--
	}
	return pos
}

func killPlayer(id uint64) {
	player := pList[id]

	//Remove from lobby
	if player.inLobby != nil {
		for p, player := range player.inLobby.Players {
			if player.ID == id {
				player.inLobby.Players = append(player.inLobby.Players[:p], player.inLobby.Players[p+1:]...)
				return
			}
		}
	}
	//Remove from personal lobby
	if player.myLobby != nil {
		for p, player := range player.myLobby.Players {
			if player.ID == id {
				player.inLobby.Players = append(player.inLobby.Players[:p], player.inLobby.Players[p+1:]...)
				return
			}
		}
	}
	delete(pList, id)
}

var UIDLock sync.Mutex

func makeUID() uint64 {
	UIDLock.Lock()
	defer UIDLock.Unlock()
	testID := rand.Uint64()

	/* Keep regenerating until id is unique */
	for findID(testID) {
		testID = rand.Uint64()
		cwlog.DoLog(true, "makeUID: Duplicate UID: %v, regenerating.", testID)
	}

	return testID
}

func findID(id uint64) bool {
	if pList[id] == nil {
		return false
	} else {
		return true
	}
}

func filterName(input string) string {
	buf := StripControlAndSpecial(input)
	buf = TruncateString(buf, 64)
	iLen := len(buf)

	if iLen < 2 {
		buf = genName()
	}
	return buf
}

var genUsernameLock sync.Mutex
var uniqueNameNum uint64
var outOfNames bool

func genName() string {
	genUsernameLock.Lock()
	defer genUsernameLock.Unlock()

	if !outOfNames {
		for x := 0; x < 10; x++ {
			name := gen.Generate()
			if playerNameUnique(name) {
				return name
			}
			cwlog.DoLog(true, "Regenerating, name dupe: %v", name)
		}
		outOfNames = true
	}

	/* Fallback */
	uniqueNameNum++
	return fmt.Sprintf("Unnamed-%v", uniqueNameNum)
}

func playerNameUnique(input string) bool {
	for _, player := range pList {
		if player.Name == input {
			return false
		}
	}
	return true
}

func playerActivity(player *playerData) {
	player.lastActive = time.Now()
}
