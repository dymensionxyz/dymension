// Package transferinject module provides IBC middleware for sending and acknowledging IBC packets with injecting additional packet metadata to IBC packets.
package transferinject

import (
	. "slices"

	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"

	"github.com/dymensionxyz/dymension/v3/x/transferinject/types"
)

type IBCSendMiddleware struct {
	porttypes.ICS4Wrapper

	rollappKeeper    types.RollappKeeper
	bankKeeper       types.BankKeeper
	delayedackKeeper delayedackkeeper.Keeper
}

// NewIBCSendMiddleware creates a new ICS4Wrapper.
// It intercepts outgoing IBC packets and adds token metadata to the memo if the rollapp doesn't have it.
// This is a solution for adding token metadata to fungible tokens transferred over IBC,
// targeted at rollapps that don't have the token metadata for the token being transferred.
// More info here: https://www.notion.so/dymension/ADR-x-IBC-Denom-Metadata-Transfer-From-Hub-to-Rollapp-d3791f524ac849a9a3eb44d17968a30b
func NewIBCSendMiddleware(
	ics porttypes.ICS4Wrapper,
	rollappKeeper types.RollappKeeper,
	bankKeeper types.BankKeeper,
	delayedackKeeper delayedackkeeper.Keeper,
) *IBCSendMiddleware {
	return &IBCSendMiddleware{
		ICS4Wrapper:      ics,
		rollappKeeper:    rollappKeeper,
		bankKeeper:       bankKeeper,
		delayedackKeeper: delayedackKeeper,
	}
}

// SendPacket wraps IBC ChannelKeeper's SendPacket function
func (m *IBCSendMiddleware) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	destinationPort string, destinationChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	transfer, err := m.delayedackKeeper.GetValidTransfer(ctx)

	packet := new(transfertypes.FungibleTokenPacketData)
	if err = types.ModuleCdc.UnmarshalJSON(data, packet); err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrJSONUnmarshal, "unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	if types.MemoAlreadyHasPacketMetadata(packet.Memo) {
		return 0, types.ErrMemoTransferInjectAlreadyExists
	}

	rollapp, err := m.rollappKeeper.ExtractRollappFromChannel(ctx, destinationPort, destinationChannel)
	if err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "extract rollapp from packet: %s", err.Error())
	}

	// TODO: currently we check if receiving chain is a rollapp, consider that other chains also might want this feature
	// meaning, find a better way to check if the receiving chain supports this middleware
	if rollapp == nil {
		return m.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
	}

	if transfertypes.ReceiverChainIsSource(destinationPort, destinationChannel, packet.Denom) {
		return m.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
	}

	// Check if the rollapp already contains the denom metadata by matching the base of the denom metadata.
	// At the first match, we assume that the rollapp already contains the metadata.
	// It would be technically possible to have a race condition where the denom metadata is added to the rollapp
	// from another packet before this packet is acknowledged.
	if Contains(rollapp.RegisteredDenoms, packet.Denom) {
		return m.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
	}

	// get the denom metadata from the bank keeper, if it doesn't exist, move on to the next middleware in the chain
	denomMetadata, ok := m.bankKeeper.GetDenomMetaData(ctx, packet.Denom)
	if !ok {
		return m.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
	}

	packet.Memo, err = types.AddDenomMetadataToMemo(packet.Memo, denomMetadata)
	if err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrUnauthorized, "add denom metadata to memo: %s", err.Error())
	}

	data, err = types.ModuleCdc.MarshalJSON(packet)
	if err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrJSONMarshal, "marshal ICS-20 transfer packet data: %s", err.Error())
	}

	return m.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
}

type IBCAckMiddleware struct {
	porttypes.IBCModule

	rollappKeeper types.RollappKeeper
}

// NewIBCAckMiddleware creates a new IBCModule.
// It intercepts acknowledged incoming IBC packets and adds token metadata that had just been registered on the rollapp itself,
// to the local rollapp record.
func NewIBCAckMiddleware(
	ibc porttypes.IBCModule,
	rollappKeeper types.RollappKeeper,
) *IBCAckMiddleware {
	return &IBCAckMiddleware{
		IBCModule:     ibc,
		rollappKeeper: rollappKeeper,
	}
}

// OnAcknowledgementPacket adds the token metadata to the rollapp if it doesn't exist
func (m *IBCAckMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := types.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return errorsmod.Wrapf(errortypes.ErrJSONUnmarshal, "unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	if !ack.Success() {
		return m.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	var data transfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return errorsmod.Wrapf(errortypes.ErrJSONUnmarshal, "unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	packetMetadata, err := types.ParsePacketMetadata(data.Memo)
	if err != nil {
		return m.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	dm := packetMetadata.DenomMetadata
	if dm == nil {
		return m.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	rollapp, err := m.rollappKeeper.ExtractRollappFromChannel(ctx, packet.SourcePort, packet.SourceChannel)
	if err != nil {
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "extract rollapp from packet: %s", err.Error())
	}
	if rollapp == nil {
		return errorsmod.Wrapf(errortypes.ErrNotFound, "rollapp not found")
	}

	if !Contains(rollapp.RegisteredDenoms, dm.Base) {
		// add the new token denom base to the list of rollapp's registered denoms
		rollapp.RegisteredDenoms = append(rollapp.RegisteredDenoms, dm.Base)

		m.rollappKeeper.SetRollapp(ctx, *rollapp)
	}

	return m.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}
