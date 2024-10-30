package internal

import (
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *PostgreSQLClient

// TestMain sets up the database connection for tests
func TestMain(m *testing.M) {
	var err error

	// Load the app configuration from the YAML file
	config, err := LoadAppConfig()
	if err != nil {
		log.Fatalf("Failed to load app config: %v", err)
	}

	// Connect to PostgreSQL using the loaded configuration
	testDB, err = ConnectPostgreSQL(*config)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer testDB.Close()

	// Create the devices table
	if err := createTestTable(); err != nil {
		log.Fatalf("Failed to create test table: %v", err)
	}

	// Run tests
	os.Exit(m.Run())
}

// createTestTable creates the devices table for testing
func createTestTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS devices (
		device_id VARCHAR PRIMARY KEY,
		type VARCHAR NOT NULL,
		state VARCHAR NOT NULL
	);`
	_, err := testDB.DB.Exec(query)
	return err
}

func TestInsertDevice(t *testing.T) {
	device := Device{ID: "1", Type: "light", State: "off"}

	// Attempt to insert the device
	err := testDB.InsertDevice(device)
	require.NoError(t, err, "Failed to insert device")

	// Verify the device was inserted
	fetchedDevice, err := testDB.GetDevice(device.ID)
	require.NoError(t, err, "Failed to get device")
	require.NotNil(t, fetchedDevice, "Expected device to be found, but it was nil")

	// Assert that the fetched device matches the inserted device
	assert.Equal(t, device.ID, fetchedDevice.ID, "Device ID should match")
	assert.Equal(t, device.Type, fetchedDevice.Type, "Device Type should match")
	assert.Equal(t, device.State, fetchedDevice.State, "Device State should match")
}

func TestUpdateDeviceState(t *testing.T) {
	device := Device{ID: "2", Type: "thermostat", State: "off"}
	err := testDB.InsertDevice(device)
	require.NoError(t, err, "Failed to insert device for update test")

	newState := "on"
	err = testDB.UpdateDeviceState(device.ID, newState)
	require.NoError(t, err, "Failed to update device state")

	// Verify the device state was updated
	fetchedDevice, err := testDB.GetDevice(device.ID)
	require.NoError(t, err, "Failed to get device after update")
	require.NotNil(t, fetchedDevice, "Expected device to be found, but it was nil")

	// Assert that the fetched device's state matches the updated state
	assert.Equal(t, newState, fetchedDevice.State, "Device State should match the updated state")
}

func TestGetDevice(t *testing.T) {
	device := Device{ID: "3", Type: "sensor", State: "active"}
	err := testDB.InsertDevice(device)
	require.NoError(t, err, "Failed to insert device for get test")

	fetchedDevice, err := testDB.GetDevice(device.ID)
	require.NoError(t, err, "Failed to get device")
	require.NotNil(t, fetchedDevice, "Expected device to be found, but it was nil")

	// Assert that the fetched device matches the inserted device
	assert.Equal(t, device.ID, fetchedDevice.ID, "Device ID should match")
	assert.Equal(t, device.Type, fetchedDevice.Type, "Device Type should match")
	assert.Equal(t, device.State, fetchedDevice.State, "Device State should match")
}
