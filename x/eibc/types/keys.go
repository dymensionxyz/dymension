package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// ModuleName defines the module name
	ModuleName = "eibc"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_eibc"
)

// Store Key Prefixes
const (
	DemandOrderKeyPrefix = "DemandOrder/value/"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

func GetDemandOrderKey(orderId string) []byte {
	var key []byte

	key = append(key, DemandOrderKeyPrefix...)
	key = append(key, []byte(orderId)...)

	return key
}
