package encoding

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

var allWhiteStr = "fc00000000007f012700000003000094fc80000000807f012700000003000094"
var allBlackStr = "fc00000000007f012700000003400094fc80000000807f012700000003000094"

func TestEncode(t *testing.T) {
	allZero := make([]byte, 296*128)
	allOne := make([]byte, 296*128)
	for i := range allOne {
		allOne[i] = 0xFF
	}
	allWhite, err := Encode(allZero, allZero)
	assert.NoError(t, err)
	assert.Equal(t, allWhiteStr, hex.EncodeToString(allWhite))
	allBlack, err := Encode(allOne, allZero)
	assert.NoError(t, err)
	assert.Equal(t, allBlackStr, hex.EncodeToString(allBlack))
}
