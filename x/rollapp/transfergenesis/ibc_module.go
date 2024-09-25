package transfergenesis

import (
	"encoding/json"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	"github.com/dymensionxyz/sdk-utils/utils/uibc"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

const (
	memoNamespaceKey = "genesis_transfer"
)

type IBCModule struct {
	porttypes.IBCModule // next one
	rollappKeeper       RollappKeeper
	transferKeeper      TransferKeeper
	denomKeeper         DenomMetadataKeeper
	iroKeeper           IROKeeper
}

func NewIBCModule(
	next porttypes.IBCModule,
	rollappKeeper RollappKeeper,
	transferKeeper TransferKeeper,
	denomKeeper DenomMetadataKeeper,
	iroKeeper IROKeeper,
) IBCModule {
	return IBCModule{
		IBCModule:      next,
		rollappKeeper:  rollappKeeper,
		transferKeeper: transferKeeper,
		denomKeeper:    denomKeeper,
		iroKeeper:      iroKeeper,
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
) (ack exported.Acknowledgement) {
	l := w.logger(ctx, packet)

	if commontypes.SkipRollappMiddleware(ctx) {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	transfer, err := w.rollappKeeper.GetValidTransfer(ctx, packet.GetData(), packet.GetDestPort(), packet.GetDestChannel())
	if err != nil {
		l.Error("Get valid transfer from received packet", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "transfer genesis: get valid transfer"))
	}

	if !transfer.IsRollapp() {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	ra := transfer.Rollapp
	l = l.With("rollapp_id", ra.RollappId)

	// rollapp transfers already enabled. skip genesis transfer middleware
	if ra.IsTransferEnabled() {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	/* ------------------------ genesis transfer required ----------------------- */
	// extract genesis transfer memo if exists
	memo, err := getMemo(transfer.GetMemo())
	if err != nil {
		l.Error("extract genesis transfer memo.", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "extract genesis transfer memo"))
	}

	// handle genesis transfer by the IRO keeper
	plan := w.iroKeeper.MustGetPlanByRollapp(ctx, ra.RollappId)

	// validate the transfer against the IRO plan
	err = w.validateGenesisTransfer(plan, transfer, memo.Denom)
	if err != nil {
		l.Error("Validate IRO plan.", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "validate IRO plan"))
	}

	err = w.registerDenomMetadata(ctx, ra.RollappId, ra.ChannelId, memo.Denom)
	if err != nil {
		l.Error("Register denom metadata.", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "transfer genesis: register denom metadata"))
	}

	// set the ctx to skip delayedack etc because we want the transfer to happen immediately
	ack = w.IBCModule.OnRecvPacket(commontypes.SkipRollappMiddlewareContext(ctx), packet, relayer)
	// if the ack is nil, we return an error as we expect immediate ack
	if ack == nil {
		l.Error("Expected immediate ack for genesis transfer.")
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(gerrc.ErrInternal, "transfer genesis: OnRecvPacket"))
	}
	err = w.EnableTransfers(ctx, ra.RollappId, transfer.Denom)
	if err != nil {
		l.Error("Enable transfers.", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "transfer genesis: enable transfers"))
	}

	return ack
}

// validate genesis transfer amount is the same as in the `iro` plan
// validate the destAddr is the same as `x/iro` module account address
func (w IBCModule) validateGenesisTransfer(plan irotypes.Plan, transfer rollapptypes.TransferData, genesisTransferDenomMetadata banktypes.Metadata) error {
	if !plan.TotalAllocation.Amount.Equal(transfer.MustAmountInt()) {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer amount does not match plan amount")
	}

	modAddr := w.iroKeeper.GetModuleAccountAddress()
	if modAddr != transfer.Receiver {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer receiver does not match module account address")
	}

	// validate the memo denom against the transfer denom
	if genesisTransferDenomMetadata.Base != transfer.FungibleTokenPacketData.Denom {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "rollapp denom does not match transfer denom")
	}

	if genesisTransferDenomMetadata.Base != transfer.Rollapp.GenesisInfo.NativeDenom.Base {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "rollapp denom does not match genesis info denom")
	}

	correct := false
	for _, unit := range genesisTransferDenomMetadata.DenomUnits {
		if transfer.Rollapp.GenesisInfo.NativeDenom.Exponent == unit.Exponent {
			// TODO: validate the symbol name as well?
			correct = true
			break
		}
	}
	if !correct {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "rollapp denom missing correct exponent")
	}
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

func (w IBCModule) EnableTransfers(ctx sdk.Context, rollappID, rollappBaseDenom string) error {
	ra := w.rollappKeeper.MustGetRollapp(ctx, rollappID)
	ra.GenesisState.TransfersEnabled = true
	w.rollappKeeper.SetRollapp(ctx, ra)

	rollappDenomOnHub := uibc.GetForeignDenomTrace(ra.ChannelId, rollappBaseDenom).IBCDenom()
	err := w.rollappKeeper.GetHooks().AfterTransfersEnabled(ctx, ra.RollappId, rollappDenomOnHub)
	if err != nil {
		return errorsmod.Wrap(err, "after transfers enabled hook")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeTransferGenesisTransfersEnabled,
		sdk.NewAttribute(types.AttributeKeyRollappId, rollappID),
	))

	return nil
}
