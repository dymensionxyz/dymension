package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

type Keeper struct {
	authority string // authority is the x/gov module account

	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	//FIXME: change to expected keeper
	AK *authkeeper.AccountKeeper
	bk bankkeeper.Keeper
	rk *rollappkeeper.Keeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ak *authkeeper.AccountKeeper,
	bk bankkeeper.Keeper,
	rk *rollappkeeper.Keeper,
) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		AK:       ak,
		bk:       bk,
		rk:       rk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetParams sets the module parameters in the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, b)
}

// GetParams returns the module parameters from the store
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.ParamsKey)
	if b == nil {
		panic("params should have been set")
	}

	k.cdc.MustUnmarshal(b, &params)
	return params
}