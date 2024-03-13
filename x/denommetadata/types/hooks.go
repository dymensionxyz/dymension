package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// DenomMetadataHooks event hooks for denom metadata creation/update
type DenomMetadataHooks interface {
	AfterDenomMetadataCreation(ctx sdk.Context, metadata banktypes.Metadata) error
	AfterDenomMetadataUpdate(ctx sdk.Context, metadata banktypes.Metadata) error
}

var _ DenomMetadataHooks = MultiDenomMetadataHooks{}

// combine multiple DenomMetadata hooks, all hook functions are run in array sequence
type MultiDenomMetadataHooks []DenomMetadataHooks

// Creates hooks for the DenomMetadata Module.
func NewMultiRollappHooks(hooks ...DenomMetadataHooks) MultiDenomMetadataHooks {
	return hooks
}

func (h MultiDenomMetadataHooks) AfterDenomMetadataCreation(ctx sdk.Context, metadata banktypes.Metadata) error {
	for i := range h {
		err := h[i].AfterDenomMetadataCreation(ctx, metadata)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h MultiDenomMetadataHooks) AfterDenomMetadataUpdate(ctx sdk.Context, metadata banktypes.Metadata) error {
	for i := range h {
		err := h[i].AfterDenomMetadataUpdate(ctx, metadata)
		if err != nil {
			return err
		}
	}
	return nil
}
