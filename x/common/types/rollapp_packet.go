package types

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
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
	var pd transfertypes.FungibleTokenPacketData
	if len(r.Packet.Data) != 0 {
		// It's okay if we can't get packet data
		pd, _ = r.GetTransferPacketData()
	}

	acknowledgement := "none"
	if len(r.Acknowledgement) != 0 {
		ack, err := r.GetAck()
		// It's okay if we can't get acknowledgement
		if err == nil {
			switch ack.GetResponse().(type) {
			case *channeltypes.Acknowledgement_Result:
				acknowledgement = "success"
			case *channeltypes.Acknowledgement_Error:
				acknowledgement = "error"
			}
		}
	}

	eventAttributes := []sdk.Attribute{
		sdk.NewAttribute(AttributeKeyRollappId, r.RollappId),
		sdk.NewAttribute(AttributeKeyPacketStatus, r.Status.String()),
		sdk.NewAttribute(AttributeKeyPacketSourcePort, r.Packet.SourcePort),
		sdk.NewAttribute(AttributeKeyPacketSourceChannel, r.Packet.SourceChannel),
		sdk.NewAttribute(AttributeKeyPacketDestinationPort, r.Packet.DestinationPort),
		sdk.NewAttribute(AttributeKeyPacketDestinationChannel, r.Packet.DestinationChannel),
		sdk.NewAttribute(AttributeKeyPacketSequence, strconv.FormatUint(r.Packet.Sequence, 10)),
		sdk.NewAttribute(AttributeKeyPacketProofHeight, strconv.FormatUint(r.ProofHeight, 10)),
		sdk.NewAttribute(AttributeKeyPacketType, r.Type.String()),
		sdk.NewAttribute(AttributeKeyPacketAcknowledgement, acknowledgement),
		sdk.NewAttribute(AttributeKeyPacketDataDenom, pd.Denom),
		sdk.NewAttribute(AttributeKeyPacketDataAmount, pd.Amount),
		sdk.NewAttribute(AttributeKeyPacketDataSender, pd.Sender),
		sdk.NewAttribute(AttributeKeyPacketDataReceiver, pd.Receiver),
		sdk.NewAttribute(AttributeKeyPacketDataMemo, pd.Memo),
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

func (r RollappPacket) MustGetTransferPacketData() transfertypes.FungibleTokenPacketData {
	data, err := r.GetTransferPacketData()
	if err != nil {
		panic(err)
	}
	return data
}

func (r RollappPacket) GetAck() (channeltypes.Acknowledgement, error) {
	var ack channeltypes.Acknowledgement
	if err := transfertypes.ModuleCdc.UnmarshalJSON(r.Acknowledgement, &ack); err != nil {
		return channeltypes.Acknowledgement{}, err
	}
	return ack, nil
}

func (r RollappPacket) RestoreOriginalTransferTarget() RollappPacket {
	transferPacketData := r.MustGetTransferPacketData()
	if r.OriginalTransferTarget != "" { // It can be empty if the eibc order was never fulfilled
		switch r.Type {
		case RollappPacket_ON_RECV:
			transferPacketData.Receiver = r.OriginalTransferTarget
		case RollappPacket_ON_ACK, RollappPacket_ON_TIMEOUT:
			transferPacketData.Sender = r.OriginalTransferTarget
		}
		r.Packet.Data = transferPacketData.GetBytes()
	}
	return r
}
