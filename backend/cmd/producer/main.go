package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"smart-home-assistant/internal"
)

var (
	registerDeviceURL = "http://localhost:8080/devices"
	publishEventURL   = "http://localhost:8080/publish"
)

// Change the function signature to accept a Device value
func registerDevice(device internal.Device) error {
	jsonData, err := json.Marshal(device)
	if err != nil {
		return err
	}

	resp, err := http.Post(registerDeviceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to register device, status code: %d", resp.StatusCode)
	}

	return nil
}

// Function to publish an event via HTTP
func publishEvent(device internal.Device) error {
	jsonData, err := json.Marshal(device)
	if err != nil {
		return err
	}

	resp, err := http.Post(publishEventURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to publish event, status code: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	// Define a device and event for demonstration
	tv := internal.Device{
		ID:    "tv1",
		Type:  "tv",
		State: "on", // or "off" depending on the event
	}

	ac := internal.Device{
		ID:    "ac1",
		Type:  "air_conditioner",
		State: "cooling",
	}

	// Register the devices
	err := registerDevice(tv)
	if err != nil {
		log.Fatalf("Error registering device: %v", err)
	}
	err = registerDevice(ac)
	if err != nil {
		log.Fatalf("Error registering device: %v", err)
	}

	// Publish events for both devices
	err = publishEvent(tv)
	if err != nil {
		log.Fatalf("Error publishing TV event: %v", err)
	}
	err = publishEvent(ac)
	if err != nil {
		log.Fatalf("Error publishing AC event: %v", err)
	}

	log.Println("Events published for TV and AC devices.")
}
