package types

import (
	"encoding/binary"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ binary.ByteOrder

var (
	// AllRollappPacketKeyPrefix is the prefix to retrieve all RollappPackets
	AllRollappPacketKeyPrefix = []byte{0x00}
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
// status/rollappID/proofHeight/packetUID
func RollappPacketKey(rollappPacket *RollappPacket) []byte {
	srppPrefix := RollappPacketByStatusByRollappIDByProofHeightPrefix(rollappPacket.RollappId, rollappPacket.Status, rollappPacket.ProofHeight)
	packetSequenceBytes := sdk.Uint64ToBigEndian(rollappPacket.Packet.Sequence)
	packetDestinationChannelBytes := []byte(rollappPacket.Packet.DestinationChannel)
	packetUIDBytes := append(packetDestinationChannelBytes, packetSequenceBytes...)
	result := append(srppPrefix, keySeparatorBytes...)
	return append(result, packetUIDBytes...)
}

// RollappPacketByStatusByRollappIDByProofHeightPrefix constructs a key prefix for a specific RollappPacket
// by rollappID, status and proofHeight:
// "rollappID/status/proofHeight"
func RollappPacketByStatusByRollappIDByProofHeightPrefix(rollappID string, status Status, proofHeight uint64) []byte {
	return append(RollappPacketByStatusByRollappIDPrefix(status, rollappID), sdk.Uint64ToBigEndian(proofHeight)...)
}

// RollappPacketByStatusByRollappIDPrefix constructs a key prefix for a specific RollappPacket
// by status and rollappID:
// "status/rollappID/"
func RollappPacketByStatusByRollappIDPrefix(status Status, rollappID string) (result []byte) {
	return append(RollappPacketByStatusPrefix(status), RollappPacketByRollappIDPrefix(rollappID)...)
}

// RollappPacketByRollappIDPrefix constructs a key prefix for a specific RollappPacket
// by rollappID: "rollappID/"
func RollappPacketByRollappIDPrefix(rollappID string) []byte {
	return append([]byte(rollappID), keySeparatorBytes...)
}

// RollappPacketByStatusPrefix constructs a key prefix for a specific RollappPacket
// by status: "status/"
func RollappPacketByStatusPrefix(status Status) []byte {
	return append(MustGetStatusBytes(status), keySeparatorBytes...)
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
