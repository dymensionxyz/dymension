package transfergenesis

import (
	"encoding/json"
	"errors"
	"fmt"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

var _ porttypes.Middleware = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks
type IBCMiddleware struct {
	porttypes.Middleware
	keeper   delayedackkeeper.Keeper
	raKeeper rollappkeeper.Keeper
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application
func NewIBCMiddleware(next porttypes.Middleware, keeper delayedackkeeper.Keeper, raK rollappkeeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		Middleware: next,
		keeper:     keeper,
		raKeeper:   raK,
	}
}

func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	// may modify packet
	err := im.handleGenesisTransfers(ctx, &packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	return im.OnRecvPacket(ctx, packet, relayer)
}

type genesisTransferDenomMemo struct {
	GenesisTransfer struct {
		Data banktypes.Metadata `json:"data"`
	} `json:"genesis_transfer"`
}

func (im IBCMiddleware) handleGenesisTransfers(
	ctx sdk.Context,
	packet *channeltypes.Packet,
) error {
	if !im.keeper.IsRollappsEnabled(ctx) {
		return nil
	}

	l := ctx.Logger().With(
		"middleware", "transferGenesis",
		"packet_source_port", packet.SourcePort,
		"packet_destination_port", packet.DestinationPort,
		"packet_sequence", packet.Sequence)

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrJSONUnmarshal, "fungible token packet")
	}

	memo := data.GetMemo()
	var wrappedDenom genesisTransferDenomMemo // wrapped for memo namespacing reasons
	err := json.Unmarshal([]byte(memo), &wrappedDenom)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrJSONUnmarshal, "memo")
	}

	denom := wrappedDenom.GenesisTransfer.Data

	l.Info("got the special memo!")

	chaID := "channel-0"
	raID := "rollappevm_1234-1"

	ra, ok := im.raKeeper.GetRollapp(ctx, raID)
	if !ok {
		panic(errors.New("must find rollapp"))
	}

	_ = ra

	err = im.raKeeper.MarkGenesisAsHappened(ctx, chaID, raID)
	if err != nil {
		err = fmt.Errorf("mark genesis: %w", err)
		l.Error("OnRecvPacket", "err", err)
		panic(err)
	}

	err = im.raKeeper.RegisterOneDenomMetadata(ctx, denom, raID, chaID)
	if err != nil {
		err = fmt.Errorf("register denom meta: %w", err)
		l.Error("OnRecvPacket", "err", err)
		panic(err)
	}

	l.Info("Registered denom meta data from genesis transfer.")

	return nil
}
