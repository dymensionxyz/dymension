package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	incentiveskeeper "github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
)

type Keeper struct {
	rollapptypes.StubRollappCreatedHooks
	authority string // authority is the x/gov module account

	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	// FIXME: change to expected keeper interface
	AK *authkeeper.AccountKeeper
	BK bankkeeper.Keeper
	rk *rollappkeeper.Keeper
	gk *gammkeeper.Keeper
	ik *incentiveskeeper.Keeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ak *authkeeper.AccountKeeper,
	bk bankkeeper.Keeper,
	rk *rollappkeeper.Keeper,
	gk *gammkeeper.Keeper,
	ik *incentiveskeeper.Keeper,
) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		AK:       ak,
		BK:       bk,
		rk:       rk,
		gk:       gk,
		ik:       ik,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetModuleAccountAddress returns the address of the module account
func (k Keeper) GetModuleAccountAddress() string {
	return k.AK.GetModuleAddress(types.ModuleName).String()
}
