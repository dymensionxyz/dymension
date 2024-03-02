package types

import (
	"encoding/binary"
	fmt "fmt"
)

var _ binary.ByteOrder

const (
	//KeySeparator defines the separator for keys
	KeySeparator = "/"
	// // RollappPacketKeyPrefix is the prefix to retrieve all RollappPackets
	// RollappPacketKeyPrefix = "RollappPacket/value/"
)

var (
	// AllRollappPacketKeyPrefix is the prefix to retrieve all RollappPackets
	AllRollappPacketKeyPrefix = []byte{0x00}
	// PendingRollappPacketKeyPrefix is the prefix for pending rollapp packets
	PendingRollappPacketKeyPrefix = []byte{0x00, 0x01}
	// FinalizedRollappPacketKeyPrefix is the prefix for finalized rollapp packets
	FinalizedRollappPacketKeyPrefix = []byte{0x00, 0x02}
	// RevertedRollappPacketKeyPrefix is the prefix for reverted rollapp packets
	RevertedRollappPacketKeyPrefix = []byte{0x00, 0x03}
)

// GetRollappPacketKey constructs a key for a specific RollappPacket
func GetRollappPacketKey(
	rollappPacket *RollappPacket,
) ([]byte, error) {
	// Get the relevant key prefix based on the packet status
	var statusPrefix []byte
	switch rollappPacket.Status {
	case Status_PENDING:
		statusPrefix = PendingRollappPacketKeyPrefix
	case Status_FINALIZED:
		statusPrefix = FinalizedRollappPacketKeyPrefix
	case Status_REVERTED:
		statusPrefix = RevertedRollappPacketKeyPrefix
	default:
		return nil, fmt.Errorf("invalid packet status: %s", rollappPacket.Status)
	}
	// %020d formats the integer with leading zeros, up to a width of 20 digits.
	// This is done in order to easily iterate over the keys in order so that e.g 421 won't come after 42 instead of 43.
	// This width is chosen to accommodate the range of uint64 which can be up to 20 digits long
	// The leading zero is not limited to one, it will add as many zeros as needed to reach the width of 20 digits
	// For example, the number 342234 will be formatted as 00000000000000342234
	packetUID := fmt.Sprintf("%s-%020d", rollappPacket.Packet.DestinationChannel, rollappPacket.Packet.Sequence)
	return []byte(fmt.Sprintf("%s%s%020d%s%s%s%s", statusPrefix, KeySeparator, rollappPacket.ProofHeight,
		KeySeparator, rollappPacket.RollappId, KeySeparator, packetUID)), nil
}
