package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMigrateChainIdsProposal(t *testing.T) {
	const title = "title"
	const description = "description"
	const prevChainId = "cosmoshub-3"
	const newChainId = "cosmoshub-4"
	got := NewMigrateChainIdsProposal(title, description, MigrateChainId{
		PreviousChainId: prevChainId,
		NewChainId:      newChainId,
	})

	require.Equal(t, title, got.GetTitle())
	require.Equal(t, description, got.GetDescription())

	require.Equal(t, RouterKey, got.ProposalRoute())
	require.Equal(t, ProposalTypeMigrateChainIdsProposal, got.ProposalType())
}

func TestNewUpdateAliasesProposal(t *testing.T) {
	const title = "title"
	const description = "description"
	const chainId = "cosmoshub-3"
	const alias = "cosmoshub-4"
	got := NewUpdateAliasesProposal(title, description, []UpdateAlias{
		{
			ChainId: chainId,
			Alias:   alias,
		},
	}, nil)

	require.Equal(t, title, got.GetTitle())
	require.Equal(t, description, got.GetDescription())

	require.Equal(t, RouterKey, got.ProposalRoute())
	require.Equal(t, ProposalTypeUpdateAliasesProposal, got.ProposalType())
}
