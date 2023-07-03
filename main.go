package main

import (
	"crypto/tls"
	"goSnakeServ/cwlog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/Distortions81/namegenerator"
)

/* Http timeout */
const (
	timeOut = 15 * time.Second
	port    = ":8080"
)

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

	cwlog.DoLog(true, "Max random names: %v\n", namegenerator.GetMaxNames())

	processLobbies()

	/* Load certificates */
	cert, err := tls.LoadX509KeyPair("fullchain.pem", "privkey.pem")
	if err != nil {
		cwlog.DoLog(true, "Error loading TLS key pair: %v (fullchain.pem, privkey.pem)", err)
		return
	}
	cwlog.DoLog(true, "Loaded certs.")

	/* Download server */
	fileServer = http.FileServer(http.Dir("www"))

	/* HTTPS server */
	http.HandleFunc("/gs", gsHandler)
	http.HandleFunc("/", siteHandler)

	/* Create TLS configuration */
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: false,
	}

	/* Create HTTPS server */
	server := &http.Server{
		Addr:         port,
		Handler:      http.DefaultServeMux,
		TLSConfig:    config,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),

		ReadTimeout:  timeOut,
		WriteTimeout: timeOut,
		IdleTimeout:  timeOut,
	}

	// Start server
	cwlog.DoLog(true, "Starting server...")
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		cwlog.DoLog(true, "ListenAndServeTLS: %v", err)
		panic(err)
	}

	cwlog.DoLog(true, "Goodbye.")
}
