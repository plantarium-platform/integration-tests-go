package tests

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Constants for the test
const (
	herbariumRepoURL      = "https://github.com/plantarium-platform/herbarium-go"
	testdataTempPath      = "testdata/temp"
	testdataHerbariumPath = "testdata/plantarium"
	buildPath             = "testdata/temp/herbarium/bin"
	executablePath        = "testdata/temp/herbarium/bin/herbarium"
	haproxyRunScript      = "haproxy-run.sh"
	haproxyRunDir         = "testdata/temp/.haproxy-run"
)

// InitPlatform sets up and runs the platform for testing.
func InitPlatform() (*os.Process, error) {
	// Step 1: Prepare the testdata/temp directory
	if err := os.MkdirAll(testdataTempPath, 0755); err != nil {
		return nil, err
	}

	// Step 2: Clone or update the Herbarium repository
	repoPath := filepath.Join(testdataTempPath, "herbarium")
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.Println("Cloning the Herbarium repository...")
		cmd := exec.Command("git", "clone", herbariumRepoURL, repoPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, err
		}
	} else {
		log.Println("Updating the Herbarium repository...")
		cmd := exec.Command("git", "-C", repoPath, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, err
		}
	}

	// Step 3: Build the Herbarium executable
	log.Println("Building the Herbarium executable...")
	if err := os.MkdirAll(buildPath, 0755); err != nil {
		return nil, err
	}
	cmd := exec.Command("go", "build", "-o", executablePath, filepath.Join(repoPath, "cmd", "herbarium", "main.go"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Step 4: Run the HAProxy setup script
	log.Println("Starting HAProxy...")
	haproxyCmd := exec.Command("bash", haproxyRunScript)
	haproxyCmd.Dir = filepath.Join(repoPath, "resources")
	haproxyCmd.Stdout = os.Stdout
	haproxyCmd.Stderr = os.Stderr
	if err := haproxyCmd.Run(); err != nil {
		return nil, err
	}

	// Wait for HAProxy to initialize
	log.Println("Waiting for HAProxy to initialize...")
	time.Sleep(5 * time.Second)

	// Step 5: Start the Herbarium platform
	log.Println("Starting the Herbarium platform...")
	cmd = exec.Command(executablePath)
	cmd.Env = append(os.Environ(), "PLANTARIUM_ROOT_FOLDER=testdata/plantarium") // Use test folder as the root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	log.Println("Platform started successfully.")
	return cmd.Process, nil
}
