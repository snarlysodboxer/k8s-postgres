package main

import (
	"golang.org/x/exp/inotify"
	"log"
	"os"
	"path"
	"regexp"
	"time"
)

func touchFile(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		_, err = os.Create(file)
		if err != nil {
			return err
		}
	} else {
		return os.Chtimes(file, time.Now(), time.Now())
	}
	return nil
}

func waitFileFor(channel chan bool, file string, signal uint32) {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.AddWatch(path.Dir(file), signal)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case event := <-watcher.Event:
			matched, err := regexp.MatchString(file, event.Name)
			if err != nil {
				log.Fatal(err)
			}
			if matched {
				channel <- true
			}
		case err := <-watcher.Error:
			log.Fatal(err)
		}
	}
}
