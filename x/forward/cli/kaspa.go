package cli

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcutil/bech32"
)

// kaspa bech32m charset
const kaspaCharset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

// H256ToKaspaAddress converts a 32-byte H256 (Hyperlane recipient) to a Kaspa address.
// Kaspa uses a bech32-like encoding with key differences from standard bech32m:
// - Separator is ':' instead of '1'
// - 40-bit checksum (8 chars) instead of standard 30-bit (6 chars)
// - Different generator polynomial and HRP expansion
// These differences require custom checksum functions (kaspaChecksum, kaspaPolymod).
func H256ToKaspaAddress(h256 []byte, mainnet bool) (string, error) {
	if len(h256) != 32 {
		return "", fmt.Errorf("invalid H256 length: expected 32, got %d", len(h256))
	}

	prefix := "kaspa"
	if !mainnet {
		prefix = "kaspatest"
	}

	// Version byte: 0 = PubKey (schnorr public key, addresses start with 'q')
	versionByte := byte(0)

	// Convert version + payload bytes to 5-bit values
	data := append([]byte{versionByte}, h256...)
	values, err := bech32.ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("convert bits: %w", err)
	}

	// Compute Kaspa 8-character checksum
	checksum := kaspaChecksum(prefix, values)

	// Encode to bech32m string
	var sb strings.Builder
	sb.WriteString(prefix)
	sb.WriteString(":")
	for _, v := range values {
		sb.WriteByte(kaspaCharset[v])
	}
	for _, v := range checksum {
		sb.WriteByte(kaspaCharset[v])
	}

	return sb.String(), nil
}

// kaspaChecksum computes the Kaspa bech32m 8-character checksum
// Kaspa uses a 40-bit checksum (8 characters) instead of the standard 30-bit (6 characters)
// It also uses a simplified HRP expansion (just hrp & 0x1f instead of standard bech32 expansion)
func kaspaChecksum(hrp string, values []byte) []byte {
	// Kaspa's simplified HRP expansion: just mask each byte to 5 bits
	hrpExp := make([]byte, 0, len(hrp)+1)
	for _, c := range hrp {
		hrpExp = append(hrpExp, byte(c)&0x1f)
	}
	hrpExp = append(hrpExp, 0) // separator

	// Concatenate: hrpExp + values + [0,0,0,0,0,0,0,0] (8 zeros for 8-char checksum)
	combined := append(hrpExp, values...)
	combined = append(combined, 0, 0, 0, 0, 0, 0, 0, 0)

	// Kaspa uses 40-bit polymod, XOR with 1 at the end
	polymod := kaspaPolymod(combined) ^ 1

	// Extract 8 checksum values (40 bits / 5 bits per char = 8 chars)
	checksum := make([]byte, 8)
	for i := 0; i < 8; i++ {
		checksum[i] = byte((polymod >> (5 * (7 - i))) & 31)
	}
	return checksum
}

// kaspaPolymod computes the Kaspa 40-bit polymod
// This is different from standard bech32 which uses 30-bit
func kaspaPolymod(values []byte) uint64 {
	// Generator polynomial coefficients for 40-bit checksum
	gen := []uint64{
		0x98f2bc8e61,
		0x79b76d99e2,
		0xf33e5fb3c4,
		0xae2eabe2a8,
		0x1e4f43e470,
	}
	chk := uint64(1)
	for _, v := range values {
		top := chk >> 35
		chk = (chk&0x07ffffffff)<<5 ^ uint64(v)
		for i := 0; i < 5; i++ {
			if (top>>i)&1 == 1 {
				chk ^= gen[i]
			}
		}
	}
	return chk
}
