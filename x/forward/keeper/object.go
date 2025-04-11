package keeper

import (
	"cosmossdk.io/log"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"

	transferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
)

type Keeper struct {
	// TODO: params collection
	warpK     types.WarpRouteKeeper
	warpQ     types.WarpQuery
	warpS     warptypes.MsgServer
	transferK transferkeeper.Keeper // TODO: interface
	bankK     types.BankKeeper
	accountK  types.AccountKeeper
}

func NewKeeper(
	warpKeeper types.WarpRouteKeeper,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	transferKeeper transferkeeper.Keeper,
	warpQueryServer warptypes.QueryServer,
	warpMsgServer warptypes.MsgServer,
) *Keeper {

	return &Keeper{
		warpK:     warpKeeper,
		bankK:     bankKeeper,
		accountK:  accountKeeper,
		transferK: transferKeeper,
		warpQ:     warpQueryServer,
		warpS:     warpMsgServer,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/forward")
}
