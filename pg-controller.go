package main

import "fmt"
import "regexp"
import "golang.org/x/exp/inotify"
import "os"
import "log"
import "strings"

func main() {

	postgresMode := os.Getenv("POSTGRES_MODE")                                                                                                 // "master" or "slave"
	postgresBaseDir := os.Getenv("POSTGRES_BASE_DIR")                                                                                          // "/data/postgres"
	triggerFilePath := strings.Replace(fmt.Sprintf("%s/%s", postgresBaseDir, os.Getenv("SLAVE_TRIGGER_FILE")), "//", "/", -1)                  // "/data/postgres/postgresql.trigger"
	shutdownFilePath := strings.Replace(fmt.Sprintf("%s/%s", postgresBaseDir, os.Getenv("SHUTDOWN_SLAVE_FILE")), "//", "/", -1)                // "/data/postgres/SHUTDOWN_SLAVE"
	shutdownSuccessFilePath := strings.Replace(fmt.Sprintf("%s/%s", postgresBaseDir, os.Getenv("SHUTDOWN_SLAVE_SUCCESS_FILE")), "//", "/", -1) // "/data/postgres/SHUTDOWN_SLAVE_SUCCESS"

	// Ensure `POSTGRES_MODE` is set
	matched, err := regexp.MatchString(`(master|slave)`, postgresMode)
	if err != nil {
		log.Fatal(err)
	}
	if !matched {
		log.Fatal("POSTGRES_MODE env var is unset, please set it")
	}

	master, err := regexp.MatchString(`master`, postgresMode)
	if err != nil {
		log.Fatal(err)
	}
	if master {
		//// Is Master

		// Create triggerFilePath
		err := touchFile(triggerFilePath)
		if err != nil {
			log.Fatal(err)
		}

		// Create shutdownFilePath
		err = touchFile(shutdownFilePath)
		if err != nil {
			log.Fatal(err)
		}

		// Wait for shutdownSuccessFilePath to appear
		shutdownSuccessChan := make(chan bool, 1)
		go waitFileFor(shutdownSuccessChan, shutdownSuccessFilePath, inotify.IN_CLOSE_WRITE)
		<-shutdownSuccessChan

		// Ensure Postgres Running as Master
		ensureRunningPostgresMaster()

	} else {
		//// Is Slave

	}
}
