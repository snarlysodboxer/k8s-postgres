package main

import (
// "time"
// "path"
// "golang.org/x/exp/inotify"
// "log"
// "regexp"
// "os"
)

type postgresInstance struct {
	Master bool
}

func (file *postgresInstance) EnsureRunning() {
	// Remove triggerFile
	// TODO see if postgresql.trigger gets changed to postgresql.done
	// err = os.Remove(triggerFilePath)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Start PostgreSQL master
}
