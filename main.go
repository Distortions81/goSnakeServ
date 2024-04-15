package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/Distortions81/namegenerator"
)

var fileServer http.Handler

func main() {

	devMode := flag.Bool("dev", false, "dev mode enable")
	bindIP := flag.String("ip", "", "IP to bind to")
	bindPort := flag.Int("port", 443, "port to bind to for HTTPS")

	flag.Parse()

	startLog()
	logDaemon()

	/* Sleep on exit, to avoid missing log output */
	defer time.Sleep(time.Second)

	/* Limit memory use just in case */
	debug.SetMemoryLimit(1024 * 1024)
	/*kb, mb*/

	/* Read database */
	err := readScores()
	if err != nil {
		doLog(true, "Error loading secrets: %v", err)
	} else {
		doLog(true, "Loaded db.")
	}

	writeScores(true)

	doLog(true, "Max random names: %v", namegenerator.GetMaxNames())

	processLobbies()

	go dbLoop()
	go timeoutLoop()

	/* Download server */
	fileServer = http.FileServer(http.Dir("www"))

	/* Create HTTPS server */
	server := &http.Server{}
	server.Addr = fmt.Sprintf("%v:%v", *bindIP, *bindPort)

	/* HTTPS server */
	http.HandleFunc("/gs", gsHandler)
	http.HandleFunc("/", siteHandler)

	if !*devMode {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if !*devMode && origin != "https://gosnake.go-game.net" {
				doLog(true, "Connection failed origin check: %v", r.RemoteAddr)
				return false
			}
			return true
		}
	}

	go func() {
		for {
			time.Sleep(time.Second * 5)

			filePath := "fullchain.pem"
			initialStat, erra := os.Stat(filePath)

			if erra != nil {
				continue
			}

			for initialStat != nil {
				time.Sleep(time.Second * 5)

				stat, errb := os.Stat(filePath)
				if errb != nil {
					break
				}

				if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
					doLog(true, "Cert updated, closing.")
					time.Sleep(time.Second * 5)
					os.Exit(0)
					break
				}
			}

		}
	}()

	/* Start server*/
	doLog(true, "Starting server...")
	err = server.ListenAndServeTLS("fullchain.pem", "privkey.pem")
	if err != nil {
		doLog(true, "ListenAndServeTLS: %v", err)
		panic(err)
	}

	doLog(true, "Goodbye.")
}
