package internal

import (
	"context"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"gopkg.in/yaml.v3"
)

// Device represents the structure of a device
type Device struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // Device type (e.g., "lightbulb", "TV")
	State string `json:"state"` // Device state (e.g., "off", "on")
}
type RabbitClient struct {
    Conn *amqp.Connection // Connection used by the client
    Ch   *amqp.Channel    // Channel used to process/send messagesjj
}

// LoadAppConfig loads the configuration from a YAML file.
func LoadAppConfig() (*AppConfig, error) {

	file, err := os.Open("path") // path to config
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config AppConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}
	return &config, nil
}

// ConnectRabbitMQ function
func ConnectRabbitMQ(username, password, host, vhost string) (*amqp.Connection, error) {
    conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/%s", username, password, host, vhost))
    if err != nil {
        return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
    }
    return conn, nil
}

func NewRabbitMQClient(conn *amqp.Connection) (RabbitClient, error) {
    ch, err := conn.Channel()
    if err != nil {
        return RabbitClient{}, fmt.Errorf("error creating channel: %w", err)
    }

    if err := ch.Confirm(false); err != nil{
        return RabbitClient{}, fmt.Errorf("error setting channel confirmation mode: %w", err)
    }
    return RabbitClient{
        Conn: conn, // Use exported field
        Ch:   ch,   // Use exported field
    }, nil
}

// ConnIsOpen returns whether the connection is open
func (rc RabbitClient) ConnIsOpen() bool {
	return !rc.Conn.IsClosed() // Check if the connection is not closed
}

// ChannelIsClosed returns whether the channel is closed
func (rc RabbitClient) ChannelIsClosed() bool {
	return rc.Ch.IsClosed() // Check if the channel is closed
}

// Close the channel
func (rc RabbitClient) Close() error {
	err := rc.Ch.Close() // Close the channel
	if err != nil {
		return fmt.Errorf("error closing channel: %w", err)
	}
	return nil
}

// ApplyQos sets Quality of Service (QoS) for the channel, limiting the number of unacknowledged messages.
func (rc RabbitClient) ApplyQos(prefetchCount int, global bool) error {
	err := rc.Ch.Qos(
		prefetchCount, // Prefetch count (e.g., 1 means one message at a time)
		0,             // Prefetch size (not used)
		global,        // Global QoS setting
	)
	if err != nil {
		return fmt.Errorf("error setting QoS: %w", err)
	}
	return nil
}

// CreateQueue registers a device by creating a durable queue with its name (deviceID).
func (rc RabbitClient) CreateQueue(deviceID string) (amqp.Queue, error) {
	q, err := rc.Ch.QueueDeclare(
		deviceID, // Device ID acts as queue name
		true,     // Durable
		false,    // AutoDelete
		false,    // Exclusive
		false,    // NoWait
		nil,      // Arguments
	)
	if err != nil {
        return amqp.Queue{}, fmt.Errorf("error creating queue for device %s: %w", deviceID, err)
	}
	return q, nil
}

// CreateTopicExchange declares a topic exchange, allowing flexible routing of events.
func (rc RabbitClient) CreateTopicExchange(exchangeName string) error {
	err := rc.Ch.ExchangeDeclare(
		exchangeName, // Name of the exchange
		"topic",      // Type of exchange
		true,         // Durable
		false,        // AutoDelete
		false,        // Internal
		false,        // NoWait
		nil,          // Arguments
	)
	if err != nil {
		return fmt.Errorf("error creating topic exchange %s: %w", exchangeName, err)
	}
	return nil
}

// CreateBinding is used to connect a queue to an Exchange using a dynamic binding rule
func (rc RabbitClient) CreateBinding(queueName, bindingKey, exchange string) error {
	// Bind the queue to an exchange with a flexible routing key (bindingKey)
	err := rc.Ch.QueueBind(
		queueName,   // Queue name (e.g., tv_queue, air_conditioner_queue)
		bindingKey,  // Binding key (e.g., device.tv.#, device.air_conditioner.#)
		exchange,    // Exchange name (e.g., device_events)
		false,       // NoWait
		nil,         // Arguments
	)
	if err != nil {
		return fmt.Errorf("error creating binding for queue %s with key %s: %w", queueName, bindingKey, err)
	}
	return nil
}

// Send is used to publish a payload onto an exchange with a given routingkey
func (rc RabbitClient) Send(ctx context.Context, exchange, routingKey string, options amqp.Publishing) error {
	// PublishWithDeferredConfirmWithContext will wait for server to ACK the message
	confirmation, err := rc.Ch.PublishWithDeferredConfirmWithContext(ctx,
		exchange,   // exchange
		routingKey, // routing key
		true, // mandatory
		false,   // immediate
		options, // amqp publishing struct
	)
	if err != nil {
		return err
	}
	// Blocks until ACK from Server is receieved
	log.Println(confirmation.Wait())
	return nil
}

// ConsumeEvent sets up a consumer to listen for messages from the specified queue.
func (rc RabbitClient) ConsumeEvent(queueName string) (<-chan amqp.Delivery, error) {
	messages, err := rc.Ch.Consume(
		queueName,
		"",
		true,  // AutoAck
		false, // Exclusive
		false, // NoLocal
		false, // NoWait
		nil,   // Arguments
	)
	if err != nil {
		return nil, fmt.Errorf("error consuming events from queue %s: %w", queueName, err)
	}
	return messages, nil
}

func CreateMessage(body string) amqp.Publishing {
	return amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(body),
	}
}

