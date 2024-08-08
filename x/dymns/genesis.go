package dymns

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k dymnskeeper.Keeper, genState dymnstypes.GenesisState) {
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
	for _, dymName := range genState.DymNames {
		if err := k.SetDymName(ctx, dymName); err != nil {
			panic(err)
		}
		if err := k.AfterDymNameOwnerChanged(ctx, dymName.Name); err != nil {
			panic(err)
		}
		if err := k.AfterDymNameConfigChanged(ctx, dymName.Name); err != nil {
			panic(err)
		}
	}
	for _, bid := range genState.SellOrderBids {
		if err := k.GenesisRefundBid(ctx, bid); err != nil {
			panic(err)
		}
	}
	for _, offer := range genState.BuyOffers {
		if err := k.GenesisRefundOffer(ctx, offer); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k dymnskeeper.Keeper) *dymnstypes.GenesisState {
	// Note: during genesis export, the context does not contain chain-id and time.
	if ctx.BlockTime().Unix() <= 0 {
		// Since the implementation relies on context time, we need to set it to an actual value.
		// The Export-Genesis action supposed to be called by a specific person,
		// on local machine so using time.Now() is fine.
		ctx = ctx.WithBlockTime(time.Now().UTC())
	}

	// Collect Dym-Names records so that we can add back later.
	// We supposed to collect only Non-Expired Dym-Names to save the genesis & store size,
	// but we also need to support those Dym-Names owners which their Dym-Name are expired but within grace period.
	params := k.GetParams(ctx)
	collectExpiredDymNamesExpiredFromEpoch := ctx.BlockTime().Add(-1 * params.Misc.GracePeriodDuration).Unix()

	dymNames := k.GetAllDymNames(ctx)
	var nonExpiredDymNameAndWithinGracePeriod []dymnstypes.DymName
	for _, dymName := range dymNames {
		if dymName.ExpireAt < collectExpiredDymNamesExpiredFromEpoch {
			continue
		}
		nonExpiredDymNameAndWithinGracePeriod = append(nonExpiredDymNameAndWithinGracePeriod, dymName)
	}

	// Collect bidders of active Sell-Orders so that we can refund them later.
	var nonRefundedBids []dymnstypes.SellOrderBid
	for _, bid := range k.GetAllSellOrders(ctx) {
		if bid.HighestBid == nil {
			continue
		}
		// we ignore check expiry here because as long as Sell Orders exists, the highest bid not processed yet.
		nonRefundedBids = append(nonRefundedBids, *bid.HighestBid)
	}

	// Collect buyers of active Buy-Orders so that we can refund them later.
	var nonRefundedBuyOffers []dymnstypes.BuyOffer
	for _, offer := range k.GetAllBuyOffers(ctx) {
		truncatedOffer := offer
		truncatedOffer.CounterpartyOfferPrice = nil
		nonRefundedBuyOffers = append(nonRefundedBuyOffers, truncatedOffer)
	}

	return &dymnstypes.GenesisState{
		Params:        params,
		DymNames:      nonExpiredDymNameAndWithinGracePeriod,
		SellOrderBids: nonRefundedBids,
		BuyOffers:     nonRefundedBuyOffers,
	}
}
