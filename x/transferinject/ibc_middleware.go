package transferinject

import (
	"errors"
	"fmt"
	. "slices"

	errorsmod "cosmossdk.io/errors"
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

type IBCMiddleware struct {
	porttypes.IBCModule
	porttypes.ICS4Wrapper

	rollappKeeper types.RollappKeeper
	bankKeeper    types.BankKeeper
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
) porttypes.ICS4Wrapper {
	return &IBCMiddleware{
		ICS4Wrapper:   ics,
		rollappKeeper: rollappKeeper,
		bankKeeper:    bankKeeper,
	}
}

// NewIBCAckMiddleware creates a new IBCModule.
// It intercepts acknowledged incoming IBC packets and adds token metadata that had just been registered on the rollapp itself,
// to the local rollapp record.
func NewIBCAckMiddleware(
	ibc porttypes.IBCModule,
	rollappKeeper types.RollappKeeper,
	bankKeeper types.BankKeeper,
) porttypes.IBCModule {
	return &IBCMiddleware{
		IBCModule:     ibc,
		rollappKeeper: rollappKeeper,
		bankKeeper:    bankKeeper,
	}
}

// SendPacket wraps IBC ChannelKeeper's SendPacket function
func (m *IBCMiddleware) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	destinationPort string, destinationChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	packet := new(transfertypes.FungibleTokenPacketData)
	if err = types.ModuleCdc.UnmarshalJSON(data, packet); err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	if _, exists := types.MemoAlreadyHasDenomMetadata(packet.Memo); exists {
		return 0, errorsmod.Wrapf(errortypes.ErrUnauthorized, "denom metadata already exists in memo")
	}

	rollapp, err := m.rollappKeeper.ExtractRollappFromChannel(ctx, destinationPort, destinationChannel)
	if err != nil {
		return 0, fmt.Errorf("extract rollapp id from packet: %w", err)
	}
	// TODO: consider that other chains also might want this feature
	if rollapp == nil {
		return m.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
	}

	if hasDenom(rollapp.TokenMetadata, packet.Denom) {
		return m.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
	}

	denomMetadata, ok := m.bankKeeper.GetDenomMetaData(ctx, packet.Denom)
	if !ok {
		return 0, errorsmod.Wrapf(errortypes.ErrNotFound, "denom metadata not found")
	}

	packet.Memo, err = types.AddDenomMetadataToMemo(packet.Memo, denomMetadata)
	if err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrInvalidType, "add denom metadata to memo: %s", err.Error())
	}

	data, err = types.ModuleCdc.MarshalJSON(packet)
	if err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrInvalidType, "marshal ICS-20 transfer packet data: %s", err.Error())
	}

	return m.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
}

// OnAcknowledgementPacket adds the token metadata to the rollapp if it doesn't exist
func (m *IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var data transfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return errorsmod.Wrapf(errortypes.ErrUnknownRequest, "unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	if len(data.Memo) == 0 {
		return m.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	packetMetaData, err := types.ParsePacketMetadata(data.Memo)
	if errors.Is(err, types.ErrMemoUnmarshal) || errors.Is(err, types.ErrMemoDenomMetadataEmpty) {
		return m.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}
	if err != nil {
		return err
	}

	rollapp, err := m.rollappKeeper.ExtractRollappFromChannel(ctx, packet.SourcePort, packet.SourceChannel)
	if err != nil {
		return fmt.Errorf("extract rollapp id from packet: %w", err)
	}
	if rollapp == nil {
		return m.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	dm := packetMetaData.DenomMetadata

	if hasDenom(rollapp.TokenMetadata, dm.Base) {
		return m.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
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

	m.rollappKeeper.SetRollapp(ctx, *rollapp)

	return m.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

func hasDenom(metadata []*rtypes.TokenMetadata, denom string) bool {
	// Check if the rollapp already contains the denom metadata by matching the base of the denom metadata.
	// At the first match, we assume that the rollapp already contains the metadata.
	// It would be technically possible to have a race condition where the denom metadata is added to the rollapp
	// from another packet before this packet is acknowledged.
	return ContainsFunc(metadata, func(dm *rtypes.TokenMetadata) bool { return dm.Base == denom })
}
