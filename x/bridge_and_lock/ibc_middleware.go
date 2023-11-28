package bridgeandlock

import (
	"encoding/json"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/dymensionxyz/dymension/x/bridge_and_lock/types"
)

var _ porttypes.Middleware = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks
type IBCMiddleware struct {
	app           porttypes.IBCModule
	accountKeeper types.AccountKeeper
	ics4Wrapper   porttypes.ICS4Wrapper
	stakingkeeper types.StakingKeeper
	lockupkeeper  types.LockupKeeper
}

// NewIBCMiddleware creates a new IBCMiddlware given the keeper and underlying application
func NewIBCMiddleware(app porttypes.IBCModule, ak types.AccountKeeper, ics4 porttypes.ICS4Wrapper, sk types.StakingKeeper, lk types.LockupKeeper) IBCMiddleware {
	return IBCMiddleware{
		app:           app,
		accountKeeper: ak,
		ics4Wrapper:   ics4,
		stakingkeeper: sk,
		lockupkeeper:  lk,
	}
}

// OnChanOpenInit implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID,
		chanCap, counterparty, version)
}

// OnChanOpenTry implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	// call underlying app's OnChanOpenAck callback with the counterparty app version.
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// call underlying app's OnChanOpenConfirm callback.
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket implements the IBCMiddleware interface.
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {

	logger := ctx.Logger().With("module", "bridge_and_lock")

	// no-op if the packet is not a fungible token packet
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		//error will be handled by the underlying layer
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	//Check if packet contains lock memo
	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(data.Memo), &d)
	if err != nil || d[types.LockMemoName] == nil {
		// not a packet that should be handled by this middleware
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	lockObject, ok := d[types.LockMemoName].(map[string]interface{})
	if !ok {
		logger.Error("failed to parse lock object", "memo", data.Memo)
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	// Marshal the map back into a JSON byte slice
	lockMemoData, err := json.Marshal(lockObject)
	if err != nil {
		logger.Error("error parsing lock metadata", "error", err, "memo", data.Memo)
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	m := &types.LockMemo{}
	err = json.Unmarshal(lockMemoData, m)
	if err != nil {
		logger.Error("error parsing lock metadata", "error", err, "memo", data.Memo)
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	if !m.ToLock {
		logger.Error("skipping locking", "error", err, "memo", data.Memo)
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	/* --------------------------- handle the transfer -------------------------- */
	ack := im.app.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}

	// use a zero gas config to avoid extra costs for the relayers
	ctx = ctx.
		WithKVGasConfig(storetypes.GasConfig{}).
		WithTransientKVGasConfig(storetypes.GasConfig{})

	// decode the receiver address
	owner, err := sdk.AccAddressFromBech32(data.Receiver)
	if err != nil {
		logger.Error("error parsing receiver", "error", err)
		return ack
	}

	//validate owner is not a module
	ownerAcc := im.accountKeeper.GetAccount(ctx, owner)

	// return acknoledgement without conversion if owner is a module account
	if IsModuleAccount(ownerAcc) {
		return ack
	}

	// parse the transfer amount
	coin := GetReceivedCoin(
		packet.SourcePort, packet.SourceChannel,
		packet.DestinationPort, packet.DestinationChannel,
		data.Denom, data.Amount,
	)

	// check if the coin is a native staking token
	bondDenom := im.stakingkeeper.BondDenom(ctx)
	if coin.Denom == bondDenom {
		// no-op, received coin is the staking denomination
		return ack
	}

	// check if there's an existing lock from the same owner with the same duration.
	// If so, simply add tokens to the existing lock.
	lockExists := im.lockupkeeper.HasLock(ctx, owner, coin.Denom, types.DefaultLockDuration)
	if lockExists {
		logger.Info("adding to existing lock", "owner", owner, "coin", coin)
		_, err := im.lockupkeeper.AddToExistingLock(ctx, owner, coin, types.DefaultLockDuration)
		if err != nil {
			logger.Error("error adding to existing lock", "error", err)
			return ack
		}
	} else {
		// if the owner + duration combination is new, create a new lock.
		_, err := im.lockupkeeper.CreateLock(ctx, owner, sdk.Coins{coin}, types.DefaultLockDuration)
		if err != nil {
			logger.Error("error creating lock", "error", err)
			return ack
		}
	}
	return ack
}

// OnAcknowledgementPacket implements the IBCMiddleware interface
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCMiddleware interface
func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	// call underlying callback
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

/* ------------------------------- ICS4Wrapper ------------------------------ */

// SendPacket implements the ICS4 Wrapper interface
func (im IBCMiddleware) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	return im.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// WriteAcknowledgement implements the ICS4 Wrapper interface
func (im IBCMiddleware) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
	ack exported.Acknowledgement,
) error {
	return im.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

// GetAppVersion returns the application version of the underlying application
func (im IBCMiddleware) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return im.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}
