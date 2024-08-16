package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

// ValidateBasic performs basic validation for the UpdateAliasesProposal.
func (m *UpdateAliasesProposal) ValidateBasic() error {
	if len(m.Add) == 0 && len(m.Remove) == 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "update list can not be empty")
	}

	uniquePairs := make(map[string]bool)
	// Describe usage of Go Map: only used for validation

	for _, r := range append(m.Add, m.Remove...) {
		if err := r.ValidateBasic(); err != nil {
			return err
		}

		pairId := fmt.Sprintf("%s|%s", r.ChainId, r.Alias)
		if _, found := uniquePairs[pairId]; found {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "duplicate chain id and alias pair: %s", pairId)
		}
		uniquePairs[pairId] = true
	}

	return v1beta1.ValidateAbstract(m)
}

// ValidateBasic performs basic validation for the UpdateAlias operation.
func (m *UpdateAlias) ValidateBasic() error {
	if m.ChainId == "" {
		return govtypes.ErrInvalidProposalContent.Wrap("chain id cannot be empty")
	}

	if !dymnsutils.IsValidChainIdFormat(m.ChainId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "chain id is not well-formed: %s", m.ChainId)
	}

	if m.Alias == "" {
		return govtypes.ErrInvalidProposalContent.Wrap("alias cannot be empty")
	}

	if !dymnsutils.IsValidAlias(m.Alias) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "alias is not well-formed: %s", m.Alias)
	}

	if m.ChainId == m.Alias {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "chain id and alias cannot be the same")
	}

	return nil
}
