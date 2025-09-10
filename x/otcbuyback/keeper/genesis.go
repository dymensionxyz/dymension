package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

// InitGenesis initializes the otcbuyback module's state from a provided genesis state
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// Set module parameters
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	lastAuctionID := uint64(0)
	expectedFunds := math.ZeroInt()
	// Set auctions from genesis state
	for _, auction := range genState.Auctions {

		// validate the auction is not active
		if auction.IsActive(ctx.BlockTime()) {
			panic(fmt.Sprintf("auction %d is active", auction.Id))
		}

		if auction.IsUpcoming(ctx.BlockTime()) {
			if auction.SoldAmount.IsPositive() {
				panic(fmt.Sprintf("upcoming auction %d has sold tokens", auction.Id))
			}
			expectedFunds = expectedFunds.Add(auction.GetRemainingAllocation())
		}

		if err := k.SetAuction(ctx, auction); err != nil {
			panic(err)
		}

		if auction.Id > lastAuctionID {
			lastAuctionID = auction.Id
		}

	}

	// validate the funds available in the module account
	if k.bankKeeper.GetBalance(ctx, k.accountKeeper.GetModuleAddress(types.ModuleName), k.baseDenom).Amount.LT(expectedFunds) {
		panic(fmt.Sprintf("module account %s has less funds than expected", types.ModuleName))
	}

	err := k.SetNextAuctionID(ctx, lastAuctionID+1)
	if err != nil {
		panic(err)
	}
}

// SetNextAuctionID sets the next auction ID using collections
func (k Keeper) SetNextAuctionID(ctx sdk.Context, id uint64) error {
	return k.nextAuctionID.Set(ctx, id)
}

// ExportGenesis returns the otcbuyback module's exported genesis state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := &types.GenesisState{}

	// Export parameters
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	genesis.Params = params

	// Export all accepted tokens
	acceptedTokens, err := k.GetAllAcceptedTokens(ctx)
	if err != nil {
		panic(err)
	}
	genesis.AcceptedTokens = acceptedTokens

	// Export all auctions
	auctions, err := k.GetAllAuctions(ctx)
	if err != nil {
		panic(err)
	}
	genesis.Auctions = auctions

	// FIXME: Export all purchases

	return genesis
}
