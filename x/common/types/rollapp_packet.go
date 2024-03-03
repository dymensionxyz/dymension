package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
)

const (
	AttributeKeyRollappId                = "rollapp_id"
	AttributeKeyPacketStatus             = "status"
	AttributeKeyPacketSourcePort         = "source_port"
	AttributeKeyPacketSourceChannel      = "source_channel"
	AttributeKeyPacketDestinationPort    = "destination_port"
	AttributeKeyPacketDestinationChannel = "destination_channel"
	AttributeKeyPacketSequence           = "packet_sequence"
	AttributeKeyPacketError              = "error"
)

func (r RollappPacket) GetEvents() []sdk.Attribute {
	eventAttributes := []sdk.Attribute{
		sdk.NewAttribute(AttributeKeyRollappId, r.RollappId),
		sdk.NewAttribute(AttributeKeyPacketStatus, r.Status.String()),
		sdk.NewAttribute(AttributeKeyPacketSourcePort, r.Packet.SourcePort),
		sdk.NewAttribute(AttributeKeyPacketSourceChannel, r.Packet.SourceChannel),
		sdk.NewAttribute(AttributeKeyPacketDestinationPort, r.Packet.DestinationPort),
		sdk.NewAttribute(AttributeKeyPacketDestinationChannel, r.Packet.DestinationChannel),
		sdk.NewAttribute(AttributeKeyPacketSequence, strconv.FormatUint(r.Packet.Sequence, 10)),
		sdk.NewAttribute(AttributeKeyPacketError, r.Error),
	}
	return eventAttributes
}

func (r RollappPacket) GetTransferPacketData() (transfertypes.FungibleTokenPacketData, error) {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(r.Packet.GetData(), &data); err != nil {
		return transfertypes.FungibleTokenPacketData{}, err
	}
	return data, nil
}
