package transfergenesis

import (
	"encoding/json"
	"errors"
	"fmt"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	uibc "github.com/dymensionxyz/dymension/v3/utils/ibc"

	"github.com/dymensionxyz/dymension/v3/utils/gerr"

	"github.com/tendermint/tendermint/libs/log"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

const (
	memoNamespaceKey = "genesis_transfer"
)

type DenomMetadataKeeper interface {
	CreateDenomMetadata(ctx sdk.Context, metadata banktypes.Metadata) error
	HasDenomMetadata(ctx sdk.Context, base string) bool
}

type TransferKeeper interface {
	SetDenomTrace(ctx sdk.Context, denomTrace transfertypes.DenomTrace)
}

type IBCModule struct {
	porttypes.IBCModule // next one
	delayedackKeeper    delayedackkeeper.Keeper
	rollappKeeper       rollappkeeper.Keeper
	transferKeeper      TransferKeeper
	denomKeeper         DenomMetadataKeeper
}

func NewIBCModule(
	next porttypes.IBCModule,
	delayedAckKeeper delayedackkeeper.Keeper,
	rollappKeeper rollappkeeper.Keeper,
	transferKeeper TransferKeeper,
	denomKeeper DenomMetadataKeeper,
) IBCModule {
	return IBCModule{
		IBCModule:        next,
		delayedackKeeper: delayedAckKeeper,
		rollappKeeper:    rollappKeeper,
		transferKeeper:   transferKeeper,
		denomKeeper:      denomKeeper,
	}
}

func (w IBCModule) logger(
	ctx sdk.Context,
	packet channeltypes.Packet,
) log.Logger {
	return ctx.Logger().With(
		"module", "transferGenesis",
		"packet_source_port", packet.SourcePort,
		"packet_destination_port", packet.DestinationPort,
		"packet_sequence", packet.Sequence,
		"method", "OnRecvPacket",
	)
}

// OnRecvPacket will, if the packet is a transfer packet:
// if it's not a genesis transfer: enable transfers and pass on the packet. This marks the end of the genesis phase.
// else:
//
//		 transfers must not have already been enabled.
//	     check it's a valid genesis transfer. If it is, then register the denom
//
// NOTE: we assume that by this point the canonical channel ID has already been set
// for the rollapp, in a secure way.
func (w IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	l := w.logger(ctx, packet)

	if commontypes.SkipRollappMiddleware(ctx) || !w.delayedackKeeper.IsRollappsEnabled(ctx) {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	transfer, err := w.rollappKeeper.GetValidTransfer(ctx, packet.GetData(), packet.GetDestPort(), packet.GetDestChannel())
	if err != nil {
		l.Error("Get valid transfer from received packet", "err", err)
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "transfer genesis: get valid transfer"))
	}

	if !transfer.IsRollapp() {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	ra := transfer.Rollapp

	memo, err := getMemo(transfer.GetMemo())
	if errorsmod.IsOf(err, gerr.ErrNotFound) {
		// The first regular transfer marks the full opening of the bridge, more genesis transfers will not be allowed.
		w.rollappKeeper.EnableTransfers(ctx, ra.RollappId)
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}
	if err != nil {
		l.Error("Get memo.", "err", err)
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "get memo"))
	}

	if ra.GenesisState.TransfersEnabled {
		// Genesis transfers are disabled once the bridge is already open
		err = w.handleFraud(ctx, ra.RollappId)
		if err != nil {
			l.Error("Handling fraud.", "err", err)
		} else {
			l.Info("Handled fraud: verify and record genesis transfer.", "err", err)
		}
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(gerrc.ErrFault, "genesis transfers are disabled")) // TODO: gerr
	}
	// it's a valid genesis transfer!

	err = w.registerDenomMetadata(ctx, ra.RollappId, ra.ChannelId, memo.Denom)
	if err != nil {
		l.Error("Register denom metadata.", "err", err)
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "transfer genesis: register denom metadata"))
	}

	l.Debug("Received valid genesis transfer. Registered denom data.")

	// we want to skip delayedack etc because we want the transfer to happen immediately
	return w.IBCModule.OnRecvPacket(commontypes.SkipRollappMiddlewareContext(ctx), packet, relayer)
}

func (w IBCModule) handleFraud(ctx sdk.Context, rollappID string) error {
	// handleFraud : the rollapp has violated the DRS!
	// TODO: finish implementing this method,  see https://github.com/dymensionxyz/dymension/issues/930
	return w.rollappKeeper.HandleFraud(ctx, rollappID, "", 0, "")
}

func getMemo(rawMemo string) (rollapptypes.GenesisTransferMemo, error) {
	if len(rawMemo) == 0 {
		return rollapptypes.GenesisTransferMemo{}, gerr.ErrNotFound
	}

	// check if the key is there, because we want to differentiate between people not sending us the data, vs
	// them sending it but it being malformed

	keyMap := make(map[string]any)

	err := json.Unmarshal([]byte(rawMemo), &keyMap)
	if err != nil {
		return rollapptypes.GenesisTransferMemo{}, errorsmod.Wrap(errors.Join(gerr.ErrInvalidArgument, sdkerrors.ErrJSONUnmarshal), "rawMemo")
	}

	if _, ok := keyMap[memoNamespaceKey]; !ok {
		return rollapptypes.GenesisTransferMemo{}, gerr.ErrNotFound
	}

	var m rollapptypes.GenesisTransferMemoNamespaced
	err = json.Unmarshal([]byte(rawMemo), &m)
	if err != nil {
		return rollapptypes.GenesisTransferMemo{}, errorsmod.Wrap(errors.Join(gerr.ErrInvalidArgument, sdkerrors.ErrJSONUnmarshal), "rawMemo")
	}

	if err := m.Data.Valid(); err != nil {
		return rollapptypes.GenesisTransferMemo{}, errorsmod.Wrap(errors.Join(gerr.ErrInvalidArgument, err), "validate data")
	}
	return m.Data, nil
}

func (w IBCModule) registerDenomMetadata(ctx sdk.Context, rollappID, channelID string, m banktypes.Metadata) error {
	trace := uibc.GetForeignDenomTrace(channelID, m.Base)
	m.Base = trace.IBCDenom()

	if w.denomKeeper.HasDenomMetadata(ctx, m.GetBase()) {
		// Not strictly necessary but an easy optimisation, as, in general, we dont place restrictions on the number
		// of genesis transfers that a rollapp might do.
		return nil
	}

	w.transferKeeper.SetDenomTrace(ctx, trace)

	/*
		Change the base to the ibc denom, and add an alias to the original
	*/
	m.Description = fmt.Sprintf("auto-generated ibc denom for rollapp: base: %s: rollapp: %s", m.GetBase(), rollappID)
	for i, u := range m.DenomUnits {
		if u.Exponent == 0 {
			m.DenomUnits[i].Aliases = append(m.DenomUnits[i].Aliases, u.Denom)
			m.DenomUnits[i].Denom = m.GetBase()
		}
	}

	if err := m.Validate(); err != nil {
		return errorsmod.Wrap(errors.Join(gerr.ErrInvalidArgument, err), "metadata validate")
	}

	// We go by the denom keeper instead of calling bank directly, as something might happen in-between
	err := w.denomKeeper.CreateDenomMetadata(ctx, m)
	if err != nil {
		return errorsmod.Wrap(err, "create denom metadata")
	}

	return nil
}
