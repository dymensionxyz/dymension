package keeper

import (
	"testing"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"

	"github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func IROKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	keys := storetypes.NewKVStoreKeys(types.StoreKey)
	logger := log.NewNopLogger()

	multiStore := integration.CreateMultiStore(keys, logger)
	params := params.MakeEncodingConfig()

	keeperInstance := keeper.NewKeeper(
		params.Codec,
		keys[types.StoreKey],
		"",
		nil, nil, nil, nil, nil, nil, nil, nil,
	)

	context := sdk.NewContext(multiStore, tmproto.Header{}, false, logger)

	keeperInstance.SetParams(context, types.DefaultParams())

	return keeperInstance, context
}
