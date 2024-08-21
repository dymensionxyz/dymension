package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"

	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper of the DymNS store
type Keeper struct {
	authority string // authority is the x/gov module account

	cdc           codec.BinaryCodec
	storeKey      storetypes.StoreKey
	paramStore    paramtypes.Subspace
	bankKeeper    dymnstypes.BankKeeper
	rollappKeeper dymnstypes.RollAppKeeper
}

// NewKeeper returns a new instance of the DymNS keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	ps paramtypes.Subspace,
	bk dymnstypes.BankKeeper,
	rk dymnstypes.RollAppKeeper,
	authority string,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(dymnstypes.ParamKeyTable())
	}
	return Keeper{
		authority: authority,

		cdc:           cdc,
		storeKey:      key,
		paramStore:    ps,
		bankKeeper:    bk,
		rollappKeeper: rk,
	}
}

// Logger returns a module-specific logger.
func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", dymnstypes.ModuleName))
}
