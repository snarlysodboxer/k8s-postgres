package main

import (
	"fmt"
	"golang.org/x/exp/inotify"
	"log"
	"os"
	"regexp"
	"strings"
)

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
		triggerFile := new(signalFile)
		triggerFile.File = os.NewFile(0, triggerFilePath)
		triggerFile.Touch()

		// Create shutdownFilePath
		shutdownFile := new(signalFile)
		shutdownFile.File = os.NewFile(0, shutdownFilePath)
		shutdownFile.Touch()

		// Wait for shutdownSuccessFilePath to appear
		shutdownSuccessFile := new(signalFile)
		shutdownSuccessFile.File = os.NewFile(0, shutdownSuccessFilePath)
		shutdownSuccessFile.Channel = make(chan bool)
		defer close(shutdownSuccessFile.Channel)
		shutdownSuccessFile.Signal = inotify.IN_CLOSE_WRITE
		go shutdownSuccessFile.WaitForSignal()
		<-shutdownSuccessFile.Channel

		// Ensure Postgres Running as Master
		ensureRunningPostgresMaster()

	} else {
		//// Is Slave

	}
}
