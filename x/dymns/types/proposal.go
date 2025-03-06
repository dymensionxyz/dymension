package types

import (
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeMigrateChainIdsProposal defines the type for MigrateChainIdsProposal
	ProposalTypeMigrateChainIdsProposal string = "MigrateChainIdsProposal"
	// ProposalTypeUpdateAliasesProposal defines the type for UpdateAliasesProposal
	ProposalTypeUpdateAliasesProposal string = "UpdateAliasesProposal"
)

// Implements Proposal Interface
var (
	_ v1beta1.Content = &MigrateChainIdsProposal{}
	_ v1beta1.Content = &UpdateAliasesProposal{}
)

// NewMigrateChainIdsProposal returns new instance of MigrateChainIdsProposal
func NewMigrateChainIdsProposal(title, description string, replacement ...MigrateChainId) v1beta1.Content {
	return &MigrateChainIdsProposal{
		Title:       title,
		Description: description,
		Replacement: replacement,
	}
}

// ProposalRoute returns router key for this proposal
func (*MigrateChainIdsProposal) ProposalRoute() string {
	return RouterKey
}

// ProposalType returns proposal type for this proposal
func (*MigrateChainIdsProposal) ProposalType() string {
	return ProposalTypeMigrateChainIdsProposal
}

// NewUpdateAliasesProposal returns new instance of UpdateAliasesProposal
func NewUpdateAliasesProposal(title, description string, add, remove []UpdateAlias) v1beta1.Content {
	return &UpdateAliasesProposal{
		Title:       title,
		Description: description,
		Add:         add,
		Remove:      remove,
	}
}

// ProposalRoute returns router key for this proposal
func (*UpdateAliasesProposal) ProposalRoute() string {
	return RouterKey
}

// ProposalType returns proposal type for this proposal
func (*UpdateAliasesProposal) ProposalType() string {
	return ProposalTypeUpdateAliasesProposal
}
