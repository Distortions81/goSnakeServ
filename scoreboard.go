package main

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"
)

const (
	scoresFileName  = "scores.json"
	maxScores       = 100
	dbWriteInterval = time.Minute
)

var (
	scoreBoard  = scoreBoardData{Version: "0002"}
	scoreMutex  sync.Mutex
	scoresDirty bool
)

/* Main database struct */
type scoreBoardData struct {
	Version string

	numScores int
	Scores    []scoreItem

	/* used for http cache */
	cache      []byte
	cacheDirty bool
}

type scoreItem struct {
	Name  string `json:"n"`
	Score uint16 `json:"s"`
	Date  int64  `json:"d"`

	IsBot bool `json:"b"`
}

func dbLoop() {
	for {
		time.Sleep(dbWriteInterval)
		if scoresDirty {
			writeScores(false)
		}
	}
}

func timeoutLoop() {
	for {

		time.Sleep(time.Second * 1)

		pListLock.Lock()
		var deleteMeList []uint32
		for _, player := range connPList {
			if time.Since(player.lastPing) > cTimeout {
				doLog(true, "Authorization timed out for #%v.", player.id)

				deleteFromLobby(player)
				killConnection(player.conn, false)
				player.conn = nil
				deleteMeList = append(deleteMeList, player.id)
			}
		}
		for _, del := range deleteMeList {
			delete(connPList, del)
			delete(playerList, del)
		}
		pListLock.Unlock()
	}
}

/*
 * Check if this is a new high score
 * If it is, sort list and truncate
 * Then mark DB dirty
 */
func addScore(player *playerData) {
	scoreMutex.Lock()
	defer scoreMutex.Unlock()

	newScore := scoreItem{Name: player.name, Date: time.Now().UTC().Unix(), Score: player.length, IsBot: player.isBot}

	if scoreBoard.numScores < 100 {
		scoreBoard.Scores = append(scoreBoard.Scores, newScore)
		sortScores()
		scoreBoard.numScores++
		scoresDirty = true
		scoreBoard.cacheDirty = true
		return
	}

	found := false
	for _, item := range scoreBoard.Scores {
		if item.Score < player.length {
			found = true
			break
		}
	}

	if found {
		scoreBoard.Scores = append(scoreBoard.Scores, newScore)
		sortScores()

		//Trancate
		scoreBoard.Scores = scoreBoard.Scores[:100]
		scoreBoard.numScores = len(scoreBoard.Scores)
		scoresDirty = true
		scoreBoard.cacheDirty = true
	}
}

/* Sort scores high to low */
func sortScores() {
	sort.Slice(scoreBoard.Scores[:], func(i, j int) bool {
		return scoreBoard.Scores[i].Score > scoreBoard.Scores[j].Score
	})
}

/* Read database from disk */
func readScores() error {
	scoreMutex.Lock()
	defer scoreMutex.Unlock()

	bytes, err := os.ReadFile(scoresFileName)
	if err != nil {
		doLog(true, "readScores: %v", err)
		return err
	}

	err = json.Unmarshal(bytes, &scoreBoard)
	if err != nil {
		doLog(true, "readScores: %v", err)
	}

	scoreBoard.numScores = len(scoreBoard.Scores)
	scoreBoard.cacheDirty = true

	return nil
}

/* Write database to disk */
func writeScores(force bool) error {
	scoreMutex.Lock()
	defer scoreMutex.Unlock()

	if !scoresDirty && !force {
		return nil
	}

	scoreBoard.numScores = len(scoreBoard.Scores)

	bytes, err := json.Marshal(scoreBoard)
	if err != nil {
		doLog(true, "writeScores: %v", err)
		return err
	}

	err = os.WriteFile(scoresFileName, bytes, 0644)
	if err != nil {
		doLog(true, "writeScores: %v", err)
		return err
	}

	scoresDirty = false
	//doLog(true, "Wrote scores.")
	return nil
}
