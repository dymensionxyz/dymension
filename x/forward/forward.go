package forward

import (
	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

type Forward struct {
	warpQ     types.WarpQuery
	warpS     types.WarpMsgServer
	transferK types.TransferKeeper
}

func New(
	transferKeeper types.TransferKeeper,
	warpQueryServer types.WarpQuery,
	warpMsgServer types.WarpMsgServer,
) *Forward {
	return &Forward{
		transferK: transferKeeper,
		warpQ:     warpQueryServer,
		warpS:     warpMsgServer,
	}
}

func (k Forward) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}
