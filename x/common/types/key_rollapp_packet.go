package types

import (
	"encoding/binary"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ binary.ByteOrder

const (
	// keySeparator defines the separator for keys
	keySeparator = "/"
)

var (
	// PendingRollappPacketKeyPrefix is the prefix for pending rollapp packets
	PendingRollappPacketKeyPrefix = []byte{0x00, 0x01}
	// FinalizedRollappPacketKeyPrefix is the prefix for finalized rollapp packets
	FinalizedRollappPacketKeyPrefix = []byte{0x00, 0x02}
	// RevertedRollappPacketKeyPrefix is the prefix for reverted rollapp packets
	RevertedRollappPacketKeyPrefix = []byte{0x00, 0x03}
)

// RollappPacketKey constructs a key for a specific RollappPacket
func RollappPacketKey(rollappPacket *RollappPacket) []byte {
	rollappIdBytes := []byte(rollappPacket.RollappId)
	// Get the relevant key prefix based on the packet status
	statusBytes := mustGetStatusBytes(rollappPacket.Status)
	// Build the key bytes repr. Convert each uint64 to big endian bytes to ensure lexicographic ordering.
	keySeparatorBytes := []byte(keySeparator)
	proofHeightBytes := sdk.Uint64ToBigEndian(rollappPacket.ProofHeight)
	// Build the packetUID from the destination channel and sequence number.
	packetSequenceBytes := sdk.Uint64ToBigEndian(rollappPacket.Packet.Sequence)
	packetDestinationChannelBytes := []byte(rollappPacket.Packet.DestinationChannel)
	packetUIDBytes := append(packetDestinationChannelBytes, packetSequenceBytes...)

	// Concatenate the byte slices directly.
	// status/rollappID/proofHeight/packetUID
	result := append(statusBytes, keySeparatorBytes...)
	result = append(result, rollappIdBytes...)
	result = append(result, keySeparatorBytes...)
	result = append(result, proofHeightBytes...)
	result = append(result, keySeparatorBytes...)
	result = append(result, packetUIDBytes...)

	return result
}

// RollappPacketStatusAndRollappIDKey constructs a key prefix for a specific RollappPacket
// by status and rollappID:
// "status/rollappID/"
// by only status:
// "status/"
func RollappPacketStatusAndRollappIDKey(rollappPacket *RollappPacket) (result []byte) {
	rollappIdBytes := []byte(rollappPacket.RollappId)
	statusPrefix := mustGetStatusBytes(rollappPacket.Status)
	keySeparatorBytes := []byte(keySeparator)
	result = append(statusPrefix, keySeparatorBytes...)
	if rollappPacket.RollappId == "" {
		return
	}
	result = append(result, rollappIdBytes...)
	result = append(result, keySeparatorBytes...)
	return
}

// mustGetStatusBytes returns the byte representation of the status
func mustGetStatusBytes(status Status) []byte {
	switch status {
	case Status_PENDING:
		return PendingRollappPacketKeyPrefix
	case Status_FINALIZED:
		return FinalizedRollappPacketKeyPrefix
	case Status_REVERTED:
		return RevertedRollappPacketKeyPrefix
	default:
		panic(fmt.Sprintf("invalid packet status: %s", status))
	}
}
