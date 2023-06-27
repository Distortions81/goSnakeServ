package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"goSnakeServ/cwlog"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/Distortions81/namegenerator"
)

var gen namegenerator.Generator

func init() {
	gen = namegenerator.NewNameGenerator(rand.Int63())
}

func reverseDir(dir uint8) uint8 {
	switch dir {
	case DIR_NORTH:
		return DIR_SOUTH
	case DIR_EAST:
		return DIR_WEST
	case DIR_SOUTH:
		return DIR_NORTH
	case DIR_WEST:
		return DIR_EAST
	}
	return dir
}

/* Generic unzip []byte */
func UncompressZip(data []byte) []byte {
	b := bytes.NewReader(data)

	z, _ := zlib.NewReader(b)
	defer z.Close()

	p, err := io.ReadAll(z)
	if err != nil {
		return nil
	}
	return p
}

/* Generic zip []byte */
func CompressZip(data []byte) []byte {
	var b bytes.Buffer
	w, _ := zlib.NewWriterLevel(&b, zlib.BestSpeed)
	w.Write(data)
	w.Close()
	return b.Bytes()
}

/* Prune and write DB if dirty */
func backgroundTasks() {

	go func() {
		for {
			pListLock.Lock()
			for _, player := range pList {
				if time.Since(player.lastActive) > MAX_IDLE {
					//killPlayer(player.ID)
				}
			}
			pListLock.Unlock()
			time.Sleep(time.Second * 5)
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

func autoStartDir(player *playerData) uint8 {
	if player != nil && player.inLobby != nil {
		hbsize := player.inLobby.boardSize / 2

		if player.Head.Y < (hbsize) {
			player.Direction = DIR_NORTH
		} else {
			player.Direction = DIR_SOUTH
		}

		if player.Head.X < (hbsize) {
			player.Direction = DIR_EAST
		} else {
			player.Direction = DIR_WEST
		}

	}
	return DIR_SOUTH
}
