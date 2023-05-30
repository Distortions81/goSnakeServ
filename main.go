package main

import (
	"goSnakeServ/cwlog"
	"net/http"
	"runtime/debug"
	"time"
)

/* Http timeout */
const timeOut = 15 * time.Second

var fileServer http.Handler

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

	http.ListenAndServe(":8080", http.HandlerFunc(httpsHandler))

	cwlog.DoLog(true, "Goodbye.")
}
