package types

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate checks if the GenesisState is valid.
func (m GenesisState) Validate() error {
	if err := (&m.Params).Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "params: %v", err)
	}

	for _, dymName := range m.DymNames {
		if err := dymName.Validate(); err != nil {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "Dym-Name '%s': %v", dymName.Name, err)
		}
	}

	for _, soBid := range m.SellOrderBids {
		if err := soBid.Validate(); err != nil {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "Sell-Order-Bid by '%s': %v", soBid.Bidder, err)
		}
	}

	for _, bo := range m.BuyOffers {
		if err := bo.Validate(); err != nil {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "Buy-Offer by '%s': %v", bo.Buyer, err)
		}
	}

	return nil
}
