package keeper

import (
	"fmt"
	"time"

	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper provides a way to manage streamer module storage.
type Keeper struct {
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace
	bk         types.BankKeeper
	ek         types.EpochKeeper
	ak         types.AccountKeeper
	ik         types.IncentivesKeeper
}

// NewKeeper returns a new instance of the incentive module keeper struct.
func NewKeeper(storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace, bk types.BankKeeper, ek types.EpochKeeper, ak types.AccountKeeper, ik types.IncentivesKeeper) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,
		bk:         bk,
		ek:         ek,
		ak:         ak,
		ik:         ik,
	}
}

// Logger returns a logger instance for the streamer module.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// CreateStream creates a stream and sends coins to the stream.
func (k Keeper) CreateStream(ctx sdk.Context, coins sdk.Coins, records []types.DistrRecord, startTime time.Time, epochIdentifier string, numEpochsPaidOver uint64) (uint64, error) {
	if !coins.IsAllPositive() {
		return 0, fmt.Errorf("all coins %s must be positive", coins)
	}

	distrInfo, err := k.NewDistrInfo(ctx, records)
	if err != nil {
		return 0, err
	}

	//TODO: it's better to check only the denoms of the requested coins. No need to iterate entire balance.
	moduleBalance := k.bk.GetAllBalances(ctx, authtypes.NewModuleAddress(types.ModuleName))
	alreadyAllocatedCoins := k.GetModuleToDistributeCoins(ctx)

	if !coins.IsAllLTE(moduleBalance.Sub(alreadyAllocatedCoins...)) {
		return 0, fmt.Errorf("insufficient module balance to distribute coins")
	}

	if (k.ek.GetEpochInfo(ctx, epochIdentifier) == epochstypes.EpochInfo{}) {
		return 0, fmt.Errorf("epoch identifier does not exist: %s", epochIdentifier)
	}

	if numEpochsPaidOver <= 0 {
		return 0, fmt.Errorf("numEpochsPaidOver must be greater than 0")
	}

	if startTime.Before(ctx.BlockTime()) {
		ctx.Logger().Info("start time is before current block time, setting start time to current block time")
		startTime = ctx.BlockTime()
	}

	stream := types.NewStream(
		k.GetLastStreamID(ctx)+1,
		distrInfo,
		coins.Sort(),
		startTime,
		epochIdentifier,
		numEpochsPaidOver,
	)

	err = k.setStream(ctx, &stream)
	if err != nil {
		return 0, err
	}
	k.SetLastStreamID(ctx, stream.Id)

	combinedKeys := combineKeys(types.KeyPrefixUpcomingStreams, getTimeKey(stream.StartTime))
	err = k.CreateStreamRefKeys(ctx, &stream, combinedKeys)
	if err != nil {
		return 0, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCreateStream,
			sdk.NewAttribute(types.AttributeStreamID, osmoutils.Uint64ToString(stream.Id)),
		),
	})

	return stream.Id, nil
}

// TerminateStream cancels a stream.
func (k Keeper) TerminateStream(ctx sdk.Context, streamID uint64) error {
	stream, err := k.GetStreamByID(ctx, streamID)
	if err != nil {
		return err
	}

	if stream.IsActiveStream(ctx.BlockTime()) {
		return k.moveActiveStreamToFinishedStream(ctx, *stream)
	} else if stream.IsUpcomingStream(ctx.BlockTime()) {
		return k.moveUpcomingStreamToFinishedStream(ctx, *stream)
	} else {
		return fmt.Errorf("stream %d is not active or upcoming", streamID)
	}
}
