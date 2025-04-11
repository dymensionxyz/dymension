package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"

	transferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
)

type Keeper struct {
	cdc    codec.BinaryCodec
	Schema collections.Schema
	// TODO: params collection
	warpK     types.WarpRouteKeeper
	warpQ     types.WarpQuery
	warpS     warptypes.MsgServer
	transferK transferkeeper.Keeper // TODO: interface
	bankK     types.BankKeeper
	accountK  types.AccountKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	service storetypes.KVStoreService,
	warpKeeper types.WarpRouteKeeper,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	transferKeeper transferkeeper.Keeper,
	warpQueryServer warptypes.QueryServer,
	warpMsgServer warptypes.MsgServer,
) *Keeper {

	return &Keeper{
		cdc:       cdc,
		warpK:     warpKeeper,
		bankK:     bankKeeper,
		accountK:  accountKeeper,
		transferK: transferKeeper,
		warpQ:     warpQueryServer,
		warpS:     warpMsgServer,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
