package types

import (
	"encoding/binary"
	"fmt"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

var _ binary.ByteOrder

const (
	// ModuleName defines the module name
	ModuleName = "eibc"

	// KeySeparator defines the separator for keys
	KeySeparator = "/"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_eibc"
)

// Store Key Prefixes
var (
	// AllDemandOrdersKeyPrefix is the prefix for all demand orders
	AllDemandOrdersKeyPrefix = []byte{0x00}
	// PendingDemandOrderKeyPrefix is the prefix for pending demand orders
	PendingDemandOrderKeyPrefix = []byte{0x00, 0x01}
	// FinalizedDemandOrderKeyPrefix is the prefix for finalized demand orders
	FinalizedDemandOrderKeyPrefix = []byte{0x00, 0x02}
	// RevertedDemandOrderKeyPrefix is the prefix for reverted demand orders
	RevertedDemandOrderKeyPrefix = []byte{0x00, 0x03}
)

// GetDemandOrderKey constructs a key for a specific DemandOrder.
func GetDemandOrderKey(packetStatus commontypes.Status, orderId string) ([]byte, error) {
	// Get the relevant key prefix based on the packet status
	var prefix []byte
	switch packetStatus {
	case commontypes.Status_PENDING:
		prefix = PendingDemandOrderKeyPrefix
	case commontypes.Status_FINALIZED:
		prefix = FinalizedDemandOrderKeyPrefix
	case commontypes.Status_REVERTED:
		prefix = RevertedDemandOrderKeyPrefix
	default:
		return nil, fmt.Errorf("invalid packet status: %s", packetStatus)
	}
	return []byte(fmt.Sprintf("%s%s%s%s%s", prefix, KeySeparator, packetStatus, KeySeparator, orderId)), nil
}
