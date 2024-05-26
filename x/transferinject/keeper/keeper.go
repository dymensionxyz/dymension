package keeper

import (
	"errors"
	"fmt"
	. "slices"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"

	rtypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/transferinject/types"
)

type Keeper struct {
	cdc codec.BinaryCodec
	porttypes.IBCModule
	porttypes.ICS4Wrapper
	rollappKeeper types.RollappKeeper
	bankKeeper    types.BankKeeper
}

func NewTransferInject(
	cdc codec.BinaryCodec,
	rollappKeeper types.RollappKeeper,
	bankKeeper types.BankKeeper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		rollappKeeper: rollappKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (t *Keeper) SetMiddleware(
	ibc porttypes.IBCModule,
	ics porttypes.ICS4Wrapper,
) {
	t.IBCModule = ibc
	t.ICS4Wrapper = ics
}

// SendPacket wraps IBC ChannelKeeper's SendPacket function
func (t *Keeper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	destinationPort string, destinationChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	rollapp, err := t.rollappKeeper.ExtractRollappFromChannel(ctx, destinationPort, destinationChannel)
	if err != nil {
		return 0, fmt.Errorf("cannot extract rollapp id from packet: %w", err)
	}
	// TODO: consider that other chains also might want this feature
	if rollapp == nil {
		return t.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
	}

	packet := new(transfertypes.FungibleTokenPacketData)
	if err = types.ModuleCdc.UnmarshalJSON(data, packet); err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	// check if the rollapp already contains the denom metadata
	if ContainsFunc(rollapp.TokenMetadata, func(dm *rtypes.TokenMetadata) bool { return dm.Base == packet.Denom }) {
		return t.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
	}

	denomMetadata, ok := t.bankKeeper.GetDenomMetaData(ctx, packet.Denom)
	if !ok {
		return 0, errorsmod.Wrapf(errortypes.ErrInvalidType, "denom metadata not found")
	}

	packet.Memo, err = types.AddDenomMetadataToMemo(packet.Memo, denomMetadata)
	if err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrInvalidType, "cannot create injected value")
	}

	data, err = types.ModuleCdc.MarshalJSON(packet)
	if err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrInvalidType, "cannot marshal ICS-20 transfer packet data")
	}

	return t.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
}

// OnAcknowledgementPacket adds the token metadata to the rollapp if it doesn't exist
func (t *Keeper) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var data transfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return errorsmod.Wrapf(errortypes.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	if len(data.Memo) == 0 {
		return t.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	packetMetaData, err := types.ParsePacketMetadata(data.Memo)
	if errors.Is(err, types.ErrMemoUnmarshal) || errors.Is(err, types.ErrMemoDMEmpty) {
		return t.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}
	if err != nil {
		return err
	}

	rollapp, err := t.rollappKeeper.ExtractRollappFromChannel(ctx, packet.SourcePort, packet.SourceChannel)
	if err != nil {
		return fmt.Errorf("cannot extract rollapp id from packet: %w", err)
	}
	if rollapp == nil {
		return t.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	dm := packetMetaData.DenomMetadata

	if ContainsFunc(rollapp.TokenMetadata, func(tm *rtypes.TokenMetadata) bool { return tm.Base == dm.Base }) {
		return t.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	// check if hub has the denom metadata
	if !t.bankKeeper.HasDenomMetaData(ctx, dm.Base) {
		return errorsmod.Wrapf(errortypes.ErrInvalidType, "denom metadata not found")
	}

	denomUnits := make([]*rtypes.DenomUnit, len(dm.DenomUnits))
	for _, du := range dm.DenomUnits {
		if du.Exponent == 0 {
			continue
		}
		ndu := &rtypes.DenomUnit{
			Denom:    du.Denom,
			Exponent: du.Exponent,
			Aliases:  du.Aliases,
		}
		denomUnits = append(denomUnits, ndu)
	}

	tokenMetaData := &rtypes.TokenMetadata{
		Description: dm.Description,
		DenomUnits:  denomUnits,
		Base:        dm.Base,
		Display:     dm.Display,
		Name:        dm.Name,
		Symbol:      dm.Symbol,
		URI:         dm.URI,
		URIHash:     dm.URIHash,
	}
	// add the new token metadata to the rollapp
	rollapp.TokenMetadata = append(rollapp.TokenMetadata, tokenMetaData)

	t.rollappKeeper.SetRollapp(ctx, *rollapp)

	return t.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}
