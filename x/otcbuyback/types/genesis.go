package types

import (
	"errors"
)

// DefaultGenesis returns the default Genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:         DefaultParams(),
		AcceptedTokens: []AcceptedToken{},
		Auctions:       []Auction{},
	}
}

// Validate performs basic validation of the GenesisState
func (gs GenesisState) ValidateBasic() error {

	// validate params
	if err := gs.Params.ValidateBasic(); err != nil {
		return err
	}

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

	// check for duplicate accepted tokens
	acceptedTokens := make(map[string]bool)
	for _, acceptedToken := range gs.AcceptedTokens {
		if acceptedTokens[acceptedToken.Denom] {
			return errors.New("duplicate accepted token found")
		}
		acceptedTokens[acceptedToken.Denom] = true
	}
	// accepted tokens validated on InitGenesis

	return nil
}
