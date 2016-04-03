package main

import "os/exec"
import "testing"
import "time"
import "os"
import "golang.org/x/exp/inotify"

func TestTouchFile(test *testing.T) {
	testFile := "./testFile.file"

	// Remove any existing testFile
	if _, err := os.Stat(testFile); err == nil {
		err := os.Remove(testFile)
		if err != nil {
			test.Fatal(err)
		}
	}

	// Test new file creation
	err := touchFile(testFile)
	if err != nil {
		test.Fatal(err)
	}
	if _, err := os.Stat(testFile); err != nil {
		test.Fatal("File was not created")
	}

	// Test existing file updation
	file, err := os.Stat(testFile)
	if err != nil {
		test.Fatal(err)
	}
	mtime := file.ModTime()
	err = touchFile(testFile)
	if err != nil {
		test.Fatal(err)
	}
	file, err = os.Stat(testFile)
	if err != nil {
		test.Fatal(err)
	}
	if file.ModTime() == mtime {
		test.Fatal("Mtime was not updated")
	} else {
		test.Logf("Mtime was updated from %s to %s", mtime, file.ModTime())
	}

	// Cleanup
	err = os.Remove(testFile)
	if err != nil {
		test.Fatal(err)
	}
}

type testFileWait struct {
	Signal         uint32
	Command        string
	CommandOptions []string
}

func TestWaitFileFor(test *testing.T) {
	testFile := "testWatchFile.file"

	//// test file creation - inotify.IN_CLOSE_WRITE
	testSet1 := new(testFileWait)
	testSet1.Signal = inotify.IN_CLOSE_WRITE
	testSet1.Command = "touch"
	testSet1.CommandOptions = []string{testFile}
	//// test file deletion - inotify.IN_DELETE
	testSet2 := new(testFileWait)
	testSet2.Signal = inotify.IN_DELETE
	testSet2.Command = "rm"
	testSet2.CommandOptions = []string{"-f", testFile}

	testSets := []*testFileWait{testSet1, testSet2}
	// testSets := []*testFileWait{testSet2, testSet1} // reversing breaks, since the second run expects the file to exist already

	// ensure file deoesn't exist
	if _, err := os.Stat(testFile); err == nil {
		err = os.Remove(testFile)
		if err != nil {
			test.Fatal(err)
		}
	}

	for _, testSet := range testSets {
		test.Logf("testSet %s", testSet)
		// setup watch
		testChan := make(chan bool)
		go waitFileFor(testChan, testFile, testSet.Signal)

		// setup timeout
		timeout := make(chan bool)
		go func() {
			time.Sleep(1 * time.Second)
			timeout <- true
		}()

		// create/remove file
		go func() {
			cmd := exec.Command(testSet.Command, testSet.CommandOptions...)
			err := cmd.Start()
			if err != nil {
				test.Fatal(err)
			}
			err = cmd.Wait()
			if err != nil {
				test.Fatalf("Command finished with error: %v", err)
			}
		}()

		// setup channel receive
		select {
		case <-testChan:
			test.Logf("Received signal for test file %s", testFile)
		case <-timeout:
			test.Fatalf("Received no signal after 2 seconds for test file %s", testFile)
		}
	}

	// cleanup
	if _, err := os.Stat(testFile); err == nil {
		err := os.Remove(testFile)
		if err != nil {
			test.Fatal(err)
		}
	}

}
