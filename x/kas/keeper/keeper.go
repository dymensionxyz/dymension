package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"github.com/dymensionxyz/dymension/v3/internal/collcompat"

	storetypes "cosmossdk.io/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type Keeper struct {
	authority string // authority is the x/gov module account

	cdc      codec.BinaryCodec
	storeKey storetypes.Key
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.Key,
	authority string,
) *Keeper {
	_, err := sdk.AccAddressFromBech32(authority)
	if err != nil {
		panic(fmt.Errorf("invalid x/sequencer authority address: %w", err))
	}
	service := collcompat.NewKVStoreService(storeKey)
	sb := collections.NewSchemaBuilder(service)

	return &Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
