package main

import (
	"golang.org/x/exp/inotify"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	// "time"
)

// signalFile represents a file to watch for creation or deletion.
type signalFile struct {
	File    *os.File
	Signal  uint32
	Channel chan bool
}

// Exists returns true or false if a file exists or not
func (file *signalFile) Exists() bool {
	if _, err := os.Stat(file.File.Name()); err == nil {
		return true
	} else {
		return false
	}
}

// Remove deletes a file if it exists, and does nothing if it doesn't.
func (file *signalFile) Remove() {
	cmd := exec.Command("rm", file.File.Name(), "-f")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	// // TODO this doesn't trigger the signal for some reason
	// if _, err := os.Stat(file.File.Name()); err == nil {
	// 	err := os.Remove(file.File.Name())
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }
	// file.File.Close()
}

// Touch creates a file if it doesn't exist, and updates it's time stamps if it does.
func (file *signalFile) Touch() {
	cmd := exec.Command("touch", file.File.Name())
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	// // TODO this doesn't trigger the signal for some reason
	// if _, err := os.Stat(file.File.Name()); os.IsNotExist(err) {
	// 	_, err = os.Create(file.File.Name())
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// } else {
	// 	log.Println(os.Chtimes(file.File.Name(), time.Now(), time.Now()))
	// }
	// file.File.Close()
}

// WaitForSignal sends on channel when inotify receives specified signal
func (file *signalFile) WaitForSignal() {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.AddWatch(path.Dir(file.File.Name()), file.Signal)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case event := <-watcher.Event:
			matched, err := regexp.MatchString(file.File.Name(), event.Name)
			if err != nil {
				log.Fatal(err)
			}
			if matched {
				file.Channel <- true
			}
		case err := <-watcher.Error:
			log.Fatal(err)
		}
	}
}
