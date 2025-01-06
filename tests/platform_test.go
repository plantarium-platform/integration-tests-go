package tests

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

var platformProcess *os.Process

func TestMain(m *testing.M) {
	log.Println("Setting up platform for tests...")

	// Prepare the platform
	startCmd, err := PreparePlatform()
	if err != nil {
		log.Fatalf("Platform preparation failed: %v", err)
	}

	// Ensure platform is shut down after tests
	var exitCode int
	defer func() {
		log.Println("Tearing down platform...")
		if startCmd.Process != nil {
			if err := ShutdownPlatform(startCmd.Process); err != nil {
				log.Printf("Error during platform shutdown: %v", err)
			}
		}
		// Exit with the test result after cleanup
		os.Exit(exitCode)
	}()

	// Set up stderr pipe
	stderrPipe, err := startCmd.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to set up stderr pipe: %v", err)
	}

	// Start the platform process
	if err := startCmd.Start(); err != nil {
		log.Fatalf("Failed to start the platform: %v", err)
	}

	// Start goroutines to monitor stderr
	ready := make(chan bool)
	go monitorPipe(stderrPipe, "[Platform stderr]", "Platform started successfully", ready)

	// Wait for readiness
	log.Println("Waiting for platform readiness...")
	select {
	case <-ready:
		log.Println("Platform readiness confirmed.")
	case <-time.After(10 * time.Second):
		log.Fatalf("Platform readiness timeout.")
	}

	// Run tests and store the result
	exitCode = m.Run()
}

func monitorPipe(pipe io.ReadCloser, prefix string, readySignal string, ready chan bool) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		log.Println(prefix, line)
		if readySignal != "" && strings.Contains(line, readySignal) {
			select {
			case ready <- true:
			default:
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("%s Error reading pipe: %v", prefix, err)
	}
}
