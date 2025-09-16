package keeper

import (
	"cosmossdk.io/collections"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
)

type Keeper struct {
	feeHooks         collections.Map[uint64, types.HLFeeHook]
	aggregationHooks collections.Map[uint64, types.AggregationHook]
}
