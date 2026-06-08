package connect

import (
	"bytes"

	"tinygo.org/x/bluetooth"
)

func isThisTheDevice(device bluetooth.ScanResult, mac bluetooth.MAC) bool {
	return bytes.Equal(device.Address.MAC[:], mac[:]) && device.HasServiceUUID(serviceUUID)
}
