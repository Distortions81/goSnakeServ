package main

import "time"

const (
	MAX_BOARD_SIZE = 0xFF
	MAX_IDLE       = time.Minute * 1

	DIR_NONE  = 0
	DIR_NORTH = 1
	DIR_EAST  = 2
	DIR_SOUTH = 3
	DIR_WEST  = 4
)
