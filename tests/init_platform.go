package tests

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Constants for the test
const (
	herbariumRepoURL      = "https://github.com/plantarium-platform/herbarium-go"
	testdataTempPath      = "../testdata/temp"
	testdataHerbariumPath = "../../../testdata/plantarium"
	buildPath             = "bin"
	executablePath        = "bin/herbarium"
)

// runCommand runs a shell command and logs its command, working directory, output, and errors.
func runCommand(cmd *exec.Cmd) error {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Log the full command and the working directory
	log.Printf("Running command: %s\nWorking directory: %s", cmd.String(), cmd.Dir)
	err := cmd.Run()
	if err != nil {
		log.Printf("Command failed: %s\nError: %v\nOutput: %s\nStderr: %s", cmd.String(), err, stdout.String(), stderr.String())
		return err
	}

	log.Printf("Command succeeded: %s\nOutput: %s", cmd.String(), stdout.String())
	return nil
}

// InitPlatform sets up and starts the platform for integration tests.
func InitPlatform() (*os.Process, error) {
	// Step 1: Prepare the testdata/temp directory
	if err := os.MkdirAll(testdataTempPath, 0755); err != nil {
		return nil, err
	}

	// Step 2: Clone or update the Herbarium repository
	repoPath := filepath.Join(testdataTempPath, "herbarium-go")
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.Println("Cloning the Herbarium repository...")
		cmd := exec.Command("git", "clone", herbariumRepoURL, repoPath)
		if err := runCommand(cmd); err != nil {
			return nil, err
		}
	} else {
		log.Println("Updating the Herbarium repository...")
		cmd := exec.Command("git", "-C", repoPath, "pull")
		if err := runCommand(cmd); err != nil {
			return nil, err
		}
	}

	// Step 3: Run go mod tidy to ensure dependencies are up to date
	log.Println("Tidying Go modules...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = repoPath
	if err := runCommand(tidyCmd); err != nil {
		return nil, err
	}

	// Step 4: Build the Herbarium executable
	log.Println("Building the Herbarium executable...")
	if err := os.MkdirAll(buildPath, 0755); err != nil {
		return nil, err
	}
	buildCmd := exec.Command("go", "build", "-o", "../bin/herbarium", "cmd/herbarium/main.go")
	buildCmd.Dir = repoPath
	if err := runCommand(buildCmd); err != nil {
		return nil, err
	}

	// Step 5: Run HAProxy
	log.Println("HAProxy starting...")
	haproxyRunScriptPath := filepath.Join("testdata", "haproxy", "haproxy-run.sh")
	haproxyCmd := exec.Command("bash", haproxyRunScriptPath)

	// Set working directory to the root of the project
	haproxyCmd.Dir = "../"

	if err := runCommand(haproxyCmd); err != nil {
		log.Fatalf("Failed to start HAProxy: %v", err)
	}
	log.Println("HAProxy started successfully.")

	// Wait for 3 seconds to ensure HAProxy is fully initialized
	log.Println("Waiting for HAProxy to initialize...")
	time.Sleep(3 * time.Second)

	// Step 6: Verify HAProxy readiness
	haproxyAPIURL := "http://localhost:5555/v3/services/haproxy/configuration/version"
	log.Println("Verifying HAProxy readiness...")
	resp, err := http.Get(haproxyAPIURL)
	if err != nil {
		log.Fatalf("Failed to verify HAProxy readiness: %v. Please check HAProxy logs.", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.Fatalf("HAProxy is not ready. Received 404 Not Found from %s", haproxyAPIURL)
	}

	log.Println("HAProxy is ready.")

	// Step 7: Start the Herbarium platform and monitor its output
	log.Println("Starting the Herbarium platform...")
	startCmd := exec.Command(executablePath)
	startCmd.Dir = repoPath
	startCmd.Env = append(os.Environ(), "PLANTARIUM_ROOT_FOLDER="+testdataHerbariumPath)

	// Capture stdout and stderr
	stdoutPipe, err := startCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := startCmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := startCmd.Start(); err != nil {
		return nil, err
	}
	ready := make(chan bool)

	// Start a goroutine to monitor stdout
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			log.Println("[Platform stdout]", line)
			if strings.Contains(line, "Platform started successfully") {
				// Notify readiness
				select {
				case ready <- true:
					// Notify once, then continue monitoring
				default:
					// Avoid blocking if the channel has already been notified
				}
			}
		}
	}()

	// Start a goroutine to monitor stderr
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			log.Println("[Platform stderr]", scanner.Text())
		}
	}()

	// Wait for readiness notification
	<-ready
	log.Println("Platform readiness confirmed. Continuing test execution...")

	return startCmd.Process, nil
}
