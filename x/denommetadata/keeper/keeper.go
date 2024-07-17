package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

// Keeper of the denommetadata store
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

func (k *Keeper) HasDenomMetadata(ctx sdk.Context, base string) bool {
	_, found := k.bankKeeper.GetDenomMetaData(ctx, base)
	return found
}

// CreateDenomMetadata creates a new denommetadata
func (k *Keeper) CreateDenomMetadata(ctx sdk.Context, metadata banktypes.Metadata) error {
	found := k.HasDenomMetadata(ctx, metadata.Base)
	if found {
		return gerrc.ErrAlreadyExists
	}
	k.bankKeeper.SetDenomMetaData(ctx, metadata)
	err := k.hooks.AfterDenomMetadataCreation(ctx, metadata)
	if err != nil {
		return err
	}
	return nil
}

// UpdateDenomMetadata returns the denommetadata of the specified denom
func (k *Keeper) UpdateDenomMetadata(ctx sdk.Context, metadata banktypes.Metadata) error {
	found := k.HasDenomMetadata(ctx, metadata.Base)
	if !found {
		return gerrc.ErrNotFound
	}
	k.bankKeeper.SetDenomMetaData(ctx, metadata)
	err := k.hooks.AfterDenomMetadataUpdate(ctx, metadata)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

/* -------------------------------------------------------------------------- */
/*                                    Hooks                                   */
/* -------------------------------------------------------------------------- */

// SetHooks sets the hooks for the denommetadata keeper
func (k *Keeper) SetHooks(sh types.MultiDenomMetadataHooks) {
	k.hooks = sh
}

// GetHooks returns the hooks for the denommetadata keeper
func (k *Keeper) GetHooks() types.MultiDenomMetadataHooks {
	return k.hooks
}
