package types

import fmt "fmt"

// PacketUID is a unique identifier for an Rollapp IBC packet on the hub
type PacketUID struct {
	Type              RollappPacket_Type
	RollappHubPort    string
	RollappHubChannel string
	Sequence          uint64
}

func (p PacketUID) String() string {
	return fmt.Sprintf("%s-%s-%s-%d", p.Type, p.RollappHubChannel, p.RollappHubPort, p.Sequence)
}

// NewPacketUID creates a new PacketUID with the provided details.
func NewPacketUID(packetType RollappPacket_Type, hubPort string, hubChannel string, sequence uint64) PacketUID {
	return PacketUID{
		Type:              packetType,
		RollappHubPort:    hubPort,
		RollappHubChannel: hubChannel,
		Sequence:          sequence,
	}
}
