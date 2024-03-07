package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper provides a way to manage denommetadata module storage.
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

// Logger returns a logger instance for the denommetadata module.
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
	err = k.SetDenomMetadataWithRefKey(ctx, &denomMetadata)
	if err != nil {
		return 0, err
	}
	k.SetLastDenomMetadataID(ctx, denomMetadata.Id)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCreateDenomMetadata,
			denomMetadata.TokenMetadata.GetEvents(denomMetadata.Id)...,
		),
	})

	return denomMetadata.Id, nil
}

// CheckExistingMetadata checks if there is any param collision with existing metadata recors and returns erros in case it happens
func (k Keeper) CheckExistingMetadata(ctx sdk.Context, record types.TokenMetadata) error {

	store := ctx.KVStore(k.storeKey)
	denomMetadataBaseKey := denomMetadataStoreBaseKey(record.Base)
	if store.Has(denomMetadataBaseKey) {
		return fmt.Errorf("DenomMetadata with base %s already exists", record.Base)
	}
	denomMetadataDisplayKey := denomMetadataStoreDisplayKey(record.Display)
	if store.Has(denomMetadataDisplayKey) {
		return fmt.Errorf("DenomMetadata with display %s already exists", record.Display)
	}
	denomMetadataSymbolKey := denomMetadataStoreSymbolKey(record.Symbol)
	if store.Has(denomMetadataSymbolKey) {
		return fmt.Errorf("DenomMetadata with base %s already exists", record.Symbol)
	}
	return nil
}

// UpdateDenomMetadata modifies an existing denom metadata record.
func (k Keeper) UpdateDenomMetadata(ctx sdk.Context, denomMetadataID uint64, record types.TokenMetadata) error {

	err := record.Validate()
	if err != nil {
		return err
	}
	denomMetadata, err := k.GetDenomMetadataByID(ctx, denomMetadataID)
	if err != nil {
		return err
	}

	denomMetadata.TokenMetadata = record

	err = k.SetDenomMetadataWithRefKey(ctx, denomMetadata)
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtUpdateDenomMetadata,
			denomMetadata.TokenMetadata.GetEvents(denomMetadataID)...,
		),
	})
	return nil
}
