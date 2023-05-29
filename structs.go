package main

import (
	"sync"
	"time"
)

const (
	dbFile = "data.json"

	/* https://pkg.go.dev/time#pkg-constants */
	tFormat string = "2006-01-02-15-04-05"
)

var (
	authData    = dbDataType{BaseDownloadPath: "https://mywebsite/dlPath/", DefaultLifeHours: 200, Builds: []buildDataType{}}
	dbMutex     sync.Mutex
	dbDirty     bool
	newestBuild buildInfoType = buildInfoType{}
)

/* Main database struct */
type dbDataType struct {
	PushPass         string
	BaseDownloadPath string
	DefaultLifeHours int
	Builds           []buildDataType
}

/* Individual builds */
type buildDataType struct {
	Valid         bool
	VersionString string

	Pass  string `json:"p,omitempty"`
	Reply string `json:"r,omitempty"`

	AuthorizationCount int
	UpdateCheckCount   int
	DownloadCount      int

	LastAccessed int64
	Birth        int64
	Lifespan     int

	versData buildInfoType
}

/* Internal build info, not saved */
type buildInfoType struct {
	major         int
	year          int
	month         int
	day           int
	hour          int
	minute        int
	second        int
	versionString string
	buildDate     time.Time

	parent *buildDataType
}
