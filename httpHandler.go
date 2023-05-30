package main

import (
	"goSnakeServ/cwlog"
	"net"
	"net/http"

	"github.com/gobwas/ws"
)

func httpsHandler(w http.ResponseWriter, r *http.Request) {

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		cwlog.DoLog(true, "httpsHandler: %v", err)
		return
	}

	clientHandle(conn)

}

func clientHandle(conn net.Conn) {
	var input []byte
	_, err := conn.Read(input)
	if err != nil {
		cwlog.DoLog(true, "clientHandle: %v", err)
	}
}
