package main

import (
	"fmt"
	"log"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

var wordCount = make(map[string]int)

type mqttMessage struct {
	topic   string
	payload string
}

func peripheralDiscovered(messages chan mqttMessage, p gatt.Peripheral) {
	id := p.ID()
	if _, ok := wordCount[id]; !ok {
		wordCount[id]++
		slug := strings.Replace(id, ":", "_", -1)
		topic := fmt.Sprintf("location/%s", slug)
		messages <- mqttMessage{topic, "home"}
	}
}

func stateChanged(device gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		device.Scan([]gatt.UUID{}, true)
		return
	default:
		device.StopScanning()
	}
}

func sendMqttMessages(messages chan mqttMessage) {

	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker("tcp://192.168.1.200:1883")
	mqttOpts.SetClientID("gohome")
	mqttOpts.SetUsername("hassio")
	mqttOpts.SetPassword("hassio")

	mqttClient := mqtt.NewClient(mqttOpts)
	mqttClient.Connect()

	for {
		msg := <-messages
		fmt.Printf("%s\n", msg.topic)
		mqttClient.Publish(msg.topic, 0, true, msg.payload)
	}

}

func main() {

	messages := make(chan mqttMessage)

	go sendMqttMessages(messages)

	device, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open drvice, err: %s\n", err)
	}

	device.Handle(gatt.PeripheralDiscovered(func(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
		peripheralDiscovered(messages, p)
	}))
	device.Init(stateChanged)

	select {}
}
