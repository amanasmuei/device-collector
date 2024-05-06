package main

import "github.com/amanasmuei/device-collector.git/config"

func main() {
	// Initialize broker
	config.ConnectBroker()

	// Initialize database
	config.ConnectMariaDb()

	// Start the MQTT client in a separate goroutine
	config.StartMQTTListener()
}
