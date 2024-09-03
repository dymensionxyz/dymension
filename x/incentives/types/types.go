package types

import lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"

type DenomLocksCache map[string][]lockuptypes.PeriodLock

func NewDenomLocksCache() DenomLocksCache {
	return make(DenomLocksCache)
}
