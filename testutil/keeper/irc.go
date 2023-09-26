package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	rollappkeeper "github.com/dymensionxyz/dymension/x/rollapp/keeper"

	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/dymensionxyz/dymension/x/irc/keeper"
	"github.com/dymensionxyz/dymension/x/irc/types"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	ibc "github.com/cosmos/ibc-go/v6/modules/core/types"
)

func IRCKeeper(t testing.TB) (*keeper.Keeper, *rollappkeeper.Keeper, sdk.Context) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	// create codec
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	ibc.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	// create roolapp keeper
	rollappParamsSubspace := typesparams.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"RollappParams",
	)
	rollappKeeper := rollappkeeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		rollappParamsSubspace,
	)

	paramsSubspace := typesparams.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"IrcParams",
	)

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		paramsSubspace,
		nil,
		nil,
		rollappKeeper,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, rollappKeeper, ctx
}
