package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

type DenomLocksCache map[string][]lockuptypes.PeriodLock

func NewDenomLocksCache() DenomLocksCache {
	return make(DenomLocksCache)
}

func (g Gauge) AddCoins(coins sdk.Coins) Gauge {
	g.Coins = g.Coins.Add(coins...)
	return g
}

func (g Gauge) Key() uint64 {
	return g.Id
}
