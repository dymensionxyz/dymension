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

	cdc            codec.BinaryCodec
	storeKey       storetypes.Key
	bankKeeper     types.BankKeeper
	accountK       types.AccountKeeper
	rollappKeeper  types.RollappKeeper
	unbondBlockers []UnbondBlocker
	hooks          types.Hooks

	dymintProposerAddrToAccAddr collections.Map[[]byte, string]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.Key,
	bankKeeper types.BankKeeper,
	accountK types.AccountKeeper,
	rollappKeeper types.RollappKeeper,
	authority string,
) *Keeper {
	_, err := sdk.AccAddressFromBech32(authority)
	if err != nil {
		panic(fmt.Errorf("invalid x/sequencer authority address: %w", err))
	}
	service := collcompat.NewKVStoreService(storeKey)
	sb := collections.NewSchemaBuilder(service)

	return &Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		bankKeeper:     bankKeeper,
		rollappKeeper:  rollappKeeper,
		accountK:       accountK,
		authority:      authority,
		unbondBlockers: []UnbondBlocker{},
		hooks:          types.NoOpHooks{},
		dymintProposerAddrToAccAddr: collections.NewMap(
			sb,
			types.DymintProposerAddrToAccAddrKeyPrefix,
			"dymintProposerAddrToAccAddr",
			collections.BytesKey,
			collections.StringValue,
		),
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *Keeper) SetUnbondBlockers(ubs ...UnbondBlocker) {
	k.unbondBlockers = ubs
}

func (k *Keeper) SetHooks(h types.Hooks) {
	k.hooks = h
}
