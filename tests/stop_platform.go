package tests

import (
	"log"
	"os"
	"os/exec"
)

// ShutdownPlatform stops the Herbarium process and the HAProxy Docker container.
func ShutdownPlatform(process *os.Process) error {
	log.Println("Stopping Herbarium platform...")

	// Kill the Herbarium process
	if err := process.Kill(); err != nil {
		log.Printf("Failed to kill the Herbarium process: %v", err)
		return err
	}
	log.Println("Herbarium platform stopped successfully.")

	// Stop the HAProxy Docker container
	log.Println("Stopping HAProxy Docker container...")
	cmd := exec.Command("docker", "stop", "integration-tests-haproxy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to stop HAProxy Docker container: %v", err)
		return err
	}
	log.Println("HAProxy Docker container stopped successfully.")

	return nil
}
