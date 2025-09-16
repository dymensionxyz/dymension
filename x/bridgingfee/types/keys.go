package types

import "cosmossdk.io/collections"

const (
	ModuleName = "bridgingfee"
	StoreKey   = ModuleName
)

var (
	KeyFeeHooks         = collections.NewPrefix(1)
	KeyAggregationHooks = collections.NewPrefix(2)
)
