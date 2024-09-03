package types

import lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

type DenomLocksCache map[string][]lockuptypes.PeriodLock

func NewDenomLocksCache() DenomLocksCache {
	return make(DenomLocksCache)
}
