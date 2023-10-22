package types

import (
	"encoding/binary"
	fmt "fmt"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
)

var _ binary.ByteOrder

const (
	// RollappPacketKeyPrefix is the prefix to retrieve all RollappPackets
	RollappPacketKeyPrefix = "RollappPacket/value/"
)

// StateInfoKey returns the store key to retrieve a StateInfo from the index fields
func RollappPacketKey(
	rollappId string,
	packetHeight int64,
	IBCPacket channeltypes.Packet,
) []byte {
	var key []byte

	rollappIdBytes := []byte(rollappId)
	key = append(key, rollappIdBytes...)
	key = append(key, []byte("/")...)

	packetHeightBytes := []byte(fmt.Sprint(packetHeight))
	key = append(key, packetHeightBytes...)
	key = append(key, []byte("/")...)

	packetUID := IBCPacket.SourceChannel + "-" + IBCPacket.DestinationChannel + "-" + fmt.Sprint(IBCPacket.Sequence)
	packetUIDBytes := []byte(packetUID)
	key = append(key, packetUIDBytes...)
	key = append(key, []byte("/")...)

	return key
}
