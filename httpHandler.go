package main

import (
	"goSnakeServ/cwlog"
	"io"
	"net/http"
	"sync"
	"time"
)

var clientIDLock sync.Mutex

func makeID() int64 {
	clientIDLock.Lock()
	defer clientIDLock.Unlock()

	return time.Now().UnixNano()
}

const readSleep = time.Millisecond * 100

func httpsHandler(w http.ResponseWriter, r *http.Request) {

	sessionID := makeID()
	player := playerData{ID: sessionID}

	/* Incoming get? Send to file server */
	if r.Method != http.MethodPost {
		/* Anything other than get or post, just silently reject it */
		return
	}
	cwlog.DoLog(true, "Starting read loop.")

	for {
		/* Read body */
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			cwlog.DoLog(true, "Error reading request body: %v", err)
			return
		}
		input := string(bytes)

		/* Empty body, silently reject */
		if input == "" {
			continue
		}

		/* Send to command parser */
		if !commandParser(input, w, &player) {
			return
		}

		time.Sleep(readSleep)
	}
}
