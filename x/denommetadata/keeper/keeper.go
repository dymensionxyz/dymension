package keeper

import (
	"fmt"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// denommetadata Keeper
type Keeper struct {
	bankKeeper types.BankKeeper
	hooks      types.MultiDenomMetadataHooks
}

// NewKeeper returns a new instance of the denommetadata keeper
func NewKeeper(bankKeeper types.BankKeeper) *Keeper {

	return &Keeper{
		bankKeeper: bankKeeper,
		hooks:      nil,
	}
}

func (k *Keeper) CreateDenomMetadata(ctx sdk.Context, metadata banktypes.Metadata) error {
	found := k.bankKeeper.HasDenomMetaData(ctx, metadata.Base)
	if found {
		return types.ErrDenomAlreadyExists
	}
	k.bankKeeper.SetDenomMetaData(ctx, metadata)
	err := k.hooks.AfterDenomMetadataCreation(ctx, metadata)
	if err != nil {
		return err
	}
	return nil

}

func (k *Keeper) UpdateDenomMetadata(ctx sdk.Context, metadata banktypes.Metadata) error {
	found := k.bankKeeper.HasDenomMetaData(ctx, metadata.Base)
	if !found {
		return types.ErrDenomDoesNotExist
	}
	k.bankKeeper.SetDenomMetaData(ctx, metadata)
	err := k.hooks.AfterDenomMetadataUpdate(ctx, metadata)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

/* -------------------------------------------------------------------------- */
/*                                    Hooks                                   */
/* -------------------------------------------------------------------------- */

// Set the denommetadata hooks
func (k *Keeper) SetHooks(sh types.MultiDenomMetadataHooks) {
	if k.hooks != nil {
		panic("cannot set rollapp hooks twice")
	}
	k.hooks = sh
}

// Get the denommetadata hooks
func (k *Keeper) GetHooks() types.MultiDenomMetadataHooks {
	return k.hooks
}
