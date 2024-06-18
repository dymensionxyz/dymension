package denommetadata

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

var _ porttypes.IBCModule = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks for the transfer middleware
type IBCMiddleware struct {
	porttypes.IBCModule
	transferKeeper types.TransferKeeper
	keeper         types.DenomMetadataKeeper
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application
func NewIBCMiddleware(
	app porttypes.IBCModule,
	transferKeeper types.TransferKeeper,
	keeper types.DenomMetadataKeeper,
) IBCMiddleware {
	return IBCMiddleware{
		IBCModule:      app,
		transferKeeper: transferKeeper,
		keeper:         keeper,
	}
}

// OnRecvPacket registers the denom metadata if it does not exist.
// It will intercept an incoming packet and check if the denom metadata exists.
// If it does not, it will register the denom metadata.
// The handler will expect a 'transferinject' object in the memo field of the packet.
// If the memo is not an object, or does not contain the metadata, it moves on to the next handler.
func (im IBCMiddleware) OnRecvPacket(
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

	dm, err := denomMetadataFromMemo(packetData.Memo, ibcDenom)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if dm == nil {
		return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	metadata := *dm

	if err = im.keeper.CreateDenomMetadata(ctx, metadata); err != nil {
		if errorsmod.IsOf(err, types.ErrDenomAlreadyExists) {
			return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
		}
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if !im.transferKeeper.HasDenomTrace(ctx, denomTrace.Hash()) {
		im.transferKeeper.SetDenomTrace(ctx, denomTrace)
	}

	return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
}

func denomMetadataFromMemo(memo, ibcDenom string) (*banktypes.Metadata, error) {
	transferInject := types.ParsePacketMetadata(memo)
	if transferInject == nil || transferInject.DenomMetadata == nil {
		return nil, nil
	}
	dm := transferInject.DenomMetadata

	if err := dm.Validate(); err != nil {
		return nil, fmt.Errorf("invalid denom metadata: %w", err)
	}

	dm.Base = ibcDenom
	dm.DenomUnits[0].Denom = dm.Base

	return dm, nil
}
