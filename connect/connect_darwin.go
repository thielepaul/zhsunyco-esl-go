package connect

import "tinygo.org/x/bluetooth"

// mac os hides the mac address, so we only check the service uuid, this is good enough as long as there is only one ESL in range.
func isThisTheDevice(device bluetooth.ScanResult, _ bluetooth.MAC) bool {
	return device.HasServiceUUID(serviceUUID)
}
