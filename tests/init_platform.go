package tests

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Constants for the test
const (
	herbariumRepoURL    = "https://github.com/plantarium-platform/herbarium-go"
	testdataTempPath    = "../testdata/temp"
	plantariumRootPath  = "../testdata/plantarium"
	herbariumBuildPath  = "../testdata/temp/bin"
	herbariumExecutable = "../testdata/temp/bin/herbarium"
	repoPath            = "../testdata/temp/herbarium-go"
	haproxyScriptPath   = "../testdata/haproxy/haproxy-run.sh" // Added this line
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

// PreparePlatform sets up and prepares the Herbarium platform for integration tests.
func PreparePlatform() (*exec.Cmd, error) {
	// Resolve absolute paths
	absolutePlantariumRoot, err := filepath.Abs(plantariumRootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for PLANTARIUM_ROOT_FOLDER: %w", err)
	}
	log.Printf("Resolved absolute path for PLANTARIUM_ROOT_FOLDER: %s", absolutePlantariumRoot)

	absoluteRepoPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for repository: %w", err)
	}
	log.Printf("Resolved absolute path for repository: %s", absoluteRepoPath)

	absoluteBuildPath, err := filepath.Abs(herbariumBuildPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for build folder: %w", err)
	}
	log.Printf("Resolved absolute path for build folder: %s", absoluteBuildPath)

	// Step 1: Ensure temp and build directories exist
	if err := os.MkdirAll(absoluteBuildPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create build directory: %w", err)
	}
	if err := os.MkdirAll(absoluteRepoPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repository directory: %w", err)
	}

	// Step 2: Clone or update the repository
	if _, err := os.Stat(absoluteRepoPath); os.IsNotExist(err) {
		log.Println("Cloning the Herbarium repository...")
		cmd := exec.Command("git", "clone", herbariumRepoURL, absoluteRepoPath)
		if err := runCommand(cmd); err != nil {
			return nil, err
		}
	} else {
		log.Println("Updating the Herbarium repository...")
		cmd := exec.Command("git", "-C", absoluteRepoPath, "pull")
		if err := runCommand(cmd); err != nil {
			return nil, err
		}
	}

	// Step 3: Run `go mod tidy`
	log.Println("Tidying Go modules...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = absoluteRepoPath
	if err := runCommand(tidyCmd); err != nil {
		return nil, err
	}

	// Step 4: Build the Herbarium executable
	log.Println("Building the Herbarium executable...")
	buildCmd := exec.Command("go", "build", "-o", absoluteBuildPath+"/herbarium", "cmd/herbarium/main.go")
	buildCmd.Dir = absoluteRepoPath
	if err := runCommand(buildCmd); err != nil {
		return nil, err
	}

	// Step 5: Start HAProxy
	// Resolve the absolute path to the HAProxy script
	haproxyRunScriptPath, err := filepath.Abs(filepath.Join("testdata", "haproxy", "haproxy-run.sh"))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for HAProxy script: %w", err)
	}
	log.Printf("Resolved absolute path for HAProxy script: %s", haproxyRunScriptPath)

	// Prepare and run the HAProxy command
	// Resolve the absolute path to the HAProxy script
	absoluteHaproxyScriptPath, err := filepath.Abs(haproxyScriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for HAProxy script: %w", err)
	}
	log.Printf("Resolved absolute path for HAProxy script: %s", absoluteHaproxyScriptPath)

	// Prepare and run the HAProxy command
	haproxyCmd := exec.Command("bash", absoluteHaproxyScriptPath)
	haproxyCmd.Dir = filepath.Dir(absoluteHaproxyScriptPath) // Set the working directory to the script's folder
	if err := runCommand(haproxyCmd); err != nil {
		return nil, fmt.Errorf("failed to start HAProxy: %w", err)
	}
	log.Println("HAProxy started successfully.")
	// Wait for a few seconds to ensure HAProxy is fully initialized
	time.Sleep(3 * time.Second)
	// Step 6: Verify HAProxy readiness
	log.Println("Verifying HAProxy readiness...")
	haproxyAPIURL := "http://localhost:5555/v3/services/haproxy/configuration/version"
	resp, err := http.Get(haproxyAPIURL)
	if err != nil || resp.StatusCode != http.StatusUnauthorized {
		return nil, fmt.Errorf("HAProxy is not ready or unreachable. Error: %v", err)
	}
	defer resp.Body.Close()
	log.Println("HAProxy is ready.")

	// Step 7: Prepare the platform command
	startCmd := exec.Command(herbariumExecutable)
	startCmd.Env = append(os.Environ(), "PLANTARIUM_ROOT_FOLDER="+absolutePlantariumRoot)
	return startCmd, nil
}
