package main

import (
	"log"
	"os"
	"os/exec"
	"regexp"
)

func main() {

	postgresMode := os.Getenv("POSTGRES_MODE") // "master" or "slave"
	// postgresBaseDir := os.Getenv("POSTGRES_BASE_DIR")      // "/data/postgres"
	triggerFilePath := os.Getenv("POSTGRES_TRIGGER_FILE") // "/data/postgres/postgresql.trigger"
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
		cmd := exec.Command("touch", triggerFilePath)
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}

		// Start port forward

	} else {
		//// Is Slave

	}
}
