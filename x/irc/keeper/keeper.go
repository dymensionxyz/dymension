package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	porttypes "github.com/cosmos/ibc-go/v5/modules/core/05-port/types"
	"github.com/dymensionxyz/dymension/x/irc/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace

		bankKeeper    types.BankKeeper
		ibcKeeper     types.IBCKeeper
		channelKeeper types.ChannelKeeper
		rollappKeeper types.RollappKeeper
		ics4Wrapper   porttypes.ICS4Wrapper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,

	bankKeeper types.BankKeeper,
	ibcKeeper types.IBCKeeper,
	channelKeeper types.ChannelKeeper,
	rollappKeeper types.RollappKeeper,
	ics4Wrapper porttypes.ICS4Wrapper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramstore:    ps,
		bankKeeper:    bankKeeper,
		ibcKeeper:     ibcKeeper,
		channelKeeper: channelKeeper,
		rollappKeeper: rollappKeeper,
		ics4Wrapper:   ics4Wrapper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
