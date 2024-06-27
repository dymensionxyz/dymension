package types

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

func (r RollappPacket) LogString() string {
	return fmt.Sprintf("RollappPacket{%s, %s, %s, %s, %s, %d, %s, %d}",
		r.RollappId, r.Type, r.Status, r.Packet.SourcePort, r.Packet.SourceChannel, r.Packet.Sequence, r.Error, r.ProofHeight)
}

func (r RollappPacket) ValidateBasic() error {
	if r.RollappId == "" {
		return fmt.Errorf("rollapp id cannot be empty")
	}
	if len(r.Relayer) == 0 {
		return fmt.Errorf("status cannot be empty")
	}
	if r.OriginalTransferTarget != "" {
		if _, err := sdk.AccAddressFromBech32(r.OriginalTransferTarget); err != nil {
			return fmt.Errorf("original transfer target: %w", err)
		}
	}
	if r.ProofHeight == 0 {
		return fmt.Errorf("proof height revision height cannot be zero")
	}
	if r.Packet == nil {
		return fmt.Errorf("packet cannot be nil")
	}
	if err := r.Packet.ValidateBasic(); err != nil {
		return fmt.Errorf("packet: %w", err)
	}
	return nil
}

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
