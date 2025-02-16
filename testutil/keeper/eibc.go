package keeper

import (
	"testing"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func EibcKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	keys := storetypes.NewKVStoreKeys(types.StoreKey)

	logger := log.NewNopLogger()
	stateStore := integration.CreateMultiStore(keys, logger)

	codec := params.MakeEncodingConfig()
	registry := codec.InterfaceRegistry
	cdc := codec.Codec

	storeKey := keys[types.StoreKey]
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		nil,
		paramstypes.Subspace{},
		nil,
		nil,
		nil,
		nil,
	)
	types.RegisterInterfaces(registry)

	ctx := sdk.NewContext(stateStore, cometbftproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}
