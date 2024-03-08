package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper provides a way to manage denommetadata module storage.
type Keeper struct {
	bankkeeper types.BankKeeper
}

// NewKeeper returns a new instance of the incentive module keeper struct.
func NewKeeper(bankkeeper types.BankKeeper) *Keeper {
	return &Keeper{
		bankkeeper: bankkeeper,
	}
}

// Logger returns a logger instance for the denommetadata module.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// CreateDenomMetadata creates a new denom metadata record.
func (k Keeper) CreateDenomMetadata(ctx sdk.Context, record types.TokenMetadata) error {

	found := k.bankkeeper.HasDenomMetaData(ctx, record.Base)
	if found {
		return fmt.Errorf("Existing denom")
	}

	k.bankkeeper.SetDenomMetaData(ctx, record.ConvertToBankMetadata())
	return nil
}

// UpdateDenomMetadata modifies an existing denom metadata record.
func (k Keeper) UpdateDenomMetadata(ctx sdk.Context, record types.TokenMetadata) error {

	found := k.bankkeeper.HasDenomMetaData(ctx, record.Base)
	if !found {
		return fmt.Errorf("Denom does not exist")
	}

	k.bankkeeper.SetDenomMetaData(ctx, record.ConvertToBankMetadata())
	return nil
}
