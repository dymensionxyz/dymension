package cli

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestRequestValidationCommandBounty(t *testing.T) {
	cmd, _, err := GetTxCmd().Find([]string{"request-validation"})
	require.NoError(t, err)
	require.NotNil(t, cmd.Flags().Lookup(flagBounty))
	args := []string{"validator", "subject", "1", strings.Repeat("01", 32), "ipfs://request"}

	msg, err := newMsgRequestValidation("dym1requester", args, "")
	require.NoError(t, err)
	require.True(t, msg.Bounty.IsNil())

	msg, err = newMsgRequestValidation("dym1requester", args, "1000adym")
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin("adym", 1000), msg.Bounty)
	require.Equal(t, uint64(1), msg.EvidenceSeq)

	_, err = newMsgRequestValidation("dym1requester", args, "lots")
	require.Error(t, err)
}

func TestReclaimValidationBountyCommand(t *testing.T) {
	cmd, _, err := GetTxCmd().Find([]string{"reclaim-validation-bounty"})
	require.NoError(t, err)
	require.Equal(t, "reclaim-validation-bounty", cmd.Name())
	require.NoError(t, cmd.Args(cmd, []string{strings.Repeat("01", 32)}))
	require.Error(t, cmd.Args(cmd, nil))
}
