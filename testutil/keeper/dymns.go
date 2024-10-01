package keeper

import (
	"testing"
	"time"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/dymensionxyz/dymension/v3/app/params"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func DymNSKeeper(t testing.TB) (dymnskeeper.Keeper, dymnstypes.BankKeeper, rollappkeeper.Keeper, sdk.Context) {
	dymNsStoreKey := sdk.NewKVStoreKey(dymnstypes.StoreKey)
	dymNsMemStoreKey := storetypes.NewMemoryStoreKey(dymnstypes.MemStoreKey)

	authStoreKey := sdk.NewKVStoreKey(authtypes.StoreKey)

	bankStoreKey := sdk.NewKVStoreKey(banktypes.StoreKey)

	rollappStoreKey := sdk.NewKVStoreKey(rollapptypes.StoreKey)
	rollappMemStoreKey := storetypes.NewMemoryStoreKey(rollapptypes.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(dymNsStoreKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(dymNsMemStoreKey, storetypes.StoreTypeMemory, nil)
	stateStore.MountStoreWithDB(authStoreKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(bankStoreKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(rollappStoreKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(rollappMemStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	dymNSParamsSubspace := typesparams.NewSubspace(cdc,
		dymnstypes.Amino,
		dymNsStoreKey,
		dymNsMemStoreKey,
		"DymNSParams",
	)

	rollappParamsSubspace := typesparams.NewSubspace(cdc,
		rollapptypes.Amino,
		rollappStoreKey,
		rollappMemStoreKey,
		"RollappParams",
	)

	authKeeper := authkeeper.NewAccountKeeper(
		cdc,
		authStoreKey,
		authtypes.ProtoBaseAccount,
		map[string][]string{
			banktypes.ModuleName:  {authtypes.Minter, authtypes.Burner},
			dymnstypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		},
		params.AccountAddressPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	authtypes.RegisterInterfaces(registry)

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		bankStoreKey,
		authKeeper,
		map[string]bool{},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	banktypes.RegisterInterfaces(registry)

	rollappKeeper := rollappkeeper.NewKeeper(
		cdc,
		rollappStoreKey,
		rollappParamsSubspace,
		nil, nil, nil,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	k := dymnskeeper.NewKeeper(cdc,
		dymNsStoreKey,
		dymNSParamsSubspace,
		bankKeeper,
		rollappKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())
	ctx = ctx.WithBlockTime(time.Now().UTC())

	// Initialize params
	moduleParams := dymnstypes.DefaultParams()
	moduleParams.Chains.AliasesOfChainIds = nil
	err := k.SetParams(ctx, moduleParams)
	require.NoError(t, err)

	return k, bankKeeper, *rollappKeeper, ctx
}
