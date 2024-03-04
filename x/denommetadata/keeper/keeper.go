package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
func (k Keeper) CreateDenomMetadata(ctx sdk.Context, coins sdk.Coins, records []types.DistrRecord, startTime time.Time, epochIdentifier string, numEpochsPaidOver uint64) (uint64, error) {

	return 0, nil
}

// TerminateStream cancels a stream.
func (k Keeper) RemoveDenomMetadata(ctx sdk.Context, streamID uint64) error {
	return nil
}
