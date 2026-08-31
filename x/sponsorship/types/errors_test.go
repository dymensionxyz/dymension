package types_test

import (
	"errors"
	"testing"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func TestErrorsHaveDistinctIdentitiesAndGRPCCategories(t *testing.T) {
	invalidErrors := []error{
		types.ErrInvalidGaugeWeight,
		types.ErrInvalidDistribution,
		types.ErrInvalidParams,
		types.ErrInvalidGenesis,
		types.ErrInvalidVote,
		types.ErrInvalidVoterInfo,
	}

	for i, err := range invalidErrors {
		codespace, code, _ := errorsmod.ABCIInfo(err, false)
		require.Equal(t, types.ModuleName, codespace)
		require.Equal(t, uint32(i+1), code)
		require.Equal(t, codes.InvalidArgument, status.Code(err))

		for j, other := range invalidErrors {
			if i != j {
				require.False(t, errors.Is(err, other), "%v must not match %v", err, other)
			}
		}
	}

	codespace, code, _ := errorsmod.ABCIInfo(types.ErrNoEndorsers, false)
	require.Equal(t, types.ModuleName, codespace)
	require.Equal(t, uint32(7), code)
	require.Equal(t, codes.FailedPrecondition, status.Code(types.ErrNoEndorsers))
	require.False(t, errors.Is(gerrc.ErrFailedPrecondition.Wrap("anything"), types.ErrNoEndorsers))
}
