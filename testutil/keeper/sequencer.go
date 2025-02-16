package keeper

import (
	"testing"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func SequencerKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	keys := storetypes.NewKVStoreKeys(types.StoreKey)
	logger := log.NewNopLogger()

	stateStore := integration.CreateMultiStore(keys, logger)
	params := params.MakeEncodingConfig()

	k := keeper.NewKeeper(
		params.Codec,
		keys[types.StoreKey],
		nil,
		&authkeeper.AccountKeeper{},
		&rollappkeeper.Keeper{},
		sample.AccAddress(),
	)

	ctx := sdk.NewContext(stateStore, cometbftproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}
