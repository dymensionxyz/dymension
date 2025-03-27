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

	// import eibc transfer keeper
	transferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
)

var _ types.MsgServer = Keeper{}
var _ types.QueryServer = Keeper{}

type Keeper struct {
	cdc    codec.BinaryCodec
	Schema collections.Schema
	// TODO: params collection
	warpKeeper     types.WarpRouteKeeper
	warpServer     warptypes.MsgServer
	transferKeeper transferkeeper.Keeper // TODO: interface
	bankKeeper     types.BankKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	service storetypes.KVStoreService,
	warpKeeper types.WarpRouteKeeper,
) *Keeper {

	sb := collections.NewSchemaBuilder(service)
	// TODO: Add collections
	_ = sb

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	return &Keeper{
		cdc:        cdc,
		Schema:     schema,
		warpKeeper: warpKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
