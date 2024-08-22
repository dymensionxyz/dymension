package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

// ValidateBasic performs basic validation for the MigrateChainIdsProposal.
func (m *MigrateChainIdsProposal) ValidateBasic() error {
	if len(m.Replacement) == 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "replacement cannot be empty")
	}

	uniqueChainIds := make(map[string]bool)
	// Describe usage of Go Map: only used for validation
	for _, r := range m.Replacement {
		if err := r.ValidateBasic(); err != nil {
			return err
		}

		normalizedPreviousChainId := strings.ToLower(r.PreviousChainId)
		normalizedNewChainId := strings.ToLower(r.NewChainId)

		if _, found := uniqueChainIds[normalizedPreviousChainId]; found {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "duplicate chain id: %s", r.PreviousChainId)
		}
		uniqueChainIds[normalizedPreviousChainId] = true

		if _, found := uniqueChainIds[normalizedNewChainId]; found {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "duplicate chain id: %s", r.NewChainId)
		}
		uniqueChainIds[normalizedNewChainId] = true
	}

	return v1beta1.ValidateAbstract(m)
}

// ValidateBasic performs basic validation for the MigrateChainId operation.
func (m MigrateChainId) ValidateBasic() error {
	if m.PreviousChainId == "" {
		return govtypes.ErrInvalidProposalContent.Wrap("previous chain id cannot be empty")
	}
	if !dymnsutils.IsValidChainIdFormat(m.PreviousChainId) {
		return govtypes.ErrInvalidProposalContent.Wrapf("previous chain id is not well-formed: %s", m.PreviousChainId)
	}

	if m.NewChainId == "" {
		return govtypes.ErrInvalidProposalContent.Wrap("new chain id cannot be empty")
	}
	if !dymnsutils.IsValidChainIdFormat(m.NewChainId) {
		return govtypes.ErrInvalidProposalContent.Wrapf("new chain id is not well-formed: %s", m.NewChainId)
	}

	if strings.EqualFold(m.PreviousChainId, m.NewChainId) {
		return govtypes.ErrInvalidProposalContent.Wrap("previous chain id and new chain id cannot be the same")
	}

	return nil
}
