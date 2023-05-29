package main

import (
	"F38Auth/cwlog"
	"io"
	"net/http"
)

func httpsHandler(w http.ResponseWriter, r *http.Request) {

	/* Incoming get? Send to file server */
	if r.Method == http.MethodGet {
		cwlog.DoLog(true, "File request: %v", r.RequestURI)
		fileServer.ServeHTTP(w, r)
		return
	} else if r.Method != http.MethodPost {
		/* Anything other than get or post, just silently reject it */
		return
	}

	/* Read body */
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		cwlog.DoLog(true, "Error reading request body: %v", err)
		return
	}
	input := string(bytes)

	/* Empty body, silently reject */
	if input == "" {
		cwlog.DoLog(true, "empty body")
		return
	}

	/* Send to command parser */
	commandParser(input, w)
}
