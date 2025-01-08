package tests

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Global HAProxy client for tests
var haproxyClient *resty.Client

func init() {
	haproxyClient = resty.New().
		SetBaseURL("http://localhost:5555/v3/services/haproxy").
		SetBasicAuth("admin", "mypassword").
		SetHeader("Content-Type", "application/json")
}

func TestHelloEndpoint(t *testing.T) {
	log.Println("Testing hello endpoint...")

	// Step 1: Send request to the hello endpoint
	resp, err := http.Get("http://localhost/hello")
	assert.NoError(t, err, "Failed to reach hello endpoint")
	defer resp.Body.Close()

	// Step 2: Validate the response
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected HTTP status 200")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "Failed to read response body")
	assert.Equal(t, "Hello, World!", string(body), "Expected response 'Hello, World!'")

	log.Println("Hello endpoint test passed.")
}

// TestHelloGraftEndpoint validates the existence of a graft node in the HAProxy backend and tests the hello-graft endpoint.
func TestHelloGraftEndpoint(t *testing.T) {
	log.Println("Testing hello-graft endpoint...")

	backendName := "hello-graft"
	expectedServerName := "hello-graft-service-graft-node"

	// Step 1: Fetch servers from the HAProxy backend
	log.Printf("Fetching servers from backend: %s", backendName)
	resp, err := haproxyClient.R().
		SetPathParam("backendName", backendName).
		Get("/configuration/backends/{backendName}/servers")

	assert.NoError(t, err, "Failed to fetch HAProxy backend configuration")
	assert.Equal(t, 200, resp.StatusCode(), "Expected HTTP status 200 from HAProxy API")

	// Parse the response body into a slice of servers
	var servers []map[string]interface{}
	err = json.Unmarshal(resp.Body(), &servers)
	assert.NoError(t, err, "Failed to parse HAProxy API response")

	// Check if the expected server exists
	foundServer := false
	for _, server := range servers {
		if server["name"] == expectedServerName {
			foundServer = true
			break
		}
	}
	require.True(t, foundServer, "Expected server %s not found in HAProxy backend %s", expectedServerName, backendName)

	if foundServer {
		log.Printf("Server %s found in backend %s.", expectedServerName, backendName)
	}

	// Step 2: Test the hello-graft endpoint
	log.Println("Sending request to hello-graft endpoint...")
	respHTTP, err := http.Get("http://localhost/hello-graft")
	assert.NoError(t, err, "Failed to reach hello-graft endpoint")
	defer respHTTP.Body.Close()

	assert.Equal(t, 200, respHTTP.StatusCode, "Expected HTTP status 200")

	body, err := io.ReadAll(respHTTP.Body)
	assert.NoError(t, err, "Failed to read response body")
	assert.Equal(t, "Hello, World!", string(body), "Expected response 'Hello, World!'")

	log.Println("Hello-graft endpoint test passed.")
}
