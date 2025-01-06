package tests

import (
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
