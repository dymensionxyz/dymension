package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
