package main

import (
	"goSnakeServ/cwlog"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func httpsHandler(w http.ResponseWriter, r *http.Request) {

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		cwlog.DoLog(true, "httpsHandler: %v", err)
		return
	}

	go clientHandle(conn)

}

func clientHandle(conn net.Conn) {
	defer conn.Close()

	for {
		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil {
			cwlog.DoLog(true, "clientHandle: %v", err)
			return
		}
		err = wsutil.WriteServerMessage(conn, op, msg)
		if err != nil {
			cwlog.DoLog(true, "clientHandle: %v", err)
			return
		}
	}
}
