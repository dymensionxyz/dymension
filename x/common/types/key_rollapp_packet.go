package types

import (
	"encoding/binary"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ binary.ByteOrder

var (
	RollappPacketKeyPrefix      = []byte{0x21}
	RollappPacketIndexKeyPrefix = []byte{0x41}
	// PendingRollappPacketKeyPrefix is the prefix for pending rollapp packets
	PendingRollappPacketKeyPrefix = []byte{0x00, 0x01}
	// FinalizedRollappPacketKeyPrefix is the prefix for finalized rollapp packets
	FinalizedRollappPacketKeyPrefix = []byte{0x00, 0x02}
	// RevertedRollappPacketKeyPrefix is the prefix for reverted rollapp packets
	RevertedRollappPacketKeyPrefix = []byte{0x00, 0x03}
	// keySeparatorBytes is used to separate the rollapp packet key parts
	keySeparatorBytes = []byte("/")
)

// RollappPacketKey constructs a key for a specific RollappPacket
func RollappPacketKey(rollappPacket *RollappPacket) []byte {
	// ^rollappID/status/proofHeight/packetUID
	prefix := RollappPacketByRollappIDByStatusPrefix(rollappPacket.RollappId, rollappPacket.Status)
	suffix := GetRollappPacketKeySuffix(rollappPacket)
	return append(prefix, suffix...)
}

// RollappPacketIndexKey constructs an index key for a specific RollappPacket
func RollappPacketIndexKey(rollappPacket *RollappPacket) []byte {
	// *status/rollappID/proofHeight/packetUID
	prefix := RollappPacketByStatusByRollappIDPrefix(rollappPacket.Status, rollappPacket.RollappId)
	suffix := GetRollappPacketKeySuffix(rollappPacket)
	return append(prefix, suffix...)
}

func GetRollappPacketKeySuffix(rollappPacket *RollappPacket) (result []byte) {
	proofHeightBytes := sdk.Uint64ToBigEndian(rollappPacket.ProofHeight)
	// Build the packetUID from the destination channel and sequence number.
	packetSequenceBytes := sdk.Uint64ToBigEndian(rollappPacket.Packet.Sequence)
	packetDestinationChannelBytes := []byte(rollappPacket.Packet.DestinationChannel)
	packetUIDBytes := append(packetDestinationChannelBytes, packetSequenceBytes...)
	result = append(result, keySeparatorBytes...)
	result = append(result, proofHeightBytes...)
	result = append(result, keySeparatorBytes...)
	return append(result, packetUIDBytes...)
}

func RollappPacketByRollappIDByStatusByMaxProofHeightPrefixes(rollappID string, status Status, proofHeight uint64) ([]byte, []byte) {
	return RollappPacketByRollappIDByStatusByProofHeightPrefix(rollappID, status, 0),
		RollappPacketByRollappIDByStatusByProofHeightPrefix(rollappID, status, proofHeight+1) // inclusive end
}

func RollappPacketByStatusByRollappIDPrefix(status Status, rollappID string) []byte {
	statusBytes := RollappPacketByStatusIndexPrefix(status)
	rollappIDBytes := []byte(rollappID)
	prefix := append(statusBytes, keySeparatorBytes...)
	return append(prefix, rollappIDBytes...)
}

func RollappPacketByRollappIDByStatusByProofHeightPrefix(rollappID string, status Status, proofHeight uint64) []byte {
	proofHeightBytes := sdk.Uint64ToBigEndian(proofHeight)
	byRollappIDByStatusPrefix := RollappPacketByRollappIDByStatusPrefix(rollappID, status)
	byRollappIDByStatusPrefix = append(byRollappIDByStatusPrefix, keySeparatorBytes...)
	return append(byRollappIDByStatusPrefix, proofHeightBytes...)
}

func RollappPacketByRollappIDByStatusPrefix(rollappID string, status Status) []byte {
	rollappIDBytes := RollappPacketByRollappIDPrefix(rollappID)
	statusBytes := MustGetStatusBytes(status)
	prefix := append(rollappIDBytes, keySeparatorBytes...)
	return append(prefix, statusBytes...)
}

func RollappPacketByRollappIDPrefix(rollappID string) []byte {
	return append(RollappPacketKeyPrefix, []byte(rollappID)...)
}

func RollappPacketByStatusIndexPrefix(status Status) []byte {
	return append(RollappPacketIndexKeyPrefix, MustGetStatusBytes(status)...)
}

// MustGetStatusBytes returns the byte representation of the status
func MustGetStatusBytes(status Status) []byte {
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
