package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	_ paramtypes.ParamSet = (*Params)(nil)
)

// Params defines the parameters for the hyperlane module
type Params struct {
	// TODO: Add parameters as needed
}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements paramtypes.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		// TODO: Add parameter pairs as needed
	}
}

// DefaultParams returns default hyperlane module parameters
func DefaultParams() *Params {
	return &Params{
		// TODO: Set default values
	}
}

// HyperlaneHooks defines the interface for hyperlane hooks
type HyperlaneHooks interface {
	// AfterMessageProcessed is called after a hyperlane message is processed
	AfterMessageProcessed(ctx sdk.Context, messageID string) error
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (h HyperlaneHooks) IsOnePerModuleType() {}
