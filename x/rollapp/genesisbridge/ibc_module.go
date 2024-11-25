package genesisbridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	"github.com/dymensionxyz/sdk-utils/utils/uibc"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
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

	v := genesisTransferValidator{
		rollapp:   genesisBridgeData,
		hub:       ra.GenesisInfo,
		channelID: ra.ChannelId,
		rollappID: ra.RollappId,
	}

	actionableData, err := v.validateAndGetActionableData()
	if err != nil {
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "validate and get actionable data"))
	}

	w.transferKeeper.SetDenomTrace(ctx, actionableData.trace)
	if err := w.denomKeeper.CreateDenomMetadata(ctx, actionableData.bankMeta); err != nil {
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "create denom metadata"))
	}

	err = w.EnableTransfers(ctx, ra, actionableData.bankMeta.Base)
	if err != nil {
		l.Error("Enable transfers.", "err", err)
		return uevent.NewErrorAcknowledgement(ctx, errorsmod.Wrap(err, "transfer genesis: enable transfers"))
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeTransfersEnabled,
		sdk.NewAttribute(types.AttributeKeyRollappId, ra.RollappId),
		sdk.NewAttribute(types.AttributeRollappIBCdenom, actionableData.bankMeta.Base),
	))

	// return success ack
	// acknowledgement will be written synchronously during IBC handler execution.
	successAck := channeltypes.NewResultAcknowledgement([]byte{byte(1)})
	return successAck
}

func compareGenesisAccounts(raCommitted *types.GenesisAccounts, gbData []types.GenesisAccount) error {
	if raCommitted == nil {
		if len(gbData) == 0 {
			return nil
		}
		return fmt.Errorf("genesis accounts length mismatch: expected 0, got %d", len(gbData))
	}

	if len(raCommitted.Accounts) != len(gbData) {
		return fmt.Errorf("genesis accounts length mismatch: expected %d, got %d", len(raCommitted.Accounts), len(gbData))
	}

	for _, acc := range raCommitted.Accounts {
		found := slices.ContainsFunc(gbData, func(dataAcc types.GenesisAccount) bool {
			return dataAcc.Address == acc.Address && dataAcc.Amount.Equal(acc.Amount)
		})

		if !found {
			return fmt.Errorf("genesis account mismatch: account %s with amount %v not found in data", acc.Address, acc.Amount)
		}
	}

	return nil
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

type genesisTransferValidator struct {
	rollapp   GenesisBridgeData // what the rollapp sent over IBC
	hub       types.GenesisInfo // what the hub thinks is correct
	channelID string            // can use "channel-0" in simulation
	rollappID string            // the actual rollapp ID
}

type genesisTransferActionableData struct {
	trace      transfertypes.DenomTrace
	bankMeta   banktypes.Metadata
	fungiDatas []transfertypes.FungibleTokenPacketData
}

func (e *genesisTransferValidator) validateAndGetActionableData() (*genesisTransferActionableData, error) {
	if err := e.rollapp.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic genesis bridge data")
	}

	if err := e.validateAgainstHub(e.rollapp.GenesisInfo, e.hub); err != nil {
		return nil, errorsmod.Wrap(err, "validate against hub")
	}

	ret := &genesisTransferActionableData{}
	trace, bankMeta, err := e.getBankStuff()
	if err != nil {
		return nil, errorsmod.Wrap(err, "get bank stuff")
	}
	ret.trace = *trace
	ret.bankMeta = *bankMeta

	fungiDatas, err := e.getFungiData()
	if err != nil {
		return nil, errorsmod.Wrap(err, "get fungi data")
	}
	ret.fungiDatas = fungiDatas
	return ret, nil
}

func (e *genesisTransferValidator) validateAgainstHub(packet GenesisBridgeInfo, hub types.GenesisInfo) error {
	if packet.GenesisChecksum != hub.GenesisChecksum {
		return fmt.Errorf("genesis checksum mismatch: expected: %v, got: %v", hub.GenesisChecksum, packet.GenesisChecksum)
	}

	if packet.Bech32Prefix != hub.Bech32Prefix {
		return fmt.Errorf("bech32 prefix mismatch: expected: %v, got: %v", hub.Bech32Prefix, packet.Bech32Prefix)
	}

	if packet.NativeDenom != hub.NativeDenom {
		return fmt.Errorf("native denom mismatch: expected: %v, got: %v", hub.NativeDenom, packet.NativeDenom)
	}

	if !packet.InitialSupply.Equal(hub.InitialSupply) {
		return fmt.Errorf("initial supply mismatch: expected: %v, got: %v", hub.InitialSupply, packet.InitialSupply)
	}

	err := compareGenesisAccounts(hub.GenesisAccounts, packet.GenesisAccounts)
	if err != nil {
		return errorsmod.Wrap(err, "genesis accounts mismatch")
	}
}

func (e *genesisTransferValidator) getBankStuff() (*transfertypes.DenomTrace, *banktypes.Metadata, error) {
	var m banktypes.Metadata
	m = e.rollapp.NativeDenom
	trace := uibc.GetForeignDenomTrace(e.channelID, m.Base)

	// Change the base to the ibc denom, and add an alias to the original
	m.Base = trace.IBCDenom()
	m.Description = fmt.Sprintf("auto-generated ibc denom for hub: base: %s: hub: %s", m.GetBase(), e.rollappID)
	for i, u := range m.DenomUnits {
		if u.Exponent == 0 {
			m.DenomUnits[i].Aliases = append(m.DenomUnits[i].Aliases, u.Denom)
			m.DenomUnits[i].Denom = m.GetBase()
		}
	}

	if err := m.Validate(); err != nil {
		return nil, nil, errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "metadata validate")
	}
	return &trace, &m, nil
}

func (e *genesisTransferValidator) getFungiData() ([]transfertypes.FungibleTokenPacketData, error) {
	gTransfer := e.rollapp.GenesisTransfer
	required := e.hub.GenesisAccounts != nil
	// required but not present
	if required && gTransfer == nil {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer required")
	}
	// not required but present
	if !required && gTransfer != nil {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer not expected")
	}
	if gTransfer == nil {
		return nil, nil
	}

	// validate the receiver
	if gTransfer.Receiver != HubRecipient {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "receiver mismatch")
	}

	// validate that the transfer amount matches the expected amount, which is the sum of all genesis accounts
	expectedAmount := e.hub.GenesisTransferAmount()
	if expectedAmount.String() != gTransfer.Amount {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "amount mismatch")
	}

	var ret []transfertypes.FungibleTokenPacketData
	for _, acc := range e.hub.GenesisAccounts.Accounts {
		// create a new packet for each account
		data := transfertypes.NewFungibleTokenPacketData(
			gTransfer.Denom,
			acc.Amount.String(),
			gTransfer.Sender,
			acc.Address,
			"",
		)
		ret = append(ret, data)
	}
	return ret, nil
}
