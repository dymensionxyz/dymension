package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	hooks      types.MultiRollappHooks
	paramstore paramtypes.Subspace

	ibcClientKeeper types.IBCClientKeeper
	channelKeeper   types.ChannelKeeper

	daLayers map[string]types.DALayer
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, ps paramtypes.Subspace, channelKeeper types.ChannelKeeper) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	daLayers := make(map[string]types.DALayer)
	daLayers[types.InterchainDALayerName] = types.NewInterchainDALayer()

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramstore:    ps,
		hooks:         nil,
		channelKeeper: channelKeeper,
		daLayers:      daLayers,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

/* -------------------------------------------------------------------------- */
/*                                    Hooks                                   */
/* -------------------------------------------------------------------------- */

func (k *Keeper) SetHooks(sh types.MultiRollappHooks) {
	if k.hooks != nil {
		panic("cannot set rollapp hooks twice")
	}
	k.hooks = sh
}

func (k *Keeper) GetHooks() types.MultiRollappHooks {
	return k.hooks
}
