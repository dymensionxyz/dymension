package keeper

import (
	"fmt"
	"slices"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"

	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// Keeper provides a way to manage streamer module storage.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	bk       types.BankKeeper
	ek       types.EpochKeeper
	ak       types.AccountKeeper
	ik       types.IncentivesKeeper
	sk       types.SponsorshipKeeper

	mintParams        types.MintParamsGetter
	iroKeeper         types.IROKeeper
	poolManagerKeeper types.PoolManagerKeeper
	rollappKeeper     types.RollappKeeper
	txFeesKeeper      types.TxFeesKeeper
	gammKeeper        types.GAMMKeeper

	authority string

	// epochPointers holds a mapping from the epoch identifier to EpochPointer.
	epochPointers collections.Map[string, types.EpochPointer]
}

// NewKeeper returns a new instance of the streamer module keeper struct.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	bk types.BankKeeper,
	ek types.EpochKeeper,
	ak types.AccountKeeper,
	ik types.IncentivesKeeper,
	sk types.SponsorshipKeeper,
	mintParams types.MintParamsGetter,
	iroKeeper types.IROKeeper,
	poolManagerKeeper types.PoolManagerKeeper,
	rollappKeeper types.RollappKeeper,
	txFeesKeeper types.TxFeesKeeper,
	gammKeeper types.GAMMKeeper,
	authority string,
) *Keeper {
	sb := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))

	return &Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		bk:                bk,
		ek:                ek,
		ak:                ak,
		ik:                ik,
		sk:                sk,
		mintParams:        mintParams,
		iroKeeper:         iroKeeper,
		poolManagerKeeper: poolManagerKeeper,
		rollappKeeper:     rollappKeeper,
		txFeesKeeper:      txFeesKeeper,
		gammKeeper:        gammKeeper,
		authority:         authority,
		epochPointers: collections.NewMap(
			sb,
			types.KeyPrefixEpochPointers,
			"epoch_pointers",
			collections.StringKey,
			collcompat.ProtoValue[types.EpochPointer](cdc),
		),
	}
}

// Logger returns a logger instance for the streamer module.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// CreateStream creates a stream and sends coins to the stream.
func (k Keeper) CreateStream(
	ctx sdk.Context,
	coins sdk.Coins,
	records []types.DistrRecord,
	startTime time.Time,
	epochIdentifier string,
	numEpochsPaidOver uint64,
	sponsored bool,
) (uint64, error) {
	err := k.ValidateStreamParams(ctx, coins, epochIdentifier, numEpochsPaidOver)
	if err != nil {
		return 0, fmt.Errorf("invalid stream params: %w", err)
	}

	var distrInfo types.DistrInfo
	switch {
	case sponsored:
		// Sponsored Stream
		distr, err := k.sk.GetDistribution(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to get sponsorship distribution: %w", err)
		}
		distrInfo = types.DistrInfoFromDistribution(distr)

	default:
		// Usual Stream
		distr, err := k.NewDistrInfo(ctx, records)
		if err != nil {
			return 0, err
		}
		distrInfo = distr
	}

	return k.newStream(
		ctx,
		distrInfo,
		coins.Sort(),
		startTime,
		epochIdentifier,
		numEpochsPaidOver,
		sponsored,
		nil,
	)
}

func (k Keeper) CreatePumpStream(
	ctx sdk.Context,
	stream types.CreateStreamGeneric,
	numPumps uint64,
	pumpDistr types.PumpDistr,
	pumpTarget types.PumpTarget,
) (uint64, error) {
	err := k.ValidateStreamParams(ctx, stream.Coins, stream.EpochIdentifier, stream.NumEpochsPaidOver)
	if err != nil {
		return 0, fmt.Errorf("validate stream: %w", err)
	}

	err = types.ValidatePumpStreamParams(stream.Coins, numPumps, pumpDistr, pumpTarget)
	if err != nil {
		return 0, fmt.Errorf("validate pump stream: %w", err)
	}

	params := types.PumpParams{
		Target:         nil, // filled below
		EpochCoinsLeft: stream.Coins.QuoInt(math.NewIntFromUint64(stream.NumEpochsPaidOver)),
		NumPumps:       numPumps,
		PumpDistr:      pumpDistr,
	}

	// Stateful validation
	switch t := pumpTarget.(type) {
	case *types.MsgCreatePumpStream_Pool:
		tokenIn := stream.Coins[0].Denom
		tokenOut := t.Pool.TokenOut

		if tokenIn == tokenOut {
			return 0, fmt.Errorf("token out must not be the same as the stream coin: stream coin: %s, token out: %s", tokenIn, tokenOut)
		}
		denoms, err := k.gammKeeper.GetPoolDenoms(ctx, t.Pool.PoolId)
		if err != nil {
			return 0, fmt.Errorf("failed to get pool denoms: %w", err)
		}
		if !slices.Contains(denoms, tokenOut) {
			return 0, fmt.Errorf("token out must be in pool denoms: pool ID: %d, pool denoms: %s, token out: %s", t.Pool.PoolId, denoms, tokenOut)
		}
		if !slices.Contains(denoms, tokenIn) {
			return 0, fmt.Errorf("stream coin must be in pool denoms: pool ID: %d, pool denoms: %s, stream coin: %s", t.Pool.PoolId, denoms, tokenIn)
		}
		params.Target = &types.PumpParams_Pool{Pool: t.Pool}

	case *types.MsgCreatePumpStream_Rollapps:
		baseDenom, err := k.txFeesKeeper.GetBaseDenom(ctx)
		if err != nil {
			return 0, fmt.Errorf("get base denom: %w", err)
		}
		if stream.Coins[0].Denom != baseDenom {
			return 0, fmt.Errorf("pump stream must have one coin with base denom: base denom: %s, stream coin: %s", baseDenom, stream.Coins[0].Denom)
		}
		params.Target = &types.PumpParams_Rollapps{Rollapps: t.Rollapps}
	}

	return k.newStream(
		ctx,
		types.DistrInfo{},
		stream.Coins,
		stream.StartTime,
		stream.EpochIdentifier,
		stream.NumEpochsPaidOver,
		false,
		&params,
	)
}

func (k Keeper) newStream(
	ctx sdk.Context,
	distrTo types.DistrInfo,
	coins sdk.Coins,
	startTime time.Time,
	epochIdentifier string,
	numEpochsPaidOver uint64,
	sponsored bool,
	pumpParams *types.PumpParams,
) (uint64, error) {
	if startTime.Before(ctx.BlockTime()) {
		ctx.Logger().Info("start time is before current block time, setting start time to current block time")
		startTime = ctx.BlockTime()
	}

	stream := types.NewStream(
		k.GetLastStreamID(ctx)+1,
		distrTo,
		coins,
		startTime,
		epochIdentifier,
		numEpochsPaidOver,
		sponsored,
		pumpParams,
	)

	err := k.SetStream(ctx, &stream)
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

func (k Keeper) ValidateGenericStream(ctx sdk.Context, stream types.CreateStreamGeneric) error {
	return k.ValidateStreamParams(ctx, stream.Coins, stream.EpochIdentifier, stream.NumEpochsPaidOver)
}

func (k Keeper) ValidateStreamParams(
	ctx sdk.Context,
	coins sdk.Coins,
	epochIdentifier string,
	numEpochsPaidOver uint64,
) error {
	if !coins.IsAllPositive() {
		return fmt.Errorf("all coins %s must be positive", coins)
	}

	moduleBalance := k.bk.GetAllBalances(ctx, authtypes.NewModuleAddress(types.ModuleName))
	alreadyAllocatedCoins := k.GetModuleToDistributeCoins(ctx)
	if !coins.IsAllLTE(moduleBalance.Sub(alreadyAllocatedCoins...)) {
		return fmt.Errorf("insufficient module balance to distribute coins")
	}

	if (k.ek.GetEpochInfo(ctx, epochIdentifier) == epochstypes.EpochInfo{}) {
		return fmt.Errorf("epoch identifier does not exist: %s", epochIdentifier)
	}

	if numEpochsPaidOver <= 0 {
		return fmt.Errorf("numEpochsPaidOver must be greater than 0")
	}

	return nil
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
