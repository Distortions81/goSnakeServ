package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
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
	w, _ := zlib.NewWriterLevel(&b, zlib.BestCompression)
	w.Write(data)
	w.Close()
	return b.Bytes()
}

func uint64ToByteArray(i uint64) []byte {
	byteArray := make([]byte, 8)
	binary.LittleEndian.PutUint64(byteArray, i)
	return byteArray
}

func uint32ToByteArray(i uint32) []byte {
	byteArray := make([]byte, 4)
	binary.LittleEndian.PutUint32(byteArray, i)
	return byteArray
}

func uint16ToByteArray(i uint16) []byte {
	byteArray := make([]byte, 2)
	binary.LittleEndian.PutUint16(byteArray, i)
	return byteArray
}

func uint8ToByteArray(i uint8) []byte {
	byteArray := make([]byte, 1)
	byteArray[0] = byte(i)
	return byteArray
}

func byteArrayToUint8(i []byte) uint8 {
	if len(i) < 1 {
		return 0
	}
	return uint8(i[0])
}

func byteArrayToUint16(i []byte) uint16 {
	if len(i) < 2 {
		return 0
	}
	return binary.LittleEndian.Uint16(i)
}

func byteArrayToUint32(i []byte) uint32 {
	if len(i) < 4 {
		return 0
	}
	return binary.LittleEndian.Uint32(i)
}

func byteArrayToUint64(i []byte) uint64 {
	if len(i) < 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(i)
}

func reverseDir(dir DIR) DIR {
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

func goDir(dir DIR, pos XY) XY {
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

var playerTop uint32 = 1

func makePlayerUID() uint32 {
	UIDLock.Lock()
	defer UIDLock.Unlock()

	playerTop++
	return playerTop
}

var lobbyTop uint16

func makeLobbyUID() uint16 {
	UIDLock.Lock()
	defer UIDLock.Unlock()

	lobbyTop++
	return lobbyTop
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
var uniqueNameNum uint32
var outOfNames bool

func genName() string {
	genUsernameLock.Lock()
	defer genUsernameLock.Unlock()

	if !outOfNames {
		for x := 0; x < 4; x++ {
			name := gen.Generate()
			if playerNameUnique(name) {
				return name
			}
			doLog(true, "Regenerating, name dupe: %v", name)
		}
		outOfNames = true
	}

	/* Fallback */
	uniqueNameNum++
	doLog(true, "Giving generic name instead: %v", uniqueNameNum)
	return fmt.Sprintf("Unnamed-%v", uniqueNameNum)
}

func playerNameUnique(input string) bool {
	pListLock.Lock()
	defer pListLock.Unlock()

	for _, player := range playerList {
		if player.name == input {
			return false
		}
	}
	return true
}

func playerActivity(player *playerData) {
	player.lastActive = time.Now()
}

func autoStartDir(player *playerData) DIR {
	if player != nil && player.inLobby != nil {
		hbsize := player.inLobby.boardSize / 2

		if player.head.Y < (hbsize) {
			player.direction = DIR_NORTH
		} else {
			player.direction = DIR_SOUTH
		}

		if player.head.X < (hbsize) {
			player.direction = DIR_EAST
		} else {
			player.direction = DIR_WEST
		}

	}
	return DIR_SOUTH
}

/* Replace me */
func dirToString(dir DIR) string {
	if dir == DIR_NORTH {
		return "north"
	} else if dir == DIR_EAST {
		return "east"
	} else if dir == DIR_SOUTH {
		return "south"
	} else if dir == DIR_WEST {
		return "west"
	} else {
		return "???"
	}
}
