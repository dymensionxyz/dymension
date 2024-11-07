package types

const (
	// ModuleName defines the module name
	ModuleName = "delayedack"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_delayedack"
)

var PendingPacketsByAddressKeyPrefix = []byte{0x01}
