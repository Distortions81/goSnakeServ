package main

import (
	"goSnakeServ/cwlog"
	"io"
	"net/http"
)

func httpsHandler(w http.ResponseWriter, r *http.Request) {

	/* Read body */
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		cwlog.DoLog(true, "Error reading request body: %v", err)
		return
	}
	input := string(bytes)

	/* Empty body, silently reject */
	if input == "" {
		return
	}

	/* Send to command parser */
	if !commandParser(input, w) {
		return
	}
}
