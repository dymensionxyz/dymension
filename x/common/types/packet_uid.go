package types

import "fmt"

// PacketUID is a unique identifier for an Rollapp IBC packet on the hub
type PacketUID struct {
	Type              Type
	RollappHubPort    string
	RollappHubChannel string
	Sequence          uint64
}

// NewPacketUID creates a new PacketUID with the provided details.
func NewPacketUID(packetType Type, hubPort string, hubChannel string, sequence uint64) PacketUID {
	return PacketUID{
		Type:              packetType,
		RollappHubPort:    hubPort,
		RollappHubChannel: hubChannel,
		Sequence:          sequence,
	}
}

// String returns a string representation of the PacketUID
func (p PacketUID) String() string {
	return fmt.Sprintf("%s-%s-%s-%d", p.Type, p.RollappHubChannel, p.RollappHubPort, p.Sequence)
}
