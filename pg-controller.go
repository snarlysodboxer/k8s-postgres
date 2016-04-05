package main

import (
	"fmt"
	"golang.org/x/exp/inotify"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"syscall"
)

func main() {

	postgresMode := os.Getenv("POSTGRES_MODE") // "master" or "slave"
	// postgresBaseDir := os.Getenv("POSTGRES_BASE_DIR")      // "/data/postgres"
	triggerFilePath := os.Getenv("POSTGRES_TRIGGER_FILE")  // "/data/postgres/postgresql.trigger"
	slaveIPFilePath := os.Getenv("POSTGRES_SLAVE_IP_FILE") // "/data/postgres/slave_ip"
	// postgresEntrypoint := os.Getenv("POSTGRES_ENTRYPOINT") // "/usr/lib/postgresql/9.5/bin/postgres"
	// postgresOptions := os.Getenv("POSTGRES_OPTIONS")       // "-D $PGDATA -c config_file=/data/postgres/conf/postgresql.conf-D $PGDATA -c config_file=/data/postgres/conf/postgresql.conf"

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

		// Wait for slaveIPFilePath to appear
		slaveIPFile := new(signalFile)
		slaveIPFile.File = os.NewFile(0, slaveIPFilePath)
		slaveIPFile.Channel = make(chan bool)
		defer close(slaveIPFile.Channel)
		slaveIPFile.Signal = inotify.IN_CLOSE_WRITE
		go slaveIPFile.WaitForSignal()
		<-slaveIPFile.Channel

		// Read slaveIPFilePath file
		bytes, err := ioutil.ReadFile(slaveIPFilePath)
		if err != nil {
			log.Fatal(err)
		}
		slaveIP := string(bytes)
		// Start reverse proxy - https://github.com/lumanetworks/go-tcp-proxy
		syscall.Exec("/tcp-proxy", []string{`-l="localhost:5432"`, fmt.Sprintf("-r=\"%s:5433\"", slaveIP)}, []string{})

	} else {
		//// Is Slave

		// syscall.Exec(postgresEntrypoint, []string{postgresOptions}, []string{})
	}
}
