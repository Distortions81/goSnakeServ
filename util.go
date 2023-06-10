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
