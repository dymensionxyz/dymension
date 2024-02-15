package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	AttributeKeyRollappIdKey                = "rollapp_id"
	AttributeKeyPacketStatusKey             = "status"
	AttributeKeyPacketSourcePortKey         = "source_port"
	AttributeKeyPacketSourceChannelKey      = "source_channel"
	AttributeKeyPacketDestinationPortKey    = "destination_port"
	AttributeKeyPacketDestinationChannelKey = "destination_channel"
	AttributeKeyPacketSequence              = "packet_sequence"
	AttributeKeyPacketError                 = "error"
)

func (r RollappPacket) GetEvents() []sdk.Attribute {
	eventAttributes := []sdk.Attribute{
		sdk.NewAttribute(AttributeKeyRollappIdKey, r.RollappId),
		sdk.NewAttribute(AttributeKeyPacketStatusKey, r.Status.String()),
		sdk.NewAttribute(AttributeKeyPacketSourcePortKey, r.Packet.SourcePort),
		sdk.NewAttribute(AttributeKeyPacketSourceChannelKey, r.Packet.SourceChannel),
		sdk.NewAttribute(AttributeKeyPacketDestinationPortKey, r.Packet.DestinationPort),
		sdk.NewAttribute(AttributeKeyPacketDestinationChannelKey, r.Packet.DestinationChannel),
		sdk.NewAttribute(AttributeKeyPacketSequence, strconv.FormatUint(r.Packet.Sequence, 10)),
		sdk.NewAttribute(AttributeKeyPacketError, r.Error),
	}
	return eventAttributes
}
