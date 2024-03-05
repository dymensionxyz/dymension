package keeper

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper provides a way to manage streamer module storage.
type Keeper struct {
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace
}

// NewKeeper returns a new instance of the incentive module keeper struct.
func NewKeeper(storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,
	}
}

// Logger returns a logger instance for the streamer module.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// CreateDenomMetadata creates a new denom metadata record.
func (k Keeper) CreateDenomMetadata(ctx sdk.Context, record types.TokenMetadata) (uint64, error) {

	err := record.Validate()
	if err != nil {
		return 0, err
	}
	denomMetadata := types.DenomMetadata{
		Id:            k.GetLastDenomMetadataID(ctx) + 1,
		TokenMetadata: record,
	}

	err = k.setDenomMetadata(ctx, &denomMetadata)
	if err != nil {
		return 0, err
	}
	k.SetLastDenomMetadataID(ctx, denomMetadata.Id)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCreateDenomMetadata,
			sdk.NewAttribute(types.AttributeDenomMetadataID, osmoutils.Uint64ToString(denomMetadata.Id)),
		),
	})

	return denomMetadata.Id, nil
}
