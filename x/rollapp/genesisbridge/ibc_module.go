package genesisbridge

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

const (
	// HubRecipient is the address of `x/hub` module's account on the hub chain.
	HubRecipient = "dym1mk7pw34ypusacm29m92zshgxee3yreums8avur"
)

// IBCModule GenesisBridge is responsible for handling the genesis bridge protocol.
// (ADR: https://www.notion.so/dymension/ADR-x-Genesis-Bridge-109a4a51f86a80ba8b50db454bee04a7?pvs=4)
//
// It validated the genesis info registered on the hub, is the same as the hub's genesis info.
// It registers the denom metadata for the native denom.
// It handles the genesis transfer.
//
// Before the genesis bridge protocol completes, no transfers are allowed to the hub.
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

// OnRecvPacket will handle the genesis bridge packet in case needed.
// no-op for non-hub chains and rollapps with transfers enabled.
//
// The genesis bridge packet is a special packet that is sent from the hub to the hub on channel creation.
// The hub will receive this packet and:
// - validated the genesis info registered on the hub, is the same as the hub's genesis info.
// - registers the denom metadata for the native denom.
// - handles the genesis transfer.
// On success, it will mark the IBC channel for this hub as enabled. This marks the end of the genesis phase.
//
// NOTE: we assume that by this point the canonical channel ID has already been set for the hub, in a secure way.
func (w IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	// Get hub from the packet
	// we don't use the commonly used GetValidTransfer because we have custom type for genesis bridge
	ra, err := w.rollappKeeper.GetRollappByPortChan(ctx, packet.GetDestPort(), packet.GetDestChannel())
	if errorsmod.IsOf(err, types.ErrRollappNotFound) {
		// no problem, it corresponds to a regular non-hub chain
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}
	if err != nil {
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "get hub id"))
	}

	// skip the genesis bridge if the hub already has transfers enabled
	if ra.IsTransferEnabled() {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	l := w.logger(ctx, packet).With("rollapp_id", ra.RollappId)

	// parse the genesis bridge data
	var genesisBridgeData GenesisBridgeData
	if err := json.Unmarshal(packet.GetData(), &genesisBridgeData); err != nil {
		l.Error("Unmarshal genesis bridge data.", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "unmarshal genesis bridge data"))
	}

	v := validator{
		rollapp:   genesisBridgeData,
		hub:       ra.GenesisInfo,
		channelID: ra.ChannelId,
		rollappID: ra.RollappId,
	}

	actionItems, err := v.validateAndGetActionItems()
	if err != nil {
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "validate and get actionable data"))
	}

	w.transferKeeper.SetDenomTrace(ctx, actionItems.trace)
	if err := w.denomKeeper.CreateDenomMetadata(ctx, actionItems.bankMeta); err != nil {
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "create denom metadata"))
	}

	if err := w.transferKeeper.OnRecvPacket(ctx, packet, actionItems.fungiDatas[0]); err != nil {
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "handle genesis transfer"))
	}
	for _, data := range actionItems.fungiDatas {
		if err := w.transferKeeper.OnRecvPacket(ctx, packet, data); err != nil {
			return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "handle genesis transfer"))
		}
	}

	err = w.EnableTransfers(ctx, ra, actionItems.bankMeta.Base)
	if err != nil {
		l.Error("Enable transfers.", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "transfer genesis: enable transfers"))
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeTransfersEnabled,
		sdk.NewAttribute(types.AttributeKeyRollappId, ra.RollappId),
		sdk.NewAttribute(types.AttributeRollappIBCdenom, actionItems.bankMeta.Base),
	))

	// return success ack
	// acknowledgement will be written synchronously during IBC handler execution.
	successAck := channeltypes.NewResultAcknowledgement([]byte{byte(1)})
	return successAck
}

// EnableTransfers marks the end of the genesis bridge phase.
// It sets the transfers enabled flag on the hub.
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
