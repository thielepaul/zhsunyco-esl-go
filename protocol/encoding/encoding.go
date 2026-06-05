package encoding

import "fmt"

const width = 296
const height = 128

func Encode(bw []byte, red []byte) ([]byte, error) {
	if len(bw) != width*height || len(red) != width*height {
		return nil, fmt.Errorf("invalid dimensions, expected %d, got bw=%d, red=%d", width*height, len(bw), len(red))
	}
	bwCompressed := rleCompress(bw)
	redCompressed := rleCompress(red)
	encoded := []byte{}
	encoded = append(encoded, bwHeader(int32(len(bwCompressed)))...)
	encoded = append(encoded, bwCompressed...)
	encoded = append(encoded, redHeader(int32(len(redCompressed)))...)
	encoded = append(encoded, redCompressed...)
	return encoded, nil
}

func rleCompress(data []byte) []byte {
	out := []byte{}
	for i := 0; i < len(data); {
		val := data[i] & 1
		run := 1
		for i+run < len(data) && data[i+run]&1 == val && run < 0xFFFF {
			run++
		}
		if run < 7 {
			b := byte(0x80)
			for k := 0; k < 7 && i+k < len(data); k++ {
				b |= (data[i+k] & 1) << (6 - k)
			}
			out = append(out, b)
			i += min(7, len(data)-i)
		} else {
			switch {
			case run <= 31:
				out = append(out, (val<<6)|byte(run))
			case run <= 255:
				out = append(out, (val<<6)|0x01, byte(run))
			default:
				out = append(out, val<<6, byte(run), byte(run>>8))
			}
			i += run
		}
	}
	return out
}

func bwHeader(length int32) []byte {
	return []byte{0xFC,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x7F, 0x01, 0x27,
		byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length)}
}

func redHeader(length int32) []byte {
	return []byte{0xFC,
		0x80, 0x00, 0x00, 0x00,
		0x80, 0x7F, 0x01, 0x27,
		byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length)}
}
