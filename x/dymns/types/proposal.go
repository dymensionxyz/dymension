package types

import (
	govcdc "github.com/cosmos/cosmos-sdk/x/gov/codec"
	v1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// constants

const (
	ProposalTypeMigrateChainIdsProposal string = "MigrateChainIdsProposal"
	ProposalTypeUpdateAliasesProposal   string = "UpdateAliasesProposal"
)

// Implements Proposal Interface
var (
	_ v1beta1.Content = &MigrateChainIdsProposal{}
	_ v1beta1.Content = &UpdateAliasesProposal{}
)

func init() {
	v1beta1.RegisterProposalType(ProposalTypeMigrateChainIdsProposal)
	govcdc.ModuleCdc.Amino.RegisterConcrete(&MigrateChainIdsProposal{}, "dymns/"+ProposalTypeMigrateChainIdsProposal, nil)

	v1beta1.RegisterProposalType(ProposalTypeUpdateAliasesProposal)
	govcdc.ModuleCdc.Amino.RegisterConcrete(&UpdateAliasesProposal{}, "dymns/"+ProposalTypeUpdateAliasesProposal, nil)
}

// NewMigrateChainIdsProposal returns new instance of MigrateChainIdsProposal
func NewMigrateChainIdsProposal(title, description string, replacement ...MigrateChainId) v1beta1.Content {
	return &MigrateChainIdsProposal{
		Title:       title,
		Description: description,
		Replacement: replacement,
	}
}

// ProposalRoute returns router key for this proposal
func (*MigrateChainIdsProposal) ProposalRoute() string { return RouterKey }

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
func (*UpdateAliasesProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*UpdateAliasesProposal) ProposalType() string {
	return ProposalTypeUpdateAliasesProposal
}
