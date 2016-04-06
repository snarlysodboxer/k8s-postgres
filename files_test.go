package main

import (
	"golang.org/x/exp/inotify"
	"os"
	"testing"
	"time"
)

// Test new file creation
func TestTouchNew(test *testing.T) {
	testFile := new(signalFile)
	testFile.File = os.NewFile(0, "./testFileTouchNew.file")

	// Remove any existing
	testFile.Remove()

	// Create
	testFile.Touch()
	if _, err := os.Stat(testFile.File.Name()); err != nil {
		test.Fatal("File was not created")
	}
	// Cleanup
	testFile.Remove()
}

// Test existing file updation
func TestTouchExisting(test *testing.T) {
	testFile := new(signalFile)
	testFile.File = os.NewFile(0, "./testFileTouch.file")

	// Create
	testFile.Touch()

	// Test
	file, err := os.Stat(testFile.File.Name())
	if err != nil {
		test.Fatal(err)
	}
	mtime := file.ModTime()
	time.Sleep(500 * time.Millisecond)
	testFile.Touch()
	file, err = os.Stat(testFile.File.Name())
	if err != nil {
		test.Fatal(err)
	}
	if file.ModTime() == mtime {
		test.Fatalf("Mtime was not updated, mtime: %s file.ModTime(): %s", mtime, file.ModTime())
	} else {
		test.Logf("Mtime was updated from %s to %s", mtime, file.ModTime())
	}

	// Cleanup
	testFile.Remove()
}

// Test Remove
func TestRemove(test *testing.T) {
	testFile := new(signalFile)
	testFile.File = os.NewFile(0, "./testFileRemove.file")
	testFile.Touch()
	testFile.Remove()
	if _, err := os.Stat(testFile.File.Name()); err == nil {
		test.Fatal("File was not removed")
	}
}

// Test file creation - inotify.IN_CLOSE_WRITE
func TestWaitForSignalCreate(test *testing.T) {
	testFile := new(signalFile)
	testFile.File = os.NewFile(0, "./testFileCreate.file")
	testFile.Signal = inotify.IN_CLOSE_WRITE
	testFile.Channel = make(chan bool)
	defer close(testFile.Channel)

	// Ensure file deoesn't exist
	testFile.Remove()

	// Setup watch
	go testFile.WaitForSignal()

	// Setup timeout
	timeout := make(chan bool)
	defer close(timeout)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()

	// Create file
	testFile.Touch()

	// Setup channel receive
	select {
	case <-testFile.Channel:
		test.Logf("Received create signal for test file %s", testFile.File.Name())
	case <-timeout:
		test.Fatalf("Received no create signal after 1 second for test file %s", testFile.File.Name())
	}

	// Cleanup
	testFile.Remove()
}

// Test file deletion - inotify.IN_DELETE
func TestWaitForSignalDelete(test *testing.T) {
	testFile := new(signalFile)
	testFile.File = os.NewFile(0, "./testFileDelete.file")
	testFile.Signal = inotify.IN_DELETE
	testFile.Channel = make(chan bool)
	defer close(testFile.Channel)

	// Create file
	testFile.Touch()

	// Setup watch
	go testFile.WaitForSignal()

	// Setup timeout
	timeout := make(chan bool)
	defer close(timeout)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()

	// Remove file
	testFile.Remove()

	// Setup channel receive
	select {
	case <-testFile.Channel:
		test.Logf("Received remove signal for test file %s", testFile.File.Name())
	case <-timeout:
		test.Fatalf("Received no remove signal after 1 second for test file %s", testFile.File.Name())
	}
}
