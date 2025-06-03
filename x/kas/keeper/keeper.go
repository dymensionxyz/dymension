package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	hypercorekeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/keeper"
	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/kas/types"
)

type Keeper struct {
	authority string // authority is the x/gov module account
	cdc       codec.BinaryCodec

	hypercoreK *hypercorekeeper.Keeper

	outpoint             collections.Item[types.TransactionOutpoint]
	processedWithdrawals collections.KeySet[collections.Pair[uint64, []byte]]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	service store.KVStoreService,
	authority string,
	hypercoreK *hypercorekeeper.Keeper,
) *Keeper {
	_, err := sdk.AccAddressFromBech32(authority)
	if err != nil {
		panic(fmt.Errorf("invalid x/sequencer authority address: %w", err))
	}
	sb := collections.NewSchemaBuilder(service)

	outpoint := collections.NewItem(sb, collections.NewPrefix(types.KeyOutpoint),
		types.KeyOutpoint,
		collcompat.ProtoValue[types.TransactionOutpoint](cdc))

	processedWithdrawals := collections.NewKeySet(sb, collections.NewPrefix(types.KeyProcessedWithdrawals),
		types.KeyProcessedWithdrawals,
		collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey))

	return &Keeper{
		cdc:                  cdc,
		authority:            authority,
		hypercoreK:           hypercoreK,
		outpoint:             outpoint,
		processedWithdrawals: processedWithdrawals,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
