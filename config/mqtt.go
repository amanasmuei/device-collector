package config

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var MqttClient mqtt.Client

func ConnectBroker() error {

	const (
		MQTTBroker     = "tcp://49.236.203.211:1883"
		MQTTTopic      = "node_data"
		MQTTClientID   = "collector"
		MQTTUsername   = "broker"
		MQTTPassword   = "Ottabroker2024!"
		MQTTStatusPath = "status"
		MQTTDataPath   = "data"
	)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(MQTTBroker)
	opts.SetClientID(MQTTClientID)
	opts.SetUsername(MQTTUsername)
	opts.SetPassword(MQTTPassword)

	MqttClient = mqtt.NewClient(opts)
	if token := MqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	return nil
}

// Define a struct to represent the MQTT message payload
type MQTTStatusMessage struct {
	Status      string `json:"status"`
	Type        string `json:"type"`
	TimeStatus  string `json:"time_status"`
	NodeName    string `json:"node_name"`
	Temperature string `json:"temperature"`
	Humidity    string `json:"humidity"`
}

func onStatusMessageReceived(client mqtt.Client, message mqtt.Message) {
	fmt.Printf("Received status message on topic: %s\n", message.Topic())
	fmt.Printf("Message status payload: %s\n", message.Payload())

	// Parse the JSON payload into MQTTMessage struct
	var msg MQTTStatusMessage
	if err := json.Unmarshal(message.Payload(), &msg); err != nil {
		log.Println("Error parsing JSON payload:", err)
		return
	}

	// Add created_at field with current timestamp in desired format
	createdAt := time.Now().Format("2006-01-02 15:04:05")
	updatedAt := createdAt

	// Prepare the SQL INSERT statement
	stmt, err := DbSql.Prepare("INSERT INTO data_status (node_id, raw_data, created_at, updated_at) VALUES ($1, $2, $3, $4)")
	if err != nil {
		log.Println("Error preparing SQL statement:", err)
		return
	}
	defer stmt.Close()

	// Convert the MQTTMessage struct to JSON
	rawJSON, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}

	// Execute the SQL INSERT statement
	_, err = stmt.Exec(msg.NodeName, rawJSON, createdAt, updatedAt)
	if err != nil {
		log.Println("Error executing SQL statement:", err)
		return
	}
}

type PreviousData struct {
	NodeName        string
	PreviousStatus1 int
	PreviousStatus2 int
}

type MQTTDataMessage struct {
	TimeData      string `json:"time_data"`
	StatusSensor1 int    `json:"status_sensor_1"`
	StatusSensor2 int    `json:"status_sensor_2"`
	StatusMachine int    `json:"status_machine"`
}

var previousDataMap map[string]*PreviousData

func onDataMessageReceived(client mqtt.Client, message mqtt.Message) {
	fmt.Printf("Received data message on topic: %s\n", message.Topic())
	fmt.Printf("Message data payload: %s\n", message.Payload())

	// Split the message using "/" as the separator
	parts := strings.Split(message.Topic(), "/")

	// get node name
	var nodeName string
	if len(parts) >= 2 {
		nodeName = parts[1]
	} else {
		fmt.Println("Invalid message format")
		return
	}

	// Parse the JSON payload into MQTTDataMessage struct
	var msg MQTTDataMessage
	if err := json.Unmarshal(message.Payload(), &msg); err != nil {
		log.Println("Error parsing JSON payload:", err)
		return
	}

	// Add created_at field with current timestamp in desired format
	createdAt := time.Now().Format("2006-01-02 15:04:05")
	updatedAt := createdAt

	// Prepare the SQL INSERT statement for data_raw
	stmt, err := DbSql.Prepare("INSERT INTO data_raw (node_id, raw_data, created_at, updated_at) VALUES ($1, $2, $3, $4)")
	if err != nil {
		log.Println("Error preparing SQL statement:", err)
		return
	}
	defer stmt.Close()

	msg.TimeData = time.Now().Format("2006-01-02 15:04:05")
	// Convert the MQTTDataMessage struct to JSON
	rawJSON, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}

	// Execute the SQL INSERT statement for data_raw
	if _, err := stmt.Exec(nodeName, rawJSON, createdAt, updatedAt); err != nil {
		log.Println("Error executing SQL statement:", err)
		return
	}

	// Initialize previousDataMap if it's nil
	if previousDataMap == nil {
		previousDataMap = make(map[string]*PreviousData)
	}

	// Check if previous data exists in the cache
	prevData, ok := previousDataMap[nodeName]
	if !ok {
		prevData = &PreviousData{
			NodeName:        nodeName,
			PreviousStatus1: msg.StatusSensor1,
			PreviousStatus2: msg.StatusSensor2,
		}
		previousDataMap[nodeName] = prevData
	}

	if prevData.PreviousStatus1 != msg.StatusSensor1 {
		// Prepare and execute SQL INSERT statements for sensor data (assuming StatusSensor1 and StatusSensor2 are string fields)
		if err := insertSensorData(DbSql, "data_sensor_1", nodeName, msg.StatusSensor1, createdAt, updatedAt); err != nil {
			log.Println("Error inserting sensor data:", err)
			return
		}
		prevData.PreviousStatus1 = msg.StatusSensor1
	}

	if prevData.PreviousStatus2 != msg.StatusSensor2 {
		// Prepare and execute SQL INSERT statements for sensor data (assuming StatusSensor1 and StatusSensor2 are string fields)
		if err := insertSensorData(DbSql, "data_sensor_2", nodeName, msg.StatusSensor2, createdAt, updatedAt); err != nil {
			log.Println("Error inserting sensor data:", err)
			return
		}
		prevData.PreviousStatus2 = msg.StatusSensor2
	}

	// Update the cache
	previousDataMap[nodeName] = prevData

	// if err := insertSensorData(DbSql, "data_sensor_1", nodeName, msg.StatusSensor1, createdAt, updatedAt); err != nil {
	// 	log.Println("Error inserting sensor data:", err)
	// 	return
	// }

	// if err := insertSensorData(DbSql, "data_sensor_2", nodeName, msg.StatusSensor2, createdAt, updatedAt); err != nil {
	// 	log.Println("Error inserting sensor data:", err)
	// 	return
	// }
}

// Function to insert sensor data into the specified table
func insertSensorData(db *sql.DB, tableName, nodeName string, data int, createdAt, updatedAt string) error {
	stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO %s (node_id, data, created_at, updated_at) VALUES ($1, $2, $3, $4)", tableName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(nodeName, data, createdAt, updatedAt); err != nil {
		return err
	}

	return nil
}

func StartMQTTListener() {

	// Subscribe to MQTT topic
	statusTopic := fmt.Sprintf("%s/+/status/#", "node_data")
	if token := MqttClient.Subscribe(statusTopic, 0, onStatusMessageReceived); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	// Subscribe to MQTT topic
	dataTopic := fmt.Sprintf("%s/+/data/#", "node_data")
	if token := MqttClient.Subscribe(dataTopic, 0, onDataMessageReceived); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	// Keep the MQTT connection alive
	for {
		time.Sleep(1 * time.Second)
	}
}
