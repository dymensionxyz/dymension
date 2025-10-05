package keeper_test

import (
	"testing"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.IROKeeper(t)
	params := types.DefaultParams()
	params.CreationFee = math.NewInt(100).MulRaw(1e18)
	k.SetParams(ctx, params)
	require.EqualValues(t, params, k.GetParams(ctx))
}

func (s *KeeperTestSuite) TestErrorString() {
	err := errorsmod.Wrapf(types.ErrPlanExists, "wrapped error")

	// Incorrect version! %v turns err into Stringer that prints the filepath
	wrappedErr1 := errorsmod.Wrapf(types.ErrPlanExists, "format without .Error() %v", err)

	// Correct version! %v turns err.Error() into Stringer as well, but .Error() hides the filepath
	wrappedErr2 := errorsmod.Wrapf(types.ErrPlanExists, "format with .Error() %v", err.Error())

	// Finds [] brackets with .go between them
	const regexp = `\[.*\.go.*\]`

	// Filepath appears
	s.Require().Regexp(regexp, wrappedErr1)
	// Filepath appears
	s.Require().Regexp(regexp, wrappedErr1.Error())
	// Filepath appears
	s.Require().Regexp(regexp, wrappedErr2)

	// Filepath doesn't appear
	s.Require().NotRegexp(regexp, wrappedErr2.Error())
}
