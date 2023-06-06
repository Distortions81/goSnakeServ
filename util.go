package main

import (
	"fmt"
	"goSnakeServ/cwlog"
	"math/rand"
	"sync"
	"time"
)

/* Prune and write DB if dirty */
func backgroundTasks() {
	go func() {
		for {
			time.Sleep(time.Minute * 5)
			writeDB(false)
		}
	}()
	go func() {
		for {
			lobbyLock.Lock()
			var delList []uint64
			for p, player := range players {
				if time.Since(player.LastActive) > MAX_IDLE {
					delList = append(delList, p)
				}
			}
			for _, item := range delList {
				cwlog.DoLog(true, "Deleting %v", item)
				delete(players, item)
			}
			lobbyLock.Unlock()
			time.Sleep(time.Second * 10)
		}
	}()
}

var userIDLock sync.Mutex

func makeUID() uint64 {
	userIDLock.Lock()
	defer userIDLock.Unlock()
	testID := rand.Uint64()

	/* Keep regenerating until id is unique */
	for findID(testID) {
		testID = rand.Uint64()
		cwlog.DoLog(true, "makeUID: Duplicate UID: %v, regenerating.", testID)
	}

	return testID
}

func findID(id uint64) bool {
	for _, player := range players {
		if player.ID == id {
			return true
		}
	}

	return false
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

func genName() string {
	genUsernameLock.Lock()
	defer genUsernameLock.Unlock()

	uniqueNameNum++
	return fmt.Sprintf("Unnamed-%v", uniqueNameNum)
}

func playerNameUnique(input string) bool {
	for _, player := range players {
		if player.Name == input {
			return false
		}
	}
	return true
}

func playerActivity(player *playerData) {
	player.LastActive = time.Now().UTC()
}
