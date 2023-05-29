package main

import (
	"encoding/json"
	"goSnakeServ/cwlog"
	"os"
)

/* Read database from disk */
func readDB() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	bytes, err := os.ReadFile(dbFile)
	if err != nil {
		cwlog.DoLog(true, "readDB: %v", err)
		return err
	}

	err = json.Unmarshal(bytes, &authData)
	if err != nil {
		cwlog.DoLog(true, "readDB: %v", err)
	}

	return nil
}

/* Write database to disk */
func writeDB(force bool) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if !dbDirty && !force {
		return nil
	}

	bytes, err := json.MarshalIndent(authData, "", "    ")
	if err != nil {
		cwlog.DoLog(true, "writeDB: %v", err)
		return err
	}

	err = os.WriteFile(dbFile, bytes, 0644)
	if err != nil {
		cwlog.DoLog(true, "writeDB: %v", err)
		return err
	}

	dbDirty = false
	cwlog.DoLog(true, "Wrote database.")
	return nil
}
