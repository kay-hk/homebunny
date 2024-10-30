package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"smart-home-assistant/internal"
)

var rabbitClient internal.RabbitClient // Global RabbitMQ client for publishing

func registerDeviceHandler(w http.ResponseWriter, r *http.Request, dbClient *internal.PostgreSQLClient) {
	var device internal.Device
	err := json.NewDecoder(r.Body).Decode(&device)
	if err != nil {
		http.Error(w, "Invalid device format", http.StatusBadRequest)
		return
	}

	err = dbClient.InsertDevice(device)
	if err != nil {
		http.Error(w, "Failed to save device", http.StatusInternalServerError)
		return
	}

	log.Printf("Device registered: %v", device)
	w.WriteHeader(http.StatusCreated)
}

func publishEventHandler(w http.ResponseWriter, r *http.Request, dbClient *internal.PostgreSQLClient) {
	var device internal.Device
	err := json.NewDecoder(r.Body).Decode(&device)
	if err != nil {
		http.Error(w, "Invalid event format", http.StatusBadRequest)
		return
	}

	routingKey := fmt.Sprintf("device.%s.%s", device.Type, device.State)
	err = rabbitClient.Send(context.TODO(), "device_events", routingKey, internal.CreateMessage(fmt.Sprintf("%v", device)))
	if err != nil {
		http.Error(w, "Failed to publish event", http.StatusInternalServerError)
		return
	}

	err = dbClient.UpdateDeviceState(device.ID, device.State)
	if err != nil {
		http.Error(w, "Failed to update device state", http.StatusInternalServerError)
		return
	}

	log.Printf("Event published and state updated for device: %s", device.Type)
	w.WriteHeader(http.StatusOK)
}

func main() {
	// Load application configuration
	appConfig, err := internal.LoadAppConfig()
	if err != nil {
		log.Fatalf("Failed to load application config: %v", err)
	}

	// Initialize RabbitMQ client with config values
	conn, err := internal.ConnectRabbitMQ(
		appConfig.RabbitMQ.User,
		appConfig.RabbitMQ.Password,
		appConfig.RabbitMQ.Host,
		appConfig.RabbitMQ.VHost,
	)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Initialize RabbitMQ client
	rabbitClient, err = internal.NewRabbitMQClient(conn)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ client: %v", err)
	}
	defer rabbitClient.Close()

	// Connect to PostgreSQL using configuration
	dbClient, err := internal.ConnectPostgreSQL(*appConfig)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer dbClient.Close()

	// Set up HTTP handlers
	http.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		registerDeviceHandler(w, r, dbClient)
	})

	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		publishEventHandler(w, r, dbClient)
	})

	// Start HTTP server
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
