package main

import (
	"crypto/tls"
	"flag"
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

	devMode := flag.Bool("dev", false, "dev mode enable")
	flag.Parse()
	if !*devMode {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == "https://facility38.xyz:8080"
		}
	}

	startLog()
	logDaemon()

	/* Sleep on exit, to avoid missing log output */
	defer time.Sleep(time.Second)

	/* Limit memory use just in case */
	debug.SetMemoryLimit(1024 * 1024)
	//kb, mb

	/* Read database */
	err := readDB()
	if err != nil {
		doLog(true, "Error loading secrets: %v", err)
	} else {
		doLog(true, "Loaded db.")
	}

	writeDB(true)

	doLog(true, "Max random names: %v", namegenerator.GetMaxNames())

	processLobbies()

	/* Load certificates */
	cert, err := tls.LoadX509KeyPair("fullchain.pem", "privkey.pem")
	if err != nil {
		doLog(true, "Error loading TLS key pair: %v (fullchain.pem, privkey.pem)", err)
		return
	}
	doLog(true, "Loaded certs.")

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
	doLog(true, "Starting server...")
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		doLog(true, "ListenAndServeTLS: %v", err)
		panic(err)
	}

	doLog(true, "Goodbye.")
}
