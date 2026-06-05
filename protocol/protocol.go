package protocol

import (
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	"github.com/thielepaul/zhsunyco-esl-go/protocol/encoding"
)

func Marshal(dataBw []byte, dataRed []byte, macAddrStr string) ([][]byte, error) {
	macAddr, err := parseMac(macAddrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mac address: %s", macAddrStr)
	}
	encoded, err := encoding.Encode(dataBw, dataRed)
	if err != nil {
		return nil, err
	}
	return append(
		[][]byte{marshalInit(uint32(len(encoded)), macAddr)},
		marshalPayload(encoded, macAddr)...,
	), nil
}

func marshalPayload(data []byte, macAddr []byte) [][]byte {
	packets := [][]byte{}
	for i := 0; i < len(data); i += 200 {
		packet := make([]byte, 204)
		seq := uint16(i/200 + 1)
		packet[0] = byte(seq >> 8)
		packet[1] = byte(seq)
		copy(packet[2:], data[i:min(i+200, len(data))])
		crc := crc16(packet[:202])
		packet[202] = crc[0]
		packet[203] = crc[1]
		xorMagic(packet, macAddr)
		packets = append(packets, packet)
	}
	return packets
}

func marshalInit(length uint32, macAddr []byte) []byte {
	count := uint16(math.Ceil(float64(length) / 200.0))
	init := make([]byte, 0, 20)
	init = append(init, 0xFF, 0xFC)
	init = append(init, 'e', 'a', 's', 'y', 'T', 'a', 'g')
	init = append(init, 0x5C)
	init = append(init, byte(length>>24), byte(length>>16), byte(length>>8), byte(length))
	init = append(init, byte(count>>8), byte(count))
	init = append(init, 'B', 'T')
	init = append(init, crc16(init)...)
	xorMagic(init, macAddr)
	return init
}

func crc16(data []byte) []byte {
	crc := uint16(0xFFFF)
	for _, byteVal := range data {
		crc ^= uint16(byteVal) << 8
		for range 8 {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x8005
			} else {
				crc <<= 1
			}
		}
	}
	return []byte{byte(crc >> 8), byte(crc)}
}

func xorMagic(data []byte, macAddr []byte) {
	magic := byte(0x63)
	for _, b := range macAddr {
		magic ^= b
	}
	for i := range data {
		if len(data) == 20 && i == 9 {
			continue // this byte in the init header is not XORed
		}
		data[i] ^= magic
	}
}

func parseMac(macStr string) ([]byte, error) {
	return hex.DecodeString(strings.ReplaceAll(macStr, ":", ""))
}
