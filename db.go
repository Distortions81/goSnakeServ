package main

import (
	"F38Auth/cwlog"
	"encoding/json"
	"os"
	"time"
)

/* Change list accessed date */
func updateAccessed(pass string) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	for i := range authData.Builds {
		if authData.Builds[i].Pass == pass {
			authData.Builds[i].LastAccessed = time.Now().Unix()
			authData.Builds[i].AuthorizationCount++
			dbDirty = true
			return
		}
	}
}

/* Disable entries that past max life */
func pruneExpired() {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	pruned := false

	for i := range authData.Builds {
		if authData.Builds[i].Valid {
			if time.Since(time.Unix(authData.Builds[i].Birth, 0)) > time.Duration(authData.Builds[i].Lifespan)*time.Hour {
				authData.Builds[i].Valid = false
				pruned = true
				dbDirty = true
				cwlog.DoLog(true, "Pruned: %v ", authData.Builds[i].VersionString)
			}
		}

	}

	if pruned {
		go writeDB(false)
	}

}

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
