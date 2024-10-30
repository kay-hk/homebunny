package main

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDelivery extends amqp091.Delivery to simulate Ack behavior in tests.
type MockDelivery struct {
	amqp091.Delivery
	AckCalled bool
}

func (m *MockDelivery) Ack(multiple bool) error {
	m.AckCalled = true
	return nil
}

// MockRabbitClient is a mock implementation of RabbitClient for testing.
type MockRabbitClient struct {
	QueueName      string
	BindingKey     string
	ExchangeName   string
	MessageChannel chan *MockDelivery
}

func (m *MockRabbitClient) CreateQueue(name string) (amqp091.Queue, error) {
	m.QueueName = name
	return amqp091.Queue{Name: name}, nil
}

func (m *MockRabbitClient) CreateBinding(queueName, bindingKey, exchangeName string) error {
	m.BindingKey = bindingKey
	m.ExchangeName = exchangeName
	return nil
}

func (m *MockRabbitClient) ConsumeEvent(queueName string) (<-chan *MockDelivery, error) {
	return m.MessageChannel, nil
}

func (m *MockRabbitClient) ConnIsOpen() bool      { return true }
func (m *MockRabbitClient) ChannelIsClosed() bool { return false }

func TestConsumer(t *testing.T) {
	// Set environment variables for device type and state pattern
	os.Setenv("DEVICE_TYPE", "air_conditioner")
	os.Setenv("STATE_PATTERN", "#")
	defer os.Unsetenv("DEVICE_TYPE")
	defer os.Unsetenv("STATE_PATTERN")

	// Initialize a mock RabbitMQ client
	mockClient := &MockRabbitClient{
		MessageChannel: make(chan *MockDelivery, 1),
	}

	// Create a queue and binding
	queue, err := mockClient.CreateQueue("air_conditioner_queue")
	require.NoError(t, err, "Failed to create queue")

	err = mockClient.CreateBinding(queue.Name, "device.air_conditioner.#", "device_events")
	require.NoError(t, err, "Failed to create binding")

	// Start consuming messages in a goroutine
	go func() {
		messages, err := mockClient.ConsumeEvent(queue.Name)
		require.NoError(t, err, "Failed to consume events")

		for msg := range messages {
			log.Printf("Received message: %s", msg.Body)
			handleDeviceEvent("air_conditioner", string(msg.Body))

			// Acknowledge the message
			err := msg.Ack(false)
			require.NoError(t, err, "Failed to acknowledge message")
			assert.True(t, msg.AckCalled, "Expected Ack to be called")
		}
	}()

	// Send a test message
	testMessage := &MockDelivery{
		Delivery: amqp091.Delivery{
			Body: []byte("Air conditioner is cooling"),
		},
	}
	mockClient.MessageChannel <- testMessage

	// Allow some time for message processing
	time.Sleep(500 * time.Millisecond)
	close(mockClient.MessageChannel)

	// Validate test results
	assert.Equal(t, "air_conditioner_queue", mockClient.QueueName, "Queue name should match")
	assert.Equal(t, "device.air_conditioner.#", mockClient.BindingKey, "Binding key should match")
	assert.Equal(t, "device_events", mockClient.ExchangeName, "Exchange name should match")
}
