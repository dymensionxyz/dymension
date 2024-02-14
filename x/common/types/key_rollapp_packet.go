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

// GetRollappPacketKey constructs a key for a specific RollappPacket
func GetRollappPacketKey(
	rollappId string,
	status Status,
	packetProofHeight uint64,
	IBCPacket channeltypes.Packet,
) []byte {
	var key []byte

	statusBytes := []byte(fmt.Sprint(status))
	key = append(key, statusBytes...)
	key = append(key, []byte("/")...)

	rollappIdBytes := []byte(rollappId)
	key = append(key, rollappIdBytes...)
	key = append(key, []byte("/")...)

	// %020d formats the integer with leading zeros, up to a width of 20 digits.
	// This is done in order to easily iterate over the keys in order.
	// This width is chosen to accommodate the range of uint64 which can be up to 20 digits long
	// The leading zero is not limited to one, it will add as many zeros as needed to reach the width of 20 digits
	// For example, the number 342234 will be formatted as 00000000000000342234
	packetHeightBytes := []byte(fmt.Sprintf("%020d", packetProofHeight))
	key = append(key, packetHeightBytes...)
	key = append(key, []byte("/")...)

	packetUID := IBCPacket.DestinationChannel + "-" + fmt.Sprint(IBCPacket.Sequence)
	packetUIDBytes := []byte(packetUID)
	key = append(key, packetUIDBytes...)
	key = append(key, []byte("/")...)

	return key
}
