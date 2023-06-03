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

func httpsHandler(w http.ResponseWriter, r *http.Request) {

	sessionID := makeID()
	player := playerData{ID: sessionID}

	cwlog.DoLog(true, "Starting read loop.")
	/* Read body */
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		cwlog.DoLog(true, "Error reading request body: %v", err)
		return
	}
	input := string(bytes)

	/* Empty body, silently reject */
	if input == "" {
		return
	}

	/* Send to command parser */
	if !commandParser(input, w, &player) {
		return
	}
}
