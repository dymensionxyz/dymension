package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func TestParams_Validate(t *testing.T) {
	require.NoError(t, types.DefaultParams().Validate())

	p := types.DefaultParams()
	p.PolicyRotationDelayBlocks = 0
	require.Error(t, p.Validate())
}
