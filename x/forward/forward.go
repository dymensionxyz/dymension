package forward

import (
	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
	warptypes "github.com/dymensionxyz/hyperlane-cosmos/x/warp/types"

	transferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
)

type Forward struct {
	warpQ     types.WarpQuery
	warpS     warptypes.MsgServer
	transferK transferkeeper.Keeper
	bankK     types.BankKeeper
	accountK  types.AccountKeeper
}

func New(
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	transferKeeper transferkeeper.Keeper,
	warpQueryServer warptypes.QueryServer,
	warpMsgServer warptypes.MsgServer,
) *Forward {

	return &Forward{
		bankK:     bankKeeper,
		accountK:  accountKeeper,
		transferK: transferKeeper,
		warpQ:     warpQueryServer,
		warpS:     warpMsgServer,
	}
}

func (k Forward) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/forward")
}
