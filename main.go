package main

import (
	"goSnakeServ/cwlog"
	"net/http"
	"runtime/debug"
	"time"
)

/* Http timeout */
const (
	timeOut = 15 * time.Second
	port    = ":8080"
)

func main() {

	cwlog.StartLog()
	cwlog.LogDaemon()
	/* Sleep on exit, to avoid missing log output */
	defer time.Sleep(time.Second)

	/* Limit memory use just in case */
	debug.SetMemoryLimit(1024 * 1024)
	//kb, mb

	/* Read database */
	err := readDB()
	if err != nil {
		cwlog.DoLog(true, "Error loading secrets: %v", err)
	} else {
		cwlog.DoLog(true, "Loaded db.")
	}

	writeDB(true)

	cwlog.DoLog(true, "starting websocket server on port %v", port)
	http.ListenAndServe(port, http.HandlerFunc(httpsHandler))

	cwlog.DoLog(true, "Goodbye.")
}
