package types

import (
	"errors"
)

// DefaultGenesis returns the default Genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:   DefaultParams(),
		Auctions: []Auction{},
	}
}

// Validate performs basic validation of the GenesisState
func (gs GenesisState) Validate() error {
	// Validate auctions
	auctionIDs := make(map[uint64]bool)
	for _, auction := range gs.Auctions {
		// Check for duplicate auction IDs
		if auctionIDs[auction.Id] {
			return errors.New("duplicate auction ID found")
		}
		auctionIDs[auction.Id] = true

		// Validate each auction
		if err := auction.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}
