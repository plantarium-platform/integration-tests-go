package tests

import (
	"log"
	"os"
	"testing"
)

var platformProcess *os.Process

// TestMain handles setup and teardown for all tests in the package.
func TestMain(m *testing.M) {
	log.Println("Setting up platform for tests...")

	// Initialize the platform
	process, err := InitPlatform()
	if err != nil {
		log.Fatalf("Platform initialization failed: %v", err)
	}
	platformProcess = process

	// Ensure platform is shut down even if tests panic or fail
	defer func() {
		log.Println("Tearing down platform...")
		if err := ShutdownPlatform(platformProcess); err != nil {
			log.Printf("Error during platform shutdown: %v", err)
		}
	}()

	// Run tests
	exitCode := m.Run()

	// Perform platform shutdown explicitly before exiting
	log.Println("Exiting test suite...")
	if err := ShutdownPlatform(platformProcess); err != nil {
		log.Printf("Error during final platform shutdown: %v", err)
	}

	// Exit with the test results
	os.Exit(exitCode)
}
