package types

import (
	"errors"

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
		soBid.Params = nil // treat it as refund name orders
		if err := soBid.Validate(TypeName); err != nil {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "Sell-Order-Bid by '%s': %v", soBid.Bidder, err)
		}
	}

	for _, bo := range m.BuyOrders {
		if err := bo.Validate(); err != nil {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "Buy-Order by '%s': %v", bo.Buyer, err)
		}
	}

	if err := validateAliasesOfChainIds(m.AliasesOfRollapps); err != nil {
		return errorsmod.Wrapf(errors.Join(gerrc.ErrInvalidArgument, err), "alias of chain-id")
	}

	return nil
}
