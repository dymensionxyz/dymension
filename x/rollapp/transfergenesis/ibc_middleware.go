package transfergenesis

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	uibc "github.com/dymensionxyz/dymension/v3/utils/ibc"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	"github.com/dymensionxyz/dymension/v3/utils/gerr"

	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"

	"github.com/tendermint/tendermint/libs/log"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	dymerror "github.com/dymensionxyz/dymension/v3/x/common/errors"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

var _ porttypes.Middleware = &IBCMiddleware{}

type DenomMetadataKeeper interface {
	CreateDenomMetadata(ctx sdk.Context, metadata banktypes.Metadata) error
	HasDenomMetadata(ctx sdk.Context, base string) bool
}

type TransferKeeper interface {
	SetDenomTrace(ctx sdk.Context, denomTrace transfertypes.DenomTrace)
}

type ChannelKeeper interface {
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error)
}

type IBCMiddleware struct {
	porttypes.Middleware // next one
	delayedackKeeper     delayedackkeeper.Keeper
	rollappKeeper        rollappkeeper.Keeper
	transferKeeper       TransferKeeper
	denomKeeper          DenomMetadataKeeper
	channelKeeper        ChannelKeeper
}

func NewIBCMiddleware(
	next porttypes.Middleware,
	delayedAckKeeper delayedackkeeper.Keeper,
	rollappKeeper rollappkeeper.Keeper,
	transferKeeper TransferKeeper,
	denomKeeper DenomMetadataKeeper,
	channelKeeper ChannelKeeper,
) IBCMiddleware {
	return IBCMiddleware{
		Middleware:       next,
		delayedackKeeper: delayedAckKeeper,
		rollappKeeper:    rollappKeeper,
		transferKeeper:   transferKeeper,
		denomKeeper:      denomKeeper,
		channelKeeper:    channelKeeper,
	}
}

func (w IBCMiddleware) logger(
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

type memo struct {
	Denom banktypes.Metadata `json:"denom"`
	// How many transfers in total will be sent in the transfer genesis period
	TotalNumTransfers uint64 `json:"total_num_transfers"`
	// Which transfer is this? If there are 5 transfers total, they will be numbered 0,1,2,3,4.
	ThisTransferIx uint64 `json:"this_transfer_ix"`
}

func hackSetCanonicalChannel(
	ctx sdk.Context,
	packet channeltypes.Packet,
	w IBCMiddleware,
) {
	// TODO: prior to this we relied on the whitelist addr to set the canonical channel
	//	 This is a hack (not secure)
	//	 The real solution will come in a followup PR
	//	 See https://github.com/dymensionxyz/research/issues/242

	l := ctx.Logger().With("hack set canonical channel")
	t, err := w.delayedackKeeper.GetValidTransferFromReceivedPacket(ctx, packet)
	if err != nil {
		l.Error("get valid transfer", "error", err)
	}

	if !t.IsFromRollapp() {
		return
	}

	// if valid t returns a rollapp, we know we must get it
	ra := w.rollappKeeper.MustGetRollapp(ctx, t.RollappID)
	ra.ChannelId = packet.GetDestChannel()
	w.rollappKeeper.SetRollapp(ctx, ra)
}

// OnRecvPacket will, if the packet is a transfer packet:
// if it's not a genesis transfer: pass on the packet only if transfers are enabled
// else: check it's a valid genesis transfer. If it is, then register the denom, if
// it's the last one, open the bridge.
// NOTE: we assume that by this point the canonical channel ID has already been set
// for the rollapp, in a secure way.
func (w IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	hackSetCanonicalChannel(ctx, packet, w) // TODO: remove!

	l := w.logger(ctx, packet)

	if !w.delayedackKeeper.IsRollappsEnabled(ctx) {
		return w.Middleware.OnRecvPacket(ctx, packet, relayer)
	}

	transfer, err := w.delayedackKeeper.GetValidTransferFromReceivedPacket(ctx, packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "get valid transfer"))
	}

	if !transfer.IsFromRollapp() {
		return w.Middleware.OnRecvPacket(ctx, packet, relayer)
	}

	// if valid transfer returns a rollapp, we know we must get it
	ra := w.rollappKeeper.MustGetRollapp(ctx, transfer.RollappID)

	m, err := getMemo(transfer.GetMemo())
	if errorsmod.IsOf(err, gerr.ErrNotFound) {
		// This is a normal transfer
		if !ra.GenesisState.TransfersEnabled {
			err = errorsmod.Wrapf(gerr.ErrFailedPrecondition, "transfers are disabled: rollapp id: %s", ra.RollappId)
			// Someone on the RA tried to send a transfer before the bridge is open! Return an err ack and they will get refunded
			return channeltypes.NewErrorAcknowledgement(err)
		}
		return w.Middleware.OnRecvPacket(ctx, packet, relayer)
	}
	if err != nil {
		l.Debug("get memo", "error", err)
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "get memo"))
	}

	nTransfersDone, err := w.rollappKeeper.VerifyAndRecordGenesisTransfer(ctx, ra.RollappId, m.ThisTransferIx, m.TotalNumTransfers)
	if errorsmod.IsOf(err, dymerror.ErrFraud) {
		l.Info("rollapp fraud: verify and record genesis transfer", "err", err)
		// The rollapp has deviated from the protocol!
		err = w.handleFraud(ra.RollappId)
		if err != nil {
			l.Error("handle fraud", "error", err)
		}
	}
	if err != nil {
		l.Error("verify and record transfer", "error", err)
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "verify and record genesis transfer"))
	}

	// it's a valid genesis transfer!

	err = w.registerDenomMetadata(ctx, ra.RollappId, ra.ChannelId, m.Denom)
	if err != nil {
		l.Error("register denom metadata", "error", err)
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "register denom metadata"))
	}

	l.Debug("Received valid genesis transfer. Registered denom data.",
		"num total", m.TotalNumTransfers,
		"num so far", nTransfersDone,
		"ix", m.ThisTransferIx,
	)

	if nTransfersDone == m.TotalNumTransfers {
		// The transfer window is finished! Queue up a finalization
		w.rollappKeeper.EnableTransfers(ctx, ra.RollappId)
		ctx.EventManager().EmitEvent(allTransfersReceivedEvent(ra.RollappId, nTransfersDone))
		l.Info("All genesis transfers received, bridge opened.",
			"rollapp", ra.RollappId,
			"n transfers", nTransfersDone)
	}

	return w.Middleware.OnRecvPacket(delayedacktypes.SkipContext(ctx), packet, relayer)
}

func (w IBCMiddleware) handleFraud(raID string) error {
	// TODO: assumes we have a robust mechanism to freeze and rollback
	// 		 NOTE: this may not be a fraud in the sense that the sequencer posted the wrong state root
	//             it might be that the canonical state transition function for the RA is wrong somehow
	//             and resulted in a protocol breakage
	// 		https://github.com/dymensionxyz/dymension/issues/921
	return nil
}

func allTransfersReceivedEvent(raID string, nReceived uint64) sdk.Event {
	return sdk.NewEvent(types.EventTypeTransferGenesisAllReceived,
		sdk.NewAttribute(types.AttributeKeyRollappId, raID),
		sdk.NewAttribute(types.AttributeKeyTransferGenesisNReceived, strconv.FormatUint(nReceived, 10)),
	)
}

func getMemo(rawMemo string) (memo, error) {
	if len(rawMemo) == 0 {
		return memo{}, gerr.ErrNotFound
	}

	key := "genesis_transfer"

	// check if the key is there, because we want to differentiate between people not sending us the data, vs
	// them sending it but it being malformed

	keyMap := make(map[string]any)

	err := json.Unmarshal([]byte(rawMemo), &keyMap)
	if err != nil {
		return memo{}, errorsmod.Wrap(errors.Join(gerr.ErrInvalidArgument, sdkerrors.ErrJSONUnmarshal), "rawMemo")
	}

	if _, ok := keyMap[key]; !ok {
		return memo{}, gerr.ErrNotFound
	}

	type t struct {
		Data memo `json:"genesis_transfer"`
	}

	var m t
	err = json.Unmarshal([]byte(rawMemo), &m)
	if err != nil {
		return memo{}, errorsmod.Wrap(errors.Join(gerr.ErrInvalidArgument, sdkerrors.ErrJSONUnmarshal), "rawMemo")
	}
	return m.Data, nil
}

func (w IBCMiddleware) registerDenomMetadata(ctx sdk.Context, rollappID, channelID string, m banktypes.Metadata) error {
	if w.denomKeeper.HasDenomMetadata(ctx, m.Base) {
		// Not strictly necessary but an easy optimisation, as, in general, we dont place restrictions on the number
		// of genesis transfers that a rollapp might do.
		return nil
	}

	trace := uibc.GetForeignDenomTrace(channelID, m.Base)

	w.transferKeeper.SetDenomTrace(ctx, trace)

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
		return errorsmod.Wrap(errors.Join(gerr.ErrInvalidArgument, err), "metadata validate")
	}

	// We go by the denom keeper instead of calling bank directly, as something might happen in-between
	err := w.denomKeeper.CreateDenomMetadata(ctx, m)
	if err != nil {
		return errorsmod.Wrap(err, "create denom metadata")
	}

	return nil
}
