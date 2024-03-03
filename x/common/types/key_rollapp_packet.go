package types

import (
	"encoding/binary"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

// RollappPacketKey constructs a key for a specific RollappPacket
func RollappPacketKey(
	rollappPacket *RollappPacket,
) ([]byte, error) {
	// Get the relevant key prefix based on the packet status
	statusPrefix, err := GetStatusBytes(rollappPacket.Status)
	if err != nil {
		return nil, err
	}
	// Build the key bytes repr. Convert each uint64 to big endian bytes to ensure lexicographic ordering.
	keySeparatorBytes := []byte(KeySeparator)
	rollappIdBytes := []byte(rollappPacket.RollappId)
	proofHeightBytes := sdk.Uint64ToBigEndian(rollappPacket.ProofHeight)
	// Build the packetUID from the destination channel and sequence number.
	packetSequenceBytes := sdk.Uint64ToBigEndian(rollappPacket.Packet.Sequence)
	packetDestinationChannelBytes := []byte(rollappPacket.Packet.DestinationChannel)
	packetUIDBytes := append(packetDestinationChannelBytes, packetSequenceBytes...)

	// Concatenate the byte slices directly.
	result := append(statusPrefix, keySeparatorBytes...)
	result = append(result, proofHeightBytes...)
	result = append(result, keySeparatorBytes...)
	result = append(result, rollappIdBytes...)
	result = append(result, keySeparatorBytes...)
	result = append(result, packetUIDBytes...)

	return result, nil
}

// GetStatusBytes returns the byte representation of the status
func GetStatusBytes(status Status) ([]byte, error) {
	switch status {
	case Status_PENDING:
		return PendingRollappPacketKeyPrefix, nil
	case Status_FINALIZED:
		return FinalizedRollappPacketKeyPrefix, nil
	case Status_REVERTED:
		return RevertedRollappPacketKeyPrefix, nil
	default:
		return nil, fmt.Errorf("invalid packet status: %s", status)
	}
}
