package types

import (
	"errors"
	"fmt"

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
		return dymerrors.Join(gerrc.ErrInvalidArgument, fmt.Errorf("params: %w", err))
	}

	uniqueNames := make(map[string]struct{})
	for _, dymName := range m.DymNames {
		if err := dymName.Validate(); err != nil {
			return dymerrors.Join(gerrc.ErrInvalidArgument, fmt.Errorf("Dym-Name '%s': %w", dymName.Name, err))
		}
		if _, duplicated := uniqueNames[dymName.Name]; duplicated {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "Dym-Name '%s': duplicate name", dymName.Name)
		}
		uniqueNames[dymName.Name] = struct{}{}
	}

	for _, soBid := range m.SellOrderBids {
		soBid.Params = nil // treat it as refund name orders
		if err := soBid.Validate(TypeName); err != nil {
			return dymerrors.Join(gerrc.ErrInvalidArgument, fmt.Errorf("Sell-Order-Bid by '%s': %w", soBid.Bidder, err))
		}
	}

	for _, bo := range m.BuyOrders {
		if err := bo.Validate(); err != nil {
			return dymerrors.Join(gerrc.ErrInvalidArgument, fmt.Errorf("Buy-Order by '%s': %w", bo.Buyer, err))
		}
	}

	if err := validateAliasesOfChainIds(m.AliasesOfRollapps); err != nil {
		return errorsmod.Wrapf(errors.Join(gerrc.ErrInvalidArgument, err), "alias of chain-id")
	}

	return nil
}
