package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/dymensionxyz/dymension/x/eibc/types"
    "github.com/dymensionxyz/dymension/x/eibc/keeper"
    keepertest "github.com/dymensionxyz/dymension/testutil/keeper"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.EibcKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
