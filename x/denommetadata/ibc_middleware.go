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

	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

var _ porttypes.IBCModule = &IBCRecvMiddleware{}

// IBCRecvMiddleware implements the ICS26 callbacks for the transfer middleware
type IBCRecvMiddleware struct {
	porttypes.IBCModule
	transferKeeper types.TransferKeeper
	keeper         types.DenomMetadataKeeper
	rollappKeeper  types.RollappKeeper
}

// NewIBCRecvMiddleware creates a new IBCMiddleware given the keeper and underlying application
func NewIBCRecvMiddleware(
	app porttypes.IBCModule,
	transferKeeper types.TransferKeeper,
	keeper types.DenomMetadataKeeper,
	rollappKeeper types.RollappKeeper,
) IBCRecvMiddleware {
	return IBCRecvMiddleware{
		IBCModule:      app,
		transferKeeper: transferKeeper,
		keeper:         keeper,
		rollappKeeper:  rollappKeeper,
	}
}

// OnRecvPacket registers the denom metadata if it does not exist.
// It will intercept an incoming packet and check if the denom metadata exists.
// If it does not, it will register the denom metadata.
// The handler will expect a 'transferinject' object in the memo field of the packet.
// If the memo is not an object, or does not contain the metadata, it moves on to the next handler.
func (im IBCRecvMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	packetData := new(transfertypes.FungibleTokenPacketData)
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), packetData); err != nil {
		err = errorsmod.Wrapf(errortypes.ErrJSONUnmarshal, "unmarshal ICS-20 transfer packet data")
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if packetData.Memo == "" {
		return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	// at this point it's safe to assume that we are not handling a native token of the rollapp
	denomTrace := utils.GetForeignDenomTrace(packet.GetDestChannel(), packetData.Denom)
	ibcDenom := denomTrace.IBCDenom()

	dm := types.ParsePacketMetadata(packetData.Memo)
	if dm == nil {
		return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	if err := dm.Validate(); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	dm.Base = ibcDenom
	dm.DenomUnits[0].Denom = dm.Base

	if err := im.keeper.CreateDenomMetadata(ctx, *dm); err != nil {
		if errorsmod.IsOf(err, types.ErrDenomAlreadyExists) {
			return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
		}
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if !im.transferKeeper.HasDenomTrace(ctx, denomTrace.Hash()) {
		im.transferKeeper.SetDenomTrace(ctx, denomTrace)
	}

	rollapp, err := im.rollappKeeper.ExtractRollappFromChannel(ctx, packet.DestinationChannel, packet.DestinationChannel)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if rollapp == nil {
		return channeltypes.NewErrorAcknowledgement(types.ErrRollappNotFound)
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
func (im IBCRecvMiddleware) OnAcknowledgementPacket(
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

	var data transfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return errorsmod.Wrapf(errortypes.ErrJSONUnmarshal, "unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	dm := types.ParsePacketMetadata(data.Memo)
	if dm == nil {
		return im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	rollapp, err := im.rollappKeeper.ExtractRollappFromChannel(ctx, packet.SourcePort, packet.SourceChannel)
	if err != nil {
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "extract rollapp from packet: %s", err.Error())
	}
	if rollapp == nil {
		return types.ErrRollappNotFound
	}

	if !Contains(rollapp.RegisteredDenoms, dm.Base) {
		// add the new token denom base to the list of rollapp's registered denoms
		rollapp.RegisteredDenoms = append(rollapp.RegisteredDenoms, dm.Base)

		im.rollappKeeper.SetRollapp(ctx, *rollapp)
	}

	return im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

type IBCSendMiddleware struct {
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
) *IBCSendMiddleware {
	return &IBCSendMiddleware{
		ICS4Wrapper:   ics,
		rollappKeeper: rollappKeeper,
		bankKeeper:    bankKeeper,
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
	packet := new(transfertypes.FungibleTokenPacketData)
	if err = types.ModuleCdc.UnmarshalJSON(data, packet); err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrJSONUnmarshal, "unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	if types.MemoAlreadyHasPacketMetadata(packet.Memo) {
		return 0, types.ErrMemoDenomMetadataAlreadyExists
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
