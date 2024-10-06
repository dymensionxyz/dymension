package genesisbridge

import (
	"encoding/json"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	"github.com/dymensionxyz/sdk-utils/utils/uibc"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// GenesisBridge IBC module is responsible for handling the genesis bridge protocol.
// (ADR: https://www.notion.so/dymension/ADR-x-Genesis-Bridge-109a4a51f86a80ba8b50db454bee04a7?pvs=4)
//
// It validated the genesis info registered on the hub, is the same as the rollapp's genesis info.
// It registers the denom metadata for the native denom.
// It handles the genesis transfer.
//
// Before the genesis bridge protocol completes, no transfers are allowed to the rollapp.
// The Hub will block transfers Hub->RA to enforce this.
//
// Important: it is now WRONG to open an ibc connection in the Rollapp->Hub direction.
// Connections should be opened in the Hub->Rollapp direction only
type IBCModule struct {
	porttypes.IBCModule // next one
	rollappKeeper       RollappKeeper
	transferKeeper      TransferKeeper
	denomKeeper         DenomMetadataKeeper
}

func NewIBCModule(
	next porttypes.IBCModule,
	rollappKeeper RollappKeeper,
	transferKeeper TransferKeeper,
	denomKeeper DenomMetadataKeeper,
) IBCModule {
	return IBCModule{
		IBCModule:      next,
		rollappKeeper:  rollappKeeper,
		transferKeeper: transferKeeper,
		denomKeeper:    denomKeeper,
	}
}

func (w IBCModule) logger(
	ctx sdk.Context,
	packet channeltypes.Packet,
) log.Logger {
	return ctx.Logger().With(
		"module", "genesisbridge",
		"packet_source_port", packet.SourcePort,
		"packet_destination_port", packet.DestinationPort,
		"packet_sequence", packet.Sequence,
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
	// Get rollapp from the packet
	// we don't use the commonly used GetValidTransfer because we have custom type for genesis bridge
	ra, err := w.rollappKeeper.GetRollappByPortChan(ctx, packet.GetDestPort(), packet.GetDestChannel())
	if errorsmod.IsOf(err, types.ErrRollappNotFound) {
		// no problem, it corresponds to a regular non-rollapp chain
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}
	if err != nil {
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "get rollapp id"))
	}

	// skip the genesis bridge if the rollapp already has transfers enabled
	if ra.IsTransferEnabled() {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	l := w.logger(ctx, packet).With("rollapp_id", ra.RollappId)

	// parse the genesis bridge data
	var genesisBridgeData rollapptypes.GenesisBridgeData
	if err := json.Unmarshal(packet.GetData(), &genesisBridgeData); err != nil {
		l.Error("Unmarshal genesis bridge data.", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "unmarshal genesis bridge data"))
	}

	// stateless validation of the genesis bridge data
	if err := genesisBridgeData.ValidateBasic(); err != nil {
		l.Error("Validate basic genesis bridge data.", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "validate basic genesis bridge data"))
	}

	// validate genesis info against the expected data set on the rollapp
	err = w.ValidateGenesisBridge(ctx, ra, genesisBridgeData.GenesisInfo)
	if err != nil {
		l.Error("Validate genesis info.", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "validate genesis info"))
	}

	// register the denom metadata. the supplied denom is validated on validateBasic
	raBaseDenom, err := w.registerDenomMetadata(ctx, ra.RollappId, ra.ChannelId, genesisBridgeData.NativeDenom)
	if err != nil {
		l.Error("Register denom metadata.", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "transfer genesis: register denom metadata"))
	}

	// validate and handle the genesis transfer
	err = w.handleGenesisTransfer(ctx, *ra, packet, genesisBridgeData.GenesisTransfer)
	if err != nil {
		l.Error("Handle genesis transfer.", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "handle genesis transfer"))
	}

	err = w.EnableTransfers(ctx, ra, raBaseDenom)
	if err != nil {
		l.Error("Enable transfers.", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "transfer genesis: enable transfers"))
	}

	// FIXME: add more data (denom, amount, etc). move events to proto
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeTransferGenesisTransfersEnabled,
		sdk.NewAttribute(types.AttributeKeyRollappId, ra.RollappId),
	))

	// return success ack
	// acknowledgement will be written synchronously during IBC handler execution.
	successAck := channeltypes.NewResultAcknowledgement([]byte{byte(1)})
	return successAck
}

// TODO: make it prettier?
func (w IBCModule) ValidateGenesisBridge(ctx sdk.Context, ra *types.Rollapp, data rollapptypes.GenesisBridgeInfo) error {
	raGenesisInfo := types.GenesisBridgeInfo{
		GenesisChecksum: ra.GenesisInfo.GenesisChecksum,
		Bech32Prefix:    ra.GenesisInfo.Bech32Prefix,
		NativeDenom:     &ra.GenesisInfo.NativeDenom,
		InitialSupply:   ra.GenesisInfo.InitialSupply,
	}

	if raGenesisInfo != data {
		return fmt.Errorf("genesis info mismatch: expected: %v, got: %v", raGenesisInfo, data)
	}

	return nil
}

func (w IBCModule) registerDenomMetadata(ctx sdk.Context, rollappID, channelID string, m banktypes.Metadata) (string, error) {
	trace := uibc.GetForeignDenomTrace(channelID, m.Base)
	m.Base = trace.IBCDenom()

	if w.denomKeeper.HasDenomMetadata(ctx, m.GetBase()) {
		return "", fmt.Errorf("denom metadata already exists for base: %s", m.GetBase())
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
		return "", errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "metadata validate")
	}

	// We go by the denom keeper instead of calling bank directly, as something might happen in-between
	err := w.denomKeeper.CreateDenomMetadata(ctx, m)
	if err != nil {
		return "", errorsmod.Wrap(err, "create denom metadata")
	}

	return m.Base, nil
}

// EnableTransfers marks the end of the genesis bridge phase.
// It sets the transfers enabled flag on the rollapp.
// It also calls the after transfers enabled hook.
func (w IBCModule) EnableTransfers(ctx sdk.Context, ra *types.Rollapp, rollappIBCtrace string) error {
	ra.GenesisState.TransfersEnabled = true
	w.rollappKeeper.SetRollapp(ctx, *ra)

	// call the after transfers enabled hook
	// currently, used for IRO settlement
	err := w.rollappKeeper.GetHooks().AfterTransfersEnabled(ctx, ra.RollappId, rollappIBCtrace)
	if err != nil {
		return errorsmod.Wrap(err, "after transfers enabled hook")
	}

	return nil
}
