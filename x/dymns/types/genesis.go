package types

import (
	"errors"

	dymerrors "github.com/dymensionxyz/dymension/v3/internal/errors"

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
		return dymerrors.Joinf(gerrc.ErrInvalidArgument, err, "params")
	}

	uniqueNames := make(map[string]struct{})
	for _, dymName := range m.DymNames {
		if err := dymName.Validate(); err != nil {
			return dymerrors.Joinf(gerrc.ErrInvalidArgument, err, "Dym-Name '%s'", dymName.Name)
		}
		if _, duplicated := uniqueNames[dymName.Name]; duplicated {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "Dym-Name '%s': duplicate name", dymName.Name)
		}
		uniqueNames[dymName.Name] = struct{}{}
	}

	for _, soBid := range m.SellOrderBids {
		soBid.Params = nil // treat it as refund name orders
		if err := soBid.Validate(TypeName); err != nil {
			return dymerrors.Joinf(gerrc.ErrInvalidArgument, err, "Sell-Order-Bid by '%s'", soBid.Bidder)
		}
	}

	for _, bo := range m.BuyOrders {
		if err := bo.Validate(); err != nil {
			return dymerrors.Joinf(gerrc.ErrInvalidArgument, err, "Buy-Order by '%s'", bo.Buyer)
		}
	}

	if err := validateAliasesOfChainIds(m.AliasesOfRollapps); err != nil {
		return errorsmod.Wrapf(errors.Join(gerrc.ErrInvalidArgument, err), "alias of chain-id")
	}

	return nil
}
