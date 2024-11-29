package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

func TestNewPunishSequencerProposal(t *testing.T) {
	const title = "title"
	const description = "description"
	var seqAddr = sample.AccAddress()
	var rewardee = sample.AccAddress()
	got := NewPunishSequencerProposal(title, description, seqAddr, rewardee)

	require.Equal(t, title, got.GetTitle())
	require.Equal(t, description, got.GetDescription())

	require.Equal(t, RouterKey, got.ProposalRoute())
	require.Equal(t, ProposalTypePunishSequencer, got.ProposalType())
}
