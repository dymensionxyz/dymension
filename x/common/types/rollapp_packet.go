package types

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
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
	}
	if r.Error != "" {
		eventAttributes = append(eventAttributes, sdk.NewAttribute(AttributeKeyPacketError, r.Error))
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

func (r RollappPacket) RestoreOriginalTransferTarget() (RollappPacket, error) {
	transferPacketData, err := r.GetTransferPacketData()
	if err != nil {
		return r, fmt.Errorf("get transfer packet data: %w", err)
	}
	if r.OriginalTransferTarget != "" {
		transferPacketData.Receiver = r.OriginalTransferTarget
	}
	if r.OriginalTransferTarget != "" { // It can be empty if the eibc order was never fulfilled
		switch r.Type {
		case RollappPacket_ON_RECV:
			transferPacketData.Receiver = r.OriginalTransferTarget
		case RollappPacket_ON_ACK, RollappPacket_ON_TIMEOUT:
			transferPacketData.Sender = r.OriginalTransferTarget
		}
		r.Packet.Data = transferPacketData.GetBytes()
	}
	return r, nil
}
