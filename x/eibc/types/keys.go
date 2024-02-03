package types

import (
	"encoding/binary"
)

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

// GetDemandOrderKey constructs a key for a specific DemandOrder.
// The key is of the form "DemandOrder/{underlying-packet-status}/{orderId}".
// The reason we add the status is that later we can clean up the non-active orders.
func GetDemandOrderKey(packetStatus string, orderId string) []byte {
	var key []byte

	key = append(key, []byte(packetStatus)...)
	key = append(key, []byte("/")...)

	key = append(key, []byte(orderId)...)
	key = append(key, []byte("/")...)

	return key
}
