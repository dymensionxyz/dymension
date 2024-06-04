package transfergenesis

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dymensionxyz/dymension/v3/utils"

	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"

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

type DenomMetadataKeeper interface {
	CreateDenomMetadata(ctx sdk.Context, metadata banktypes.Metadata) error
}

type TransferKeeper interface {
	SetDenomTrace(ctx sdk.Context, denomTrace transfertypes.DenomTrace)
}

type IBCMiddleware struct {
	porttypes.Middleware // next one
	delayedackKeeper     delayedackkeeper.Keeper
	rollappKeeper        rollappkeeper.Keeper
	transferKeeper       TransferKeeper
	denomKeeper          DenomMetadataKeeper
}

func NewIBCMiddleware(
	next porttypes.Middleware,
	keeper delayedackkeeper.Keeper,
	raK rollappkeeper.Keeper,
	transferKeeper TransferKeeper,
	denomKeeper DenomMetadataKeeper,
) IBCMiddleware {
	return IBCMiddleware{
		Middleware:       next,
		delayedackKeeper: keeper,
		rollappKeeper:    raK,
		transferKeeper:   transferKeeper,
		denomKeeper:      denomKeeper,
	}
}

func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	ctx, err := im.handleGenesisTransfers(ctx, packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	return im.OnRecvPacket(ctx, packet, relayer)
}

type genesisTransferDenomMemo struct {
	Data struct {
		Denom banktypes.Metadata `json:"denom"`
	} `json:"genesis_transfer"`
}

func (im IBCMiddleware) handleGenesisTransfers(
	ctx sdk.Context,
	packet channeltypes.Packet,
) (sdk.Context, error) {
	if !im.delayedackKeeper.IsRollappsEnabled(ctx) {
		// TODO: makes sense?
		return ctx, nil
	}

	l := ctx.Logger().With(
		"middleware", "transferGenesis",
		"packet_source_port", packet.SourcePort,
		"packet_destination_port", packet.DestinationPort,
		"packet_sequence", packet.Sequence)

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrJSONUnmarshal, "fungible token packet")
	}

	memo := data.GetMemo()
	var wrappedDenom genesisTransferDenomMemo // wrapped for memo namespacing reasons
	err := json.Unmarshal([]byte(memo), &wrappedDenom)
	if err != nil {
		return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrJSONUnmarshal, "memo")
	}

	denom := wrappedDenom.Data.Denom

	l.Info("got the special memo!") // TODO: fix

	chaID := "channel-0"
	raID := "rollappevm_1234-1"

	ra, ok := im.rollappKeeper.GetRollapp(ctx, raID)
	if !ok {
		panic(errors.New("must find rollapp"))
	}

	_ = ra

	err = im.rollappKeeper.MarkGenesisAsHappened(ctx, chaID, raID)
	if err != nil {
		err = fmt.Errorf("mark genesis: %w", err)
		l.Error("OnRecvPacket", "err", err)
		panic(err)
	}

	err = im.rollappKeeper.RegisterDenomMetadata(ctx, raID, chaID, denom)
	if err != nil {
		err = fmt.Errorf("register denom meta: %w", err)
		l.Error("OnRecvPacket", "err", err)
		panic(err)
	}

	l.Info("Registered denom meta data from genesis transfer.")

	return delayedacktypes.SkipContext(ctx), nil
}

func (im IBCMiddleware) RegisterDenomMetadata(ctx sdk.Context, rollappID, channelID string, m banktypes.Metadata) error {
	// TODO: only do it if it hasn't been done before?

	trace := utils.GetForeignDenomTrace(channelID, m.Base)

	im.transferKeeper.SetDenomTrace(ctx, trace)

	ibcDenom := trace.IBCDenom()

	/*
		Change the base to the ibc denom, and add an alias to the original
	*/
	m.Description = fmt.Sprintf("auto-generated ibc denom for rollapp: base: %s: rollapp: %s", ibcDenom, rollappID)
	m.Base = ibcDenom
	for i, u := range m.DenomUnits {
		if u.Exponent == 0 {
			m.DenomUnits[i].Aliases = append(m.DenomUnits[i].Aliases, u.Denom)
			m.DenomUnits[i].Denom = ibcDenom
		}
	}

	if err := m.Validate(); err != nil {
		// TODO: errorsmod with nice wrapping
		return fmt.Errorf("invalid denom metadata on genesis event: %w", err)
	}

	/*
		TODO: should not be direct as need to make sure vfc contracts etc are created
	*/
	err := im.denomKeeper.CreateDenomMetadata(ctx, m)
	if err != nil {
		sdkerrors.ErrNotSupported
	}
	k.bankKeeper.SetDenomMetaData(ctx, m)

	k.Logger(ctx).Info("Registered denom metadata for IBC token.", "rollappID", rollappID, "denom", ibcDenom)
	return nil
}
