package transfergenesis

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/transfersenabled"

	"github.com/dymensionxyz/dymension/v3/utils/derr"

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
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
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

type memo struct {
	Denom banktypes.Metadata `json:"denom"`
	// How many transfers in total will be sent in the transfer genesis period
	TotalNumTransfers uint64 `json:"total_num_transfers"`
	// Which transfer is this? If there are 5 transfers total, they will be numbered 0,1,2,3,4.
	ThisTransferIx uint64 `json:"this_transfer_ix"`
}

// OnRecvPacket will, if the packet is a transfer packet:
// if it's not a genesis transfer: pass on the packet only if transfers are enabled
// else: check it's a valid genesis transfer. If it is, then register the denom, if
// it's the last one, open the bridge.
// NOTE: we assume that by this point the canonical channel ID has already been set
// for the rollapp, in a secure way.
func (w IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	l := w.logger(ctx, packet)

	l.Debug("Processing inbound packet.")

	if !w.delayedackKeeper.IsRollappsEnabled(ctx) {
		l.Debug("Rollapps are disabled.")
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	transfer, err := w.rollappKeeper.GetValidTransferFromReceivedPacket(ctx, packet)
	if err != nil {
		l.Error("Get valid transfer from received packet", "err", err)
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "get valid transfer"))
	}

	if !transfer.IsRollapp() {
		l.Debug("Transfer is not from a rollapp.")
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	m, err := getMemo(transfer.GetMemo())
	if errorsmod.IsOf(err, gerr.ErrNotFound) {
		l.Debug("Memo not found.")
		// If someone tries to send a transfer without the memo before the bridge is open, they will
		// be blocked at the transfersenabled middleware
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}
	if err != nil {
		l.Error("Get memo.", "err", err)
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "get memo"))
	}

	// if valid transfer returns a rollapp, we know we must get it
	ra := w.rollappKeeper.MustGetRollapp(ctx, transfer.RollappID)

	nTransfersDone, err := w.rollappKeeper.VerifyAndRecordGenesisTransfer(ctx, ra.RollappId, m.ThisTransferIx, m.TotalNumTransfers)
	if errorsmod.IsOf(err, derr.ErrViolatesDymensionRollappStandard) {
		// The rollapp has deviated from the protocol!
		handleFraudErr := w.handleFraud(ra.RollappId)
		if err != nil {
			l.Error("Handling fraud.", "err", handleFraudErr)
		} else {
			l.Info("Handled fraud: verify and record genesis transfer.", "err", err)
		}
		return channeltypes.NewErrorAcknowledgement(err)
	}
	if err != nil {
		l.Error("Verify and record transfer.", "err", err)
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "verify and record genesis transfer"))
	}

	// it's a valid genesis transfer!

	err = w.registerDenomMetadata(ctx, ra.RollappId, ra.ChannelId, m.Denom)
	if err != nil {
		l.Error("Register denom metadata.", "err", err)
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "register denom metadata"))
	}

	l.Debug("Received valid genesis transfer. Registered denom data.",
		"num total", m.TotalNumTransfers,
		"num received so far", nTransfersDone,
		"this transfer ix", m.ThisTransferIx,
	)

	if nTransfersDone == m.TotalNumTransfers {
		// The transfer window is finished! Queue up a finalization
		w.rollappKeeper.EnableTransfers(ctx, ra.RollappId)
		ctx.EventManager().EmitEvent(allTransfersReceivedEvent(ra.RollappId, nTransfersDone))
		l.Info("All genesis transfers received, bridge opened.",
			"rollapp", ra.RollappId,
			"n transfers", nTransfersDone)
	}

	l.Debug("Passing on the transfer down the stack, but skipping delayedack and the transferEnabled blocker.")

	return w.IBCModule.OnRecvPacket(transfersenabled.SkipContext(delayedacktypes.SkipContext(ctx)), packet, relayer)
}

// handleFraud : the rollapp has violated the DRS!
func (w IBCModule) handleFraud(raID string) error {
	// TODO: see https://github.com/dymensionxyz/dymension/issues/930
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
