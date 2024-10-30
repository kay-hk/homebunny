package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"smart-home-assistant/internal"

	"github.com/stretchr/testify/assert"
)

// Mock server to handle device registration and event publishing
func setupMockServer(t *testing.T, path string, expectedStatusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate the request path and method
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Read the request body
		body, err := io.ReadAll(r.Body) 
		assert.NoError(t, err, "should read request body without error")

		// Unmarshal to validate payload structure
		var device internal.Device
		err = json.Unmarshal(body, &device)
		assert.NoError(t, err, "should unmarshal device from request body")

		// Respond with the expected status code
		w.WriteHeader(expectedStatusCode)
	}))
}

func TestRegisterDevice(t *testing.T) {
	mockServer := setupMockServer(t, "/devices", http.StatusCreated)
	defer mockServer.Close()

	// Override the server URL for testing
	originalURL := registerDeviceURL
	registerDeviceURL = mockServer.URL + "/devices"
	defer func() { registerDeviceURL = originalURL }() // Restore the original URL after test

	// Define a sample device
	device := internal.Device{ID: "tv1", Type: "tv", State: "off"}

	err := registerDevice(device)
	assert.NoError(t, err, "expected no error when registering device")
}

func TestPublishEvent(t *testing.T) {
	mockServer := setupMockServer(t, "/publish", http.StatusOK)
	defer mockServer.Close()

	// Override the server URL for testing
	originalURL := publishEventURL
	publishEventURL = mockServer.URL + "/publish"
	defer func() { publishEventURL = originalURL }() // Restore the original URL after test

	// Define a sample device event
	device := internal.Device{ID: "tv1", Type: "tv", State: "on"}

	err := publishEvent(device)
	assert.NoError(t, err, "expected no error when publishing event")
}