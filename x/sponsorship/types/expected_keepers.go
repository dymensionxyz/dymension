package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
}
