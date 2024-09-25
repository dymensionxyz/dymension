package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// DenomLocksCache is a cache that persists the list of lockups per denom, so
// a string key is a denom name.
type DenomLocksCache map[string][]lockuptypes.PeriodLock

func NewDenomLocksCache() DenomLocksCache {
	return make(DenomLocksCache)
}

func (g *Gauge) AddCoins(coins sdk.Coins) {
	g.Coins = g.Coins.Add(coins...)
}

func (g Gauge) Key() uint64 {
	return g.Id
}
