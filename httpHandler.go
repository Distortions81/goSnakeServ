package main

import (
	"goSnakeServ/cwlog"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}
var connections []*websocket.Conn

func httpsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		cwlog.DoLog(true, "File request: %v", r.RequestURI)
		fileServer.ServeHTTP(w, r)
		return
	} else if r.Method != http.MethodPost {
		/* Anything other than get or post, just silently reject it */
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	go handleConnection(c)
}

func handleConnection(conn *websocket.Conn) {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			cwlog.DoLog(true, "Error on connection read: %v", err)
			conn.Close()
			break
		}
		commandParser(string(data), conn)
	}
}
