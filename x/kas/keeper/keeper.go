package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"github.com/dymensionxyz/dymension/v3/internal/collcompat"

	storetypes "cosmossdk.io/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/kas/types"
)

type Keeper struct {
	authority string // authority is the x/gov module account

	cdc      codec.BinaryCodec
	storeKey storetypes.Key
}

func (k Keeper) Foo(context.Context, *types.QueryFooRequest) (*types.QueryFooResponse, error) {
	panic("unimplemented")
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
	_ = sb

	return &Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
