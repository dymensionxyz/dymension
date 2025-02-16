package keeper

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	"cosmossdk.io/log"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

func RollappKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	keys := storetypes.NewKVStoreKeys(types.StoreKey)
	logger := log.NewNopLogger()

	stateStore := integration.CreateMultiStore(keys, logger)
	params := params.MakeEncodingConfig()

	subspace := paramstypes.NewSubspace(params.Codec, params.Amino, keys[paramstypes.StoreKey], nil, "rollapp")

	k := keeper.NewKeeper(params.Codec, keys[types.StoreKey], subspace, nil, nil, nil, nil, authtypes.NewModuleAddress(govtypes.ModuleName).String(), nil)

	ctx := sdk.NewContext(stateStore, cometbftproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}
