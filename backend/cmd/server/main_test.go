package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"smart-home-assistant/internal"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

var (
	testDB           *internal.PostgreSQLClient // Mock or test DB client
	testRabbitClient internal.RabbitClient      // Adjusted to match the type returned by NewRabbitMQClient
)

func setup() {
	var err error

	// Load database configuration 
	dbConfig, err := internal.LoadAppConfig()
	if err != nil {
		log.Fatalf("Failed to load database config: %v", err)
	}

	// Connect to PostgreSQL for testing
	testDB, err = internal.ConnectPostgreSQL(*dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL for testing: %v", err)
	}

	// Connect to RabbitMQ for testing
	conn, err := amqp091.Dial("amqp://kay:secret@localhost:5672/customers")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ for testing: %v", err)
	}
	testRabbitClient, err = internal.NewRabbitMQClient(conn) // NewRabbitMQClient returns internal.RabbitClient
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ client for testing: %v", err)
	}
}

func teardown() {
	// Close the test database connection
	if err := testDB.Close(); err != nil {
		log.Printf("Error closing test database: %v", err)
	}
	// Close the RabbitMQ client
	if err := testRabbitClient.Close(); err != nil {
		log.Printf("Error closing test RabbitMQ client: %v", err)
	}
}

func TestRegisterDeviceHandler(t *testing.T) {
	setup()
	defer teardown()

	t.Log("Starting TestRegisterDeviceHandler")

	device := internal.Device{ID: "1", Type: "light", State: "off"}
	body, err := json.Marshal(device)
	assert.NoError(t, err, "should marshal device to JSON")

	req := httptest.NewRequest(http.MethodPost, "/devices", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Inject the testDB dependency
	registerDeviceHandler(w, req, testDB)

	res := w.Result()
	assert.Equal(t, http.StatusCreated, res.StatusCode, "expected status code 201 for created")

	t.Log("Device registered successfully with response status 201")

	// Verify the device is stored in the database
	fetchedDevice, err := testDB.GetDevice(device.ID)
	assert.NoError(t, err, "should fetch the device without error")
	assert.NotNil(t, fetchedDevice, "expected device to be found")
	assert.Equal(t, device.ID, fetchedDevice.ID, "device ID should match")
	assert.Equal(t, device.Type, fetchedDevice.Type, "device Type should match")
	assert.Equal(t, device.State, fetchedDevice.State, "device State should match")

	t.Log("TestRegisterDeviceHandler completed successfully")
}

func TestPublishEventHandler(t *testing.T) {
	setup()
	defer teardown()

	t.Log("Starting TestPublishEventHandler")

	device := internal.Device{ID: "1", Type: "light", State: "on"}
	body, err := json.Marshal(device)
	assert.NoError(t, err, "should marshal device to JSON")

	req := httptest.NewRequest(http.MethodPost, "/publish", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Inject test dependencies
	rabbitClient = testRabbitClient // Ensure rabbitClient matches testRabbitClient's type

	publishEventHandler(w, req, testDB)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode, "expected status code 200 for OK")

	t.Log("Event published successfully with response status 200")

	// Check if the device state has been updated in the database
	updatedDevice, err := testDB.GetDevice(device.ID)
	assert.NoError(t, err, "should fetch the updated device without error")
	assert.NotNil(t, updatedDevice, "expected device to be found")
	assert.Equal(t, device.State, updatedDevice.State, "device State should be updated")

	t.Log("TestPublishEventHandler completed successfully")
}
