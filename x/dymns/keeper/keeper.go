package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store"
	"github.com/cosmos/cosmos-sdk/codec"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper of the DymNS store
type Keeper struct {
	authority string // authority is the x/gov module account

	cdc           codec.BinaryCodec
	storeKey      storetypes.Key
	bankKeeper    dymnstypes.BankKeeper
	rollappKeeper dymnstypes.RollAppKeeper
	txFeesKeeper  dymnstypes.TxFeesKeeper
}

// NewKeeper returns a new instance of the DymNS keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.Key,
	bk dymnstypes.BankKeeper,
	rk dymnstypes.RollAppKeeper,
	tk dymnstypes.TxFeesKeeper,
	authority string,
) Keeper {
	return Keeper{
		authority:     authority,
		cdc:           cdc,
		storeKey:      key,
		bankKeeper:    bk,
		rollappKeeper: rk,
		txFeesKeeper:  tk,
	}
}

// Logger returns a module-specific logger.
func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", dymnstypes.ModuleName))
}
