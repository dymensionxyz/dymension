package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/dymensionxyz/dymension/x/streamer/types"

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
	ck         types.CommunityPoolKeeper
	tk         types.TxFeesKeeper
}

// NewKeeper returns a new instance of the incentive module keeper struct.
func NewKeeper(storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace, bk types.BankKeeper, ek types.EpochKeeper, ck types.CommunityPoolKeeper, txfk types.TxFeesKeeper) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,
		bk:         bk,
		ek:         ek,
		ck:         ck,
		tk:         txfk,
	}
}

// Logger returns a logger instance for the streamer module.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
