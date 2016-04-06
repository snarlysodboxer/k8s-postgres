package main

import (
	"fmt"
	"golang.org/x/exp/inotify"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"
)

func main() {

	postgresMode := os.Getenv("POSTGRES_MODE")                       // "master" or "slave"
	postgresTriggerFile := os.Getenv("POSTGRES_TRIGGER_FILE")        // "/data/postgres/postgresql.trigger"
	postgresSlaveIPFile := os.Getenv("POSTGRES_SLAVE_IP_FILE")       // "/data/postgres/slave_ip"
	postgresDataDir := os.Getenv("POSTGRES_DATA_DIR")                // "/data/postgres/data"
	postgresRecoveryConfFile := os.Getenv("POSTGRES_RECOVERY_FILE")  // "/data/postgres/data/recovery.conf"
	postgresEntrypoint := os.Getenv("POSTGRES_ENTRYPOINT")           // "/usr/lib/postgresql/9.5/bin/postgres"
	postgresOptions := os.Getenv("POSTGRES_OPTIONS")                 // "-D /data/postgres/data -c config_file=/data/postgres/conf/postgresql.conf"
	postgresServiceHost := os.Getenv("POSTGRES_MASTER_SERVICE_HOST") // set automatically by the "postgres_master" service you create before starting this container
	postgresServicePort := os.Getenv("POSTGRES_MASTER_SERVICE_PORT") // set automatically by the "postgres_master" service you create before starting this container
	postgresReplicatorUser := os.Getenv("POSTGRES_REPLICATOR_USER")  // replicator
	postgresReplicatorPass := os.Getenv("POSTGRES_REPLICATOR_PASS")  // your pass

	// Ensure `POSTGRES_MODE` is set
	matched, err := regexp.MatchString(`(master|slave)`, postgresMode)
	if err != nil {
		log.Fatal(err)
	}
	if !matched {
		log.Fatal("POSTGRES_MODE env var is unset, please set it")
	}

	// Master or Slave
	master, err := regexp.MatchString(`master`, postgresMode)
	if err != nil {
		log.Fatal(err)
	}
	if master {
		//// Is Master

		// Create postgresTriggerFile
		triggerFile := new(signalFile)
		triggerFile.File = os.NewFile(0, postgresTriggerFile)
		triggerFile.Touch()

		// Wait for postgresSlaveIPFile to appear
		slaveIPFile := new(signalFile)
		slaveIPFile.File = os.NewFile(0, postgresSlaveIPFile)
		slaveIPFile.Channel = make(chan bool)
		defer close(slaveIPFile.Channel)
		slaveIPFile.Signal = inotify.IN_CLOSE_WRITE
		go slaveIPFile.WaitForSignal()
		<-slaveIPFile.Channel

		// Read postgresSlaveIPFile file
		bytes, err := ioutil.ReadFile(postgresSlaveIPFile)
		if err != nil {
			log.Fatal(err)
		}
		slaveIP := strings.Replace(string(bytes), `/n`, "", -1)

		// Start reverse proxy - https://github.com/lumanetworks/go-tcp-proxy
		syscall.Exec("/tcp-proxy", []string{`-l="localhost:5432"`, fmt.Sprintf("-r=\"%s:5433\"", slaveIP)}, []string{})

	} else {
		//// Is Slave

		// Watch for positive health check from postgres master
		masterAlive := make(chan bool)
		defer close(masterAlive)
		command := "nc"
		options := fmt.Sprintf("%s %s < /dev/null > /dev/null; [ `echo $?` -eq 0 ]", postgresServiceHost, postgresServicePort)
		go func() {
			for {
				time.Sleep(1 * time.Second)
				cmd := exec.Command(command, options)
				err := cmd.Run()
				if err == nil {
					masterAlive <- true
				}
			}
		}()

		// Watch for tigger file to exist or appear
		triggerFile := new(signalFile)
		triggerFile.File = os.NewFile(0, postgresSlaveIPFile)
		triggerFile.Channel = make(chan bool)
		defer close(triggerFile.Channel)
		triggerFile.Signal = inotify.IN_CLOSE_WRITE
		if triggerFile.Exists() {
			<-triggerFile.Channel
		} else {
			go triggerFile.WaitForSignal()
		}

		// Wait for positive health check from postgres master
		//   OR
		// Tiggerfile to exist or appear
		select {
		// Run as master
		case <-triggerFile.Channel:
			// Start PostgreSQL as master
			syscall.Exec(postgresEntrypoint, []string{postgresOptions}, []string{})

		// Run as slave
		case <-masterAlive:
			// Create `recovery.conf` file
			contents := fmt.Sprintf(
				"standby_mode = 'on'\nprimary_conninfo = 'host=%s port=%s user=%s password=%s sslmode=require'\ntrigger_file = '%s'\n",
				postgresServiceHost,
				postgresServicePort,
				postgresReplicatorUser,
				postgresReplicatorPass,
				postgresTriggerFile,
			)
			bytes := []byte(contents)
			err := ioutil.WriteFile(postgresRecoveryConfFile, bytes, 0640)
			if err != nil {
				log.Fatal(err)
			}

			// Run `pg_basebackup`
			command = "/usr/bin/pg_basebackup"
			options = fmt.Sprintf("-w -h %s -p %s -U %s -D %s -v -x", postgresServiceHost, postgresServicePort, postgresReplicatorUser, postgresDataDir)
			cmd := exec.Command(command, options)
			err = cmd.Run()
			if err != nil {
				log.Fatal(err)
			}

			// Start PostgreSQL as slave
			syscall.Exec(postgresEntrypoint, []string{postgresOptions}, []string{})
		}
	}

}
