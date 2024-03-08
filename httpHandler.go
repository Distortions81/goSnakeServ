package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{EnableCompression: false}

func gsHandler(w http.ResponseWriter, r *http.Request) {

	c, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	go handleConnection(c)
}

func siteHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/scoreboard" {
		showScoreboard(w)
		return
	}
	fileServer.ServeHTTP(w, r)
}

const scPrefix = "<html><head><title>goSnake Scoreboard</title>"
const colsSetup = "<style>.column{float:left}.left{width:20%}.middle{width:30%}.right{width:50%}.row:after{content:\"\";display:table;clear:both}</style>"
const header = "<script>function autoRefresh(){window.location=window.location.href;}setInterval('autoRefresh()', 5000);</script></head><body bgcolor=black>"
const cols = "<div class=\"row\"><div class=\"column left\" style=\"color:white;font-size:150%%;\"><h2>%v</h2><p>%v</p></div><div class=\"column middle\" style=\"color:white;font-size:150%%;\"><h2>%v</h2><p>%v</p></div><div class=\"column right\" style=\"color:white;font-size:150%%;\"><h2>%v</h2><p>%v</p></div></div>"
const scSuffix = "</p></body></html>"

func showScoreboard(w http.ResponseWriter) {
	scoreMutex.Lock()
	defer scoreMutex.Unlock()

	if scoreBoard.cacheDirty {

		var outBuf []byte
		outBuf = append(outBuf, []byte(scPrefix)...)
		outBuf = append(outBuf, []byte(colsSetup)...)
		outBuf = append(outBuf, []byte(header)...)

		if len(scoreBoard.Scores) > 0 {
			leftStr := ""
			midStr := ""
			rightStr := ""
			for _, item := range scoreBoard.Scores {
				date := time.Unix(item.Date, 0)
				leftStr = leftStr + strconv.FormatUint(uint64(item.Score), 10) + "<br>"
				midStr = midStr + item.Name + "<br>"
				rightStr = rightStr + fmt.Sprintf("%v %v %v<br>", date.Month(), date.Day(), date.Year())
			}
			buf := fmt.Sprintf(cols, "Score", leftStr, "Name", midStr, "Date", rightStr)
			outBuf = append(outBuf, buf...)

		} else {
			outBuf = append(outBuf, []byte("No high scores found.")...)
		}
		outBuf = append(outBuf, []byte(scSuffix)...)

		w.Write(outBuf)
		scoreBoard.cacheDirty = false
		scoreBoard.cache = outBuf
	} else {
		w.Write([]byte(scoreBoard.cache))
	}
}

var numConnections int = 0
var numConnectionsLock sync.Mutex

func handleConnection(conn *websocket.Conn) {
	if conn == nil {
		return
	}
	player := &playerData{conn: conn, lastPing: time.Now()}
	addConnection()
	for {
		_, data, err := conn.ReadMessage()

		if err != nil {
			doLog(true, "Error on connection read: %v", err)

			killConnection(conn, true)
			deleteFromLobby(player)
			player.conn = nil

			pListLock.Lock()
			delete(playerList, player.id)
			delete(connPList, player.id)
			pListLock.Unlock()

			break
		}
		newParser(data, player)
	}
}

func killConnection(conn *websocket.Conn, force bool) {
	if conn != nil {
		err := conn.Close()
		if err == nil || force {
			numConnectionsLock.Lock()
			if numConnections > 0 {
				numConnections--
			}
			numConnectionsLock.Unlock()
		}
		conn = nil
	}
}

func getNumberConnections() int {
	numConnectionsLock.Lock()
	defer numConnectionsLock.Unlock()

	return numConnections
}

func addConnection() {
	numConnectionsLock.Lock()
	numConnections++
	numConnectionsLock.Unlock()
}
