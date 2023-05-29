package main

import (
	"time"
)

/* Prune and write DB if dirty */
func backgroundTasks() {
	for {
		time.Sleep(time.Minute * 5)
		writeDB(false)
	}
}
