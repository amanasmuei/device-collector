package main

import (
	"log"

	"github.com/amanasmuei/device-collector.git/config"
)

func main() {
	// Initialize broker
	err := config.ConnectBroker()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Initialize database
	err = config.ConnectMariaDb()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Start the MQTT client in a separate goroutine
	config.StartMQTTListener()
}
