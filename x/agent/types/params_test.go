package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func TestParams_Validate(t *testing.T) {
	require.NoError(t, types.DefaultParams().Validate())
	require.Equal(t, uint64(50), types.DefaultParams().SpendRecipientAllowlistMax)
	require.Equal(t, uint64(8), types.DefaultParams().ValidationMaxResponsesPerRequest)
	require.Equal(t, uint64(100_800), types.DefaultParams().ValidationBountyExpiryBlocks)

	p := types.DefaultParams()
	p.PolicyRotationDelayBlocks = 0
	require.Error(t, p.Validate())

	p = types.DefaultParams()
	p.SpendRecipientAllowlistMax = 0
	require.Error(t, p.Validate())

	p = types.DefaultParams()
	p.ValidationMaxResponsesPerRequest = 0
	require.NoError(t, p.Validate())

	p = types.DefaultParams()
	p.ValidationBountyExpiryBlocks = 0
	require.ErrorContains(t, p.Validate(), "validation bounty expiry blocks must be positive")
}
