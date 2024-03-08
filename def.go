package main

import "time"

const (
	FrameSpeedMS  = 200
	lobbyMaxNames = 5

	MAX_BOARD_SIZE = 0xFF
	MAX_IDLE       = time.Minute * 5
	MAX_KEEPALIVE  = time.Second * 15

	protoVersion = 7
	cTimeout     = time.Second * 15
)

/*
 * Below are custom typed
 * so we must cast to use them (to prevent unintended problems)
 * iota will automatically number them
 */

/* Directions */
type DIR uint8

const (
	DIR_NORTH DIR = iota
	DIR_EAST
	DIR_SOUTH
	DIR_WEST
)

/* Network commands */
type CMD uint8

const (
	CMD_INIT CMD = iota
	CMD_PINGPONG

	CMD_GETLOBBIES
	CMD_JOINLOBBY
	CMD_CREATELOBBY

	CMD_GODIR

	RECV_LOCALPLAYER
	RECV_LOBBYLIST
	RECV_KEYFRAME
	RECV_PLAYERUPDATE
)

/* Used for debug messages, this could be better */
var cmdNames map[CMD]string

func init() {
	cmdNames = make(map[CMD]string)
	cmdNames[CMD_INIT] = "CMD_INIT"
	cmdNames[CMD_PINGPONG] = "CMD_PINGPONG"
	cmdNames[CMD_GETLOBBIES] = "CMD_GETLOBBIES"
	cmdNames[CMD_JOINLOBBY] = "CMD_JOINLOBBY"
	cmdNames[CMD_CREATELOBBY] = "CMD_CREATELOBBY"
	cmdNames[CMD_GODIR] = "CMD_GODIR"
	cmdNames[RECV_LOCALPLAYER] = "RECV_LOCALPLAYER"
	cmdNames[RECV_LOBBYLIST] = "RECV_LOBBYLIST"
	cmdNames[RECV_KEYFRAME] = "RECV_KEYFRAME"
	cmdNames[RECV_PLAYERUPDATE] = "RECV_PLAYERUPDATE"
}
