package main

import "time"

const (
	MAX_BOARD_SIZE = 0xFF
	MAX_IDLE       = time.Minute * 5
	MAX_KEEPALIVE  = time.Second * 15

	DIR_NORTH = 0
	DIR_EAST  = 1
	DIR_SOUTH = 2
	DIR_WEST  = 3
	DIR_NONE  = 4

	CMD_INIT     = 0x01
	CMD_PINGPONG = 0x02

	CMD_LOGIN = 0x10
	CMD_NAME  = 0x11

	CMD_GETLOBBIES  = 0x20
	CMD_JOINLOBBY   = 0x21
	CMD_CREATELOBBY = 0x22
	CMD_SETLOBBY    = 0x23

	CMD_SPAWN = 0x30
	CMD_GODIR = 0x31

	RECV_LOCALPLAYER  = 0x50
	RECV_LOBBYLIST    = 0x51
	RECV_LOBBYDATA    = 0x52
	RECV_KEYFRAME     = 0x53
	RECV_PLAYERUPDATE = 0x54
	RECV_APPLEPOS     = 0x55
)

var cmdNames map[byte]string

func init() {
	cmdNames = make(map[byte]string)
	cmdNames[CMD_INIT] = "CMD_INIT"
	cmdNames[CMD_PINGPONG] = "CMD_PINGPONG"
	cmdNames[CMD_LOGIN] = "CMD_LOGIN"
	cmdNames[CMD_NAME] = "CMD_NAME"
	cmdNames[CMD_GETLOBBIES] = "CMD_GETLOBBIES"
	cmdNames[CMD_JOINLOBBY] = "CMD_JOINLOBBY"
	cmdNames[CMD_CREATELOBBY] = "CMD_CREATELOBBY"
	cmdNames[CMD_SETLOBBY] = "CMD_SETLOBBY"
	cmdNames[CMD_SPAWN] = "CMD_SPAWN"
	cmdNames[CMD_GODIR] = "CMD_GODIR"
	cmdNames[RECV_LOCALPLAYER] = "RECV_LOCALPLAYER"
	cmdNames[RECV_LOBBYLIST] = "RECV_LOBBYLIST"
	cmdNames[RECV_LOBBYDATA] = "RECV_LOBBYDATA"
	cmdNames[RECV_KEYFRAME] = "RECV_KEYFRAME"
	cmdNames[RECV_PLAYERUPDATE] = "RECV_PLAYERUPDATE"
	cmdNames[RECV_APPLEPOS] = "RECV_APPLEPOS"
}
