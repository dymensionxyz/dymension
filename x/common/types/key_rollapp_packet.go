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

// RollappPacketKey constructs a key for a specific RollappPacket of the form:
// status/rollappID/proofHeight/packetType/packetSourceChannel/packetSequence.
//
// In order to build a packet UID We need to take the source channel + packet + packet type as the packet UID
// otherwise we're not guaranteed with uniqueness as we could have:
// Same rollapp id, same status, same proof height same sequence (as it refers to the source chain) and same channel.
// Example would be, both rollapp and hub have channel-0 and we have at the same proof height of the rollapp
// AckPacket with sequence 1 (originated on the hub) and OnRecvPacket with sequence 1 (originated on the rollapp).
// Adding the packet type guarantees uniqueness as the type differentiate the source.
func RollappPacketKey(rollappPacket *RollappPacket) []byte {
	// Get the bytes rep
	srppPrefix := RollappPacketByStatusByRollappIDByProofHeightPrefix(rollappPacket.RollappId, rollappPacket.Status, rollappPacket.ProofHeight)
	packetTypeBytes := []byte(rollappPacket.Type.String())
	packetSequenceBytes := sdk.Uint64ToBigEndian(rollappPacket.Packet.Sequence)
	packetSourceChannelBytes := []byte(rollappPacket.Packet.SourceChannel)
	// Construct the key
	result := append(srppPrefix, keySeparatorBytes...)
	result = append(result, packetTypeBytes...)
	result = append(result, keySeparatorBytes...)
	result = append(result, packetSourceChannelBytes...)
	result = append(result, keySeparatorBytes...)
	result = append(result, packetSequenceBytes...)

	return result
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
