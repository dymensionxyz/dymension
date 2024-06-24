package denommetadata

import (
	. "slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	utilsibc "github.com/dymensionxyz/dymension/v3/utils/ibc"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

var _ porttypes.IBCModule = &IBCModule{}

// IBCModule implements the ICS26 callbacks for the transfer middleware
type IBCModule struct {
	porttypes.IBCModule
	keeper        types.DenomMetadataKeeper
	rollappKeeper types.RollappKeeper
}

// NewIBCModule creates a new IBCModule given the keepers and underlying application
func NewIBCModule(
	app porttypes.IBCModule,
	keeper types.DenomMetadataKeeper,
	rollappKeeper types.RollappKeeper,
) IBCModule {
	return IBCModule{
		IBCModule:     app,
		keeper:        keeper,
		rollappKeeper: rollappKeeper,
	}
}

// OnRecvPacket registers the denom metadata if it does not exist.
// It will intercept an incoming packet and check if the denom metadata exists.
// If it does not, it will register the denom metadata.
// The handler will expect a 'denom_metadata' object in the memo field of the packet.
// If the memo is not an object, or does not contain the metadata, it moves on to the next handler.
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	if commontypes.SkipRollappMiddleware(ctx) {
		return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	transferData, err := im.rollappKeeper.GetValidTransfer(ctx, packet.Data, packet.DestinationPort, packet.DestinationChannel)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	rollapp, packetData := transferData.Rollapp, transferData.FungibleTokenPacketData

	if packetData.Memo == "" {
		return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	// at this point it's safe to assume that we are not handling a native token of the rollapp
	denomTrace := utilsibc.GetForeignDenomTrace(packet.GetDestChannel(), packetData.Denom)
	ibcDenom := denomTrace.IBCDenom()

	dm := types.ParsePacketMetadata(packetData.Memo)
	if dm == nil {
		return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	if err = dm.Validate(); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if dm.Base != packetData.Denom {
		return channeltypes.NewErrorAcknowledgement(gerrc.ErrInvalidArgument)
	}

	// if denom metadata was found in the memo, it means we should have the rollapp record
	if rollapp == nil {
		return channeltypes.NewErrorAcknowledgement(gerrc.ErrNotFound)
	}

	dm.Base = ibcDenom
	dm.DenomUnits[0].Denom = dm.Base

	if err = im.keeper.CreateDenomMetadata(ctx, *dm); err != nil {
		if errorsmod.IsOf(err, gerrc.ErrAlreadyExist) {
			return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
		}
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if !Contains(rollapp.RegisteredDenoms, dm.Base) {
		// add the new token denom base to the list of rollapp's registered denoms
		// this is to prevent the same denom metadata from getting sent to the rollapp more than once
		rollapp.RegisteredDenoms = append(rollapp.RegisteredDenoms, dm.Base)

		im.rollappKeeper.SetRollapp(ctx, *rollapp)
	}

	return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket adds the token metadata to the rollapp if it doesn't exist
func (im IBCModule) OnAcknowledgementPacket(
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
		return im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	transferData, err := im.rollappKeeper.GetValidTransfer(ctx, packet.Data, packet.DestinationPort, packet.DestinationChannel)
	if err != nil {
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "get valid transfer data: %s", err.Error())
	}

	rollapp, packetData := transferData.Rollapp, transferData.FungibleTokenPacketData

	if packetData.Memo == "" {
		return im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	dm := types.ParsePacketMetadata(packetData.Memo)
	if dm == nil {
		return im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	if err = dm.Validate(); err != nil {
		return im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	// if denom metadata was found in the memo, it means we should have the rollapp record
	if rollapp == nil {
		return gerrc.ErrNotFound
	}

	if !Contains(rollapp.RegisteredDenoms, dm.Base) {
		// add the new token denom base to the list of rollapp's registered denoms
		rollapp.RegisteredDenoms = append(rollapp.RegisteredDenoms, dm.Base)

		im.rollappKeeper.SetRollapp(ctx, *rollapp)
	}

	return im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// ICS4Wrapper intercepts outgoing IBC packets and adds token metadata to the memo if the rollapp doesn't have it.
// This is a solution for adding token metadata to fungible tokens transferred over IBC,
// targeted at rollapps that don't have the token metadata for the token being transferred.
// More info here: https://www.notion.so/dymension/ADR-x-IBC-Denom-Metadata-Transfer-From-Hub-to-Rollapp-d3791f524ac849a9a3eb44d17968a30b
type ICS4Wrapper struct {
	porttypes.ICS4Wrapper

	rollappKeeper types.RollappKeeper
	bankKeeper    types.BankKeeper
}

// NewICS4Wrapper creates a new ICS4Wrapper
func NewICS4Wrapper(
	ics porttypes.ICS4Wrapper,
	rollappKeeper types.RollappKeeper,
	bankKeeper types.BankKeeper,
) *ICS4Wrapper {
	return &ICS4Wrapper{
		ICS4Wrapper:   ics,
		rollappKeeper: rollappKeeper,
		bankKeeper:    bankKeeper,
	}
}

// SendPacket wraps IBC ChannelKeeper's SendPacket function
func (m *ICS4Wrapper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	destinationPort string, destinationChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	packet := new(transfertypes.FungibleTokenPacketData)
	if err = types.ModuleCdc.UnmarshalJSON(data, packet); err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrJSONUnmarshal, "unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	if types.MemoHasPacketMetadata(packet.Memo) {
		return 0, types.ErrMemoDenomMetadataAlreadyExists
	}

	transferData, err := m.rollappKeeper.GetValidTransfer(ctx, data, destinationPort, destinationChannel)
	if err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "get valid transfer data: %s", err.Error())
	}

	rollapp := transferData.Rollapp
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
		return 0, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "add denom metadata to memo: %s", err.Error())
	}

	data, err = types.ModuleCdc.MarshalJSON(packet)
	if err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrJSONMarshal, "marshal ICS-20 transfer packet data: %s", err.Error())
	}

	return m.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
}
