package transfergenesis

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dymensionxyz/sdk-utils/utils/uibc"

	"github.com/cometbft/cometbft/libs/log"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var ErrDisabled = errorsmod.Wrap(gerrc.ErrFault, "genesis transfers are disabled")

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

type IROKeeper interface {
	GetPlanByRollapp(ctx sdk.Context, rollappID string) (irotypes.Plan, bool)
	GetModuleAccountAddress() string
}

type IBCModule struct {
	porttypes.IBCModule // next one
	delayedackKeeper    delayedackkeeper.Keeper
	rollappKeeper       rollappkeeper.Keeper
	transferKeeper      TransferKeeper
	denomKeeper         DenomMetadataKeeper
	iroKeeper           IROKeeper
}

func NewIBCModule(
	next porttypes.IBCModule,
	delayedAckKeeper delayedackkeeper.Keeper,
	rollappKeeper rollappkeeper.Keeper,
	transferKeeper TransferKeeper,
	denomKeeper DenomMetadataKeeper,
	iroKeeper IROKeeper,
) IBCModule {
	return IBCModule{
		IBCModule:        next,
		delayedackKeeper: delayedAckKeeper,
		rollappKeeper:    rollappKeeper,
		transferKeeper:   transferKeeper,
		denomKeeper:      denomKeeper,
		iroKeeper:        iroKeeper,
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

// OnRecvPacket will,
// if iro plan exists for this rollapp:
//	    check it's a valid genesis transfer.
// 		If it is, then pass the packet, register the denom and settle the plan.
// In any case, mark the transfers enabled.
// This marks the end of the genesis phase.

// NOTE: we assume that by this point the canonical channel ID has already been set
// for the rollapp, in a secure way.
func (w IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	l := w.logger(ctx, packet)

	if commontypes.SkipRollappMiddleware(ctx) {
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
	l = l.With("rollapp_id", ra.RollappId)

	noMemo := false
	memo, err := getMemo(transfer.GetMemo())
	if errorsmod.IsOf(err, gerrc.ErrNotFound) {
		noMemo = true
	}
	if err != nil && !noMemo {
		l.Error("Get memo.", "err", err)
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "get memo"))
	}

	// if already enabled, skip this middleware
	// genesis transfer memo NOT allowed
	if ra.GenesisState.TransfersEnabled {
		if !noMemo {
			l.Error("Genesis transfers already enabled.")
			_ = w.handleDRSViolation(ctx, ra.RollappId)
			return channeltypes.NewErrorAcknowledgement(ErrDisabled)
		}
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	// check if iro plan exists. if it does, a genesis transfer is required
	plan, found := w.iroKeeper.GetPlanByRollapp(ctx, ra.RollappId)
	if found {
		// plan exists, genesis transfer required
		if noMemo {
			l.Error("genesis transfer required for rollapp with IRO plan.")
			return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(gerrc.ErrFailedPrecondition, "no memo found for rollapp with plan"))
		}

		// validate the transfer against the IRO plan
		err = w.validateGenesisTransfer(plan, transfer, l)
		if err != nil {
			l.Error("Validate IRO plan.", "err", err)
			return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "validate IRO plan"))
		}

		// register the denom metadata
		err = w.registerDenomMetadata(ctx, ra.RollappId, ra.ChannelId, memo.Denom)
		if err != nil {
			l.Error("Register denom metadata.", "err", err)
			return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "transfer genesis: register denom metadata"))
		}

		// set the ctx to skip delayedack etc because we want the transfer to happen immediately
		ctx = commontypes.SkipRollappMiddlewareContext(ctx)
	} else {
		// no plan found, genesis transfer memo not allowed
		if !noMemo {
			l.Error("No plan found for rollapp. Genesis transfer memo not allowed.")
			return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer not allowed"))
		}
	}

	transferAck := w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	if !transferAck.Success() {
		return transferAck
	}

	w.EnableTransfers(ctx, ra.RollappId)
	err = w.rollappKeeper.GetHooks().TransfersEnabled(ctx, ra.RollappId)
	if err != nil {
		l.Error("Transfers enabled hook.", "err", err)
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "transfer genesis: transfers enabled hook"))
	}

	l.Info("Received valid genesis transfer. Registered denom data.")

	return transferAck
}

// validate genesis transfer amount is the same as in the `iro` plan
// validate the destAddr is the same as `x/iro` module account address
func (w IBCModule) validateGenesisTransfer(plan irotypes.Plan, transfer rollapptypes.TransferData, l log.Logger) error {
	if plan.TotalAllocation.Amount != transfer.MustAmountInt() {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer amount does not match plan amount")
	}

	modAddr := w.iroKeeper.GetModuleAccountAddress()
	if modAddr != transfer.Receiver {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer receiver does not match module account address")
	}

	return nil
}

func (w IBCModule) handleDRSViolation(ctx sdk.Context, rollappID string) error {
	// handleFraud : the rollapp has violated the DRS!
	// TODO: finish implementing this method,  see https://github.com/dymensionxyz/dymension/issues/930
	return nil
}

func getMemo(rawMemo string) (rollapptypes.GenesisTransferMemo, error) {
	if len(rawMemo) == 0 {
		return rollapptypes.GenesisTransferMemo{}, gerrc.ErrNotFound
	}

	// check if the key is there, because we want to differentiate between people not sending us the data, vs
	// them sending it but it being malformed

	keyMap := make(map[string]any)

	err := json.Unmarshal([]byte(rawMemo), &keyMap)
	if err != nil {
		return rollapptypes.GenesisTransferMemo{}, errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, sdkerrors.ErrJSONUnmarshal), "rawMemo")
	}

	if _, ok := keyMap[memoNamespaceKey]; !ok {
		return rollapptypes.GenesisTransferMemo{}, gerrc.ErrNotFound
	}

	var m rollapptypes.GenesisTransferMemoNamespaced
	err = json.Unmarshal([]byte(rawMemo), &m)
	if err != nil {
		return rollapptypes.GenesisTransferMemo{}, errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, sdkerrors.ErrJSONUnmarshal), "rawMemo")
	}

	if err := m.Data.Valid(); err != nil {
		return rollapptypes.GenesisTransferMemo{}, errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "validate data")
	}
	return m.Data, nil
}

func (w IBCModule) registerDenomMetadata(ctx sdk.Context, rollappID, channelID string, m banktypes.Metadata) error {
	trace := uibc.GetForeignDenomTrace(channelID, m.Base)
	m.Base = trace.IBCDenom()

	if w.denomKeeper.HasDenomMetadata(ctx, m.GetBase()) {
		return fmt.Errorf("denom metadata already exists for base: %s", m.GetBase())
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
		return errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "metadata validate")
	}

	// We go by the denom keeper instead of calling bank directly, as something might happen in-between
	err := w.denomKeeper.CreateDenomMetadata(ctx, m)
	if err != nil {
		return errorsmod.Wrap(err, "create denom metadata")
	}

	return nil
}

func (w IBCModule) EnableTransfers(ctx sdk.Context, rollappID string) {
	ra := w.rollappKeeper.MustGetRollapp(ctx, rollappID)
	ra.GenesisState.TransfersEnabled = true
	w.rollappKeeper.SetRollapp(ctx, ra)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeTransferGenesisTransfersEnabled,
		sdk.NewAttribute(types.AttributeKeyRollappId, rollappID),
	))
}
