package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const BTCPRICE float64 = 25000 // $25,000.00 per bitcoin

// Sha256Hash returns the SHA256 hash in hex of the input hex
func Sha256Hash(hexString string) string {
	bytes, _ := hex.DecodeString(hexString)
	// Create a new SHA256 hash
	h := sha256.New()

	// Write the input hex to the hash
	h.Write(bytes)

	return fmt.Sprintf("%x", h.Sum(nil))
}

