package types

import "cosmossdk.io/collections"

const (
	ModuleName = "bridgingfee"
	StoreKey   = ModuleName
)

// Custom hook types to avoid conflicts with upstream Hyperlane
// https://github.com/dymensionxyz/hyperlane-cosmos/blob/ace3bf75a3a2a0e611a18fd47868c0f756915e6a/x/core/02_post_dispatch/types/types.go#L21-L35
const (
	PostDispatchHookDymProtocolFee = iota + 100 // Protocol fee HL hook
	PostDispatchHookDymAggregation              // Aggregation HL hook
)

var (
	KeyFeeHooks         = collections.NewPrefix(1)
	KeyAggregationHooks = collections.NewPrefix(2)
)
