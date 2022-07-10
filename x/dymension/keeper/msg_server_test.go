package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dYmension/testutil/keeper"
	"github.com/dymensionxyz/dYmension/x/dymension/keeper"
	"github.com/dymensionxyz/dYmension/x/dymension/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.DymensionKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
