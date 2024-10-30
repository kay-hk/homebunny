package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConnectRabbitMQ tests the connection to RabbitMQ using the configuration
func TestConnectRabbitMQ(t *testing.T) {
	t.Log("Starting TestConnectRabbitMQ")

	// Load application configuration
	config, err := LoadAppConfig()
	assert.NoError(t, err, "should load config without error")

	// Connect to RabbitMQ using loaded configuration
	conn, err := ConnectRabbitMQ(config.RabbitMQ.User, config.RabbitMQ.Password, config.RabbitMQ.Host, config.RabbitMQ.VHost)
	assert.NoError(t, err, "should connect without error")
	assert.NotNil(t, conn, "connection should not be nil")
	t.Log("Connection established successfully")

	// Ensure cleanup
	if conn != nil {
		t.Log("Closing RabbitMQ connection")
		conn.Close()
	}

	t.Log("TestConnectRabbitMQ completed successfully")
}

func TestNewRabbitMQClient(t *testing.T) {
	t.Log("Starting TestNewRabbitMQClient")

	// Load application configuration
	config, err := LoadAppConfig()
	assert.NoError(t, err, "should load config without error")

	// Establish a connection to RabbitMQ
	conn, err := ConnectRabbitMQ(config.RabbitMQ.User, config.RabbitMQ.Password, config.RabbitMQ.Host, config.RabbitMQ.VHost)
	assert.NoError(t, err, "should connect without error")
	assert.NotNil(t, conn, "connection should not be nil")
	defer conn.Close()
	t.Log("Connection established successfully for RabbitMQ client")

	// Create a new RabbitMQ client
	client, err := NewRabbitMQClient(conn)
	assert.NoError(t, err, "should create RabbitMQ client without error")
	assert.NotNil(t, client.Ch, "channel should not be nil")
	defer client.Close()
	t.Log("RabbitMQ client created successfully")
	t.Log("TestNewRabbitMQClient completed successfully")
}

func TestCreateTopicExchange(t *testing.T) {
	t.Log("Starting TestCreateTopicExchange")

	// Load application configuration
	config, err := LoadAppConfig()
	assert.NoError(t, err, "should load config without error")

	// Establish a connection to RabbitMQ
	conn, err := ConnectRabbitMQ(config.RabbitMQ.User, config.RabbitMQ.Password, config.RabbitMQ.Host, config.RabbitMQ.VHost)
	assert.NoError(t, err, "should connect without error")
	assert.NotNil(t, conn, "connection should not be nil")
	defer conn.Close()
	t.Log("Connection established successfully for RabbitMQ client")

	// Create a new RabbitMQ client
	client, err := NewRabbitMQClient(conn)
	assert.NoError(t, err, "should create RabbitMQ client without error")
	assert.NotNil(t, client.Ch, "channel should not be nil")
	defer client.Close()
	t.Log("RabbitMQ client created successfully")

	// Attempt to create a topic exchange
	exchangeName := "test_device_events"
	err = client.CreateTopicExchange(exchangeName)
	assert.NoError(t, err, "should create topic exchange without error")
	t.Logf("Topic exchange %s created successfully", exchangeName)
	t.Log("TestCreateTopicExchange completed successfully")
}
