package main

import (
	"log"
	"time"

	"github.com/thielepaul/zhsunyco-esl-go/image"
	"github.com/thielepaul/zhsunyco-esl-go/protocol"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

const serviceUUIDStr = "00001523-1212-efde-1523-785feabcd123"
const writeUUIDStr = "00001525-1212-efde-1523-785feabcd123"

func main() {
	weather := []image.Weather{
		{High: 10, Low: 20, Icon: "cloudy", Text: "Hello World!"},
		{High: 10, Low: 20, Icon: "snow", Text: "Hello World!"},
		{High: 10, Low: 20, Icon: "thunderstorm", Text: "Hello World!"},
	}

	imgBytesBw, imgBytesRed, err := image.Generate(weather...)
	if err != nil {
		log.Fatalf("failed to generate image: %v", err)
	}

	packets, err := protocol.Marshal(imgBytesBw, imgBytesRed, "3D:00:00:7B:D5:F8")
	if err != nil {
		log.Fatalf("failed to marshal: %v", err)
	}

	uuid, err := bluetooth.ParseUUID("527B25EE-9E41-095F-28BE-B3BCB86A33E3")
	if err != nil {
		log.Fatalf("failed to parse UUID: %v", err)
	}
	address := bluetooth.Address{UUID: uuid}

	serviceUUID, err := bluetooth.ParseUUID(serviceUUIDStr)
	if err != nil {
		log.Fatalf("failed to parse UUID: %v", err)
	}
	writeUUID, err := bluetooth.ParseUUID(writeUUIDStr)
	if err != nil {
		log.Fatalf("failed to parse UUID: %v", err)
	}

	if err := adapter.Enable(); err != nil {
		log.Fatalf("failed to enable BLE stack: %v", err)
	}

	log.Println("starting scan")
	var foundDevice bluetooth.ScanResult
	err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if device.Address != address {
			return
		}
		foundDevice = device
		log.Println("found device:", device.Address.String(), device.RSSI, device.LocalName(), device.ServiceUUIDs())
		if err := adapter.StopScan(); err != nil {
			log.Fatalf("Failed to stop scan: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("failed to scan: %v", err)
	}

	log.Println("connecting")
	conn, err := adapter.Connect(foundDevice.Address, bluetooth.ConnectionParams{
		ConnectionTimeout: bluetooth.NewDuration(60 * time.Second),
	})
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Disconnect()

	services, err := conn.DiscoverServices([]bluetooth.UUID{serviceUUID})
	if err != nil {
		log.Fatalf("failed to discover services: %v", err)
	}
	for _, service := range services {
		log.Println("service:", service.UUID().String())
	}

	chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{writeUUID})
	if err != nil {
		log.Fatalf("failed to discover characteristics: %v", err)
	}
	for i, p := range packets {
		log.Printf("write: %v/%v", i+1, len(packets))
		_, err = chars[0].Write(p)
		if err != nil {
			log.Fatalf("failed to write: %v", err)
		}
	}
	log.Println("write success")
}
