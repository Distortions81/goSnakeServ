package main

import (
	"goSnakeServ/cwlog"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func gsHandler(w http.ResponseWriter, r *http.Request) {

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	go handleConnection(c)
}

func siteHandler(w http.ResponseWriter, r *http.Request) {
	fileServer.ServeHTTP(w, r)
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
