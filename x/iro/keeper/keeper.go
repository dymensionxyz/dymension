package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type Keeper struct {
	rollapptypes.StubRollappCreatedHooks
	authority string // authority is the x/gov module account

	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	AK types.AccountKeeper
	BK types.BankKeeper
	rk types.RollappKeeper
	gk types.GammKeeper
	pm types.PoolManagerKeeper
	ik types.IncentivesKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	rk types.RollappKeeper,
	gk types.GammKeeper,
	ik types.IncentivesKeeper,
	pm types.PoolManagerKeeper,
) *Keeper {
	return &Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
		AK:        ak,
		BK:        bk,
		rk:        rk,
		gk:        gk,
		ik:        ik,
		pm:        pm,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetModuleAccountAddress returns the address of the module account
func (k Keeper) GetModuleAccountAddress() string {
	return k.AK.GetModuleAddress(types.ModuleName).String()
}
