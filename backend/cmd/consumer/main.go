package main

import (
	"fmt"
	"log"
	"os"
	"smart-home-assistant/internal"
)

func main() {
	// Load the application configuration
	config, err := internal.LoadAppConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Get device type and state pattern from environment or default values
	deviceType := getEnv("DEVICE_TYPE", "air_conditioner") // Device type, e.g., "tv", "lights", "air_conditioner", "heater"
	statePattern := getEnv("STATE_PATTERN", "#")           // State pattern, e.g., "on", "off", "cooling", "heating", "#"

	// Connect to RabbitMQ using values from the config
	conn, err := internal.ConnectRabbitMQ(
		config.RabbitMQ.User,
		config.RabbitMQ.Password,
		config.RabbitMQ.Host,
		config.RabbitMQ.VHost,
	)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Create RabbitMQ client
	client, err := internal.NewRabbitMQClient(conn)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ client: %v", err)
	}
	defer client.Close()

	// Log if the connection or channel is closed
	if !client.ConnIsOpen() {
		log.Println("RabbitMQ connection is closed.")
	}

	if client.ChannelIsClosed() {
		log.Println("RabbitMQ channel is closed.")
	}

	// Create queue for the device type using the config
	queue, err := client.CreateQueue(fmt.Sprintf("%s_queue", deviceType))
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}

	// Bind the queue to the topic exchange with a flexible routing key like "device.air_conditioner.#"
	exchangeName := "device_events"
	routingKey := fmt.Sprintf("device.%s.%s", deviceType, statePattern)
	err = client.CreateBinding(queue.Name, routingKey, exchangeName)
	if err != nil {
		log.Fatalf("Failed to create binding: %v", err)
	}

	// Consume events for the device
	messages, err := client.ConsumeEvent(queue.Name)
	if err != nil {
		log.Fatalf("Failed to consume events: %v", err)
	}

	go func() {
		for msg := range messages {
			log.Printf("%s received event: %s", deviceType, msg.Body)
			handleDeviceEvent(deviceType, string(msg.Body))

			// Manually acknowledge the message
			msg.Ack(false)
		}
	}()

	// Keep the consumer running
	select {}
}

// Helper function to get environment variables or fallback to default values
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// Utility function to handle events for a device
func handleDeviceEvent(deviceID, event string) {
	switch event {
	case "Air conditioner is cooling":
		log.Printf("Cooling on device %s", deviceID)
	case "Air conditioner is turned off":
		log.Printf("Turning off device %s", deviceID)
	default:
		log.Printf("Unknown event for device %s: %s", deviceID, event)
	}
}
