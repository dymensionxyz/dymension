package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

// Keeper of the denommetadata store
type Keeper struct {
	bankKeeper types.BankKeeper
	rk         types.RollappKeeper
	hooks      types.MultiDenomMetadataHooks
	warpK      *warpkeeper.Keeper
}

// NewKeeper returns a new instance of the denommetadata keeper
func NewKeeper(bankKeeper types.BankKeeper, rk types.RollappKeeper, warpK *warpkeeper.Keeper) *Keeper {
	return &Keeper{
		bankKeeper: bankKeeper,
		rk:         rk,
		hooks:      nil,
		warpK:      warpK,
	}
}

func (k *Keeper) SetWarpKeeper(warpK *warpkeeper.Keeper) {
	k.warpK = warpK
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

	if err = uevent.EmitTypedEvent(ctx, types.NewEventDenomMetadataCreated(metadata)); err != nil {
		return fmt.Errorf("emit event: %w", err)
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

	if err = uevent.EmitTypedEvent(ctx, types.NewEventDenomMetadataUpdated(metadata)); err != nil {
		return fmt.Errorf("emit event: %w", err)
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
