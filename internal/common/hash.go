package common

import (
	"crypto/sha256"
	"encoding/hex"
)

// computes sha256 hash and returns hex string.
func SHA256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
