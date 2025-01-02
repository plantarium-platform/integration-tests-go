package tests

import (
	"fmt"
	"log"
	"os"
)

// ShutdownPlatform kills the platform process.
func ShutdownPlatform(process *os.Process) error {
	if err := process.Kill(); err != nil {
		return fmt.Errorf("failed to kill the platform process: %w", err)
	}
	log.Println("Platform process terminated successfully.")
	return nil
}
