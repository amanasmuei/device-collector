package main

import "device-collector/config"

func main() {
	// Initialize broker
	config.ConnectBroker()

	// Initialize database
	config.ConnectMariaDb()

	// Start the MQTT client in a separate goroutine
	config.StartMQTTListener()
}
