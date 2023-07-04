package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{EnableCompression: true}

func gsHandler(w http.ResponseWriter, r *http.Request) {

	c, err := upgrader.Upgrade(w, r, w.Header())
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
	player := &playerData{conn: conn}
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			doLog(true, "Error on connection read: %v", err)
			conn.Close()
			break
		}
		newParser(data, player)
	}
}
