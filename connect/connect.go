package connect

import (
	"fmt"
	"log"
	"time"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

const serviceUUIDStr = "00001523-1212-efde-1523-785feabcd123"
const writeUUIDStr = "00001525-1212-efde-1523-785feabcd123"

var serviceUUID bluetooth.UUID
var writeUUID bluetooth.UUID

func init() {
	var err error
	serviceUUID, err = bluetooth.ParseUUID(serviceUUIDStr)
	if err != nil {
		log.Fatalf("failed to parse UUID: %v", err)
	}
	writeUUID, err = bluetooth.ParseUUID(writeUUIDStr)
	if err != nil {
		log.Fatalf("failed to parse UUID: %v", err)
	}
}
func Update(macStr string, packets [][]byte) error {
	if err := adapter.Enable(); err != nil {
		return fmt.Errorf("failed to enable BLE stack: %v", err)
	}

	mac, err := bluetooth.ParseMAC(macStr)
	if err != nil {
		return fmt.Errorf("failed to parse MAC address: %v", err)
	}

	log.Println("starting scan")
	var foundDevice bluetooth.ScanResult
	err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if !isThisTheDevice(device, mac) {
			return
		}
		foundDevice = device
		log.Println("found device:", device.Address.String(), device.RSSI, device.LocalName(), device.ServiceUUIDs())
		if err := adapter.StopScan(); err != nil {
			log.Fatalf("Failed to stop scan: %v", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to scan: %v", err)
	}

	log.Println("connecting")
	conn, err := adapter.Connect(foundDevice.Address, bluetooth.ConnectionParams{
		ConnectionTimeout: bluetooth.NewDuration(60 * time.Second),
	})
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Disconnect()

	services, err := conn.DiscoverServices([]bluetooth.UUID{serviceUUID})
	if err != nil {
		return fmt.Errorf("failed to discover services: %v", err)
	}
	for _, service := range services {
		log.Println("service:", service.UUID().String())
	}

	chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{writeUUID})
	if err != nil {
		return fmt.Errorf("failed to discover characteristics: %v", err)
	}
	for i, p := range packets {
		log.Printf("write: %v/%v", i+1, len(packets))
		_, err = chars[0].Write(p)
		if err != nil {
			return fmt.Errorf("failed to write: %v", err)
		}
	}
	log.Println("write success")
	return nil
}
