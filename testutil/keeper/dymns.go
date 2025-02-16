package keeper

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/runtime"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/cosmos-sdk/testutil/integration"

	"github.com/dymensionxyz/dymension/v3/app/params"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func DymNSKeeper(t testing.TB) (dymnskeeper.Keeper, dymnstypes.BankKeeper, rollappkeeper.Keeper, sdk.Context) {
	keys := storetypes.NewKVStoreKeys(dymnstypes.StoreKey, authtypes.StoreKey, banktypes.StoreKey, rollapptypes.StoreKey)

	logger := log.NewNopLogger()
	stateStore := integration.CreateMultiStore(keys, logger)

	codec := params.MakeEncodingConfig()
	registry := codec.InterfaceRegistry
	cdc := codec.Codec

	authKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		map[string][]string{
			banktypes.ModuleName:  {authtypes.Minter, authtypes.Burner},
			dymnstypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		},
		addresscodec.NewBech32Codec(params.AccountAddressPrefix),
		params.AccountAddressPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	authtypes.RegisterInterfaces(registry)

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		authKeeper,
		map[string]bool{},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)
	banktypes.RegisterInterfaces(registry)

	rollappKeeper := rollappkeeper.NewKeeper(
		cdc,
		keys[rollapptypes.StoreKey],
		paramstypes.Subspace{},
		nil, nil,
		bankKeeper,
		nil,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		nil,
	)

	k := dymnskeeper.NewKeeper(cdc,
		keys[dymnstypes.StoreKey],
		paramstypes.Subspace{},
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
