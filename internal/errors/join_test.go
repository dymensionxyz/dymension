package errors_test

import (
	"errors"
	"regexp"
	"testing"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/require"

	dymerrors "github.com/dymensionxyz/dymension/v3/internal/errors"
)

func TestJoinRendersCauseBeforeRegisteredError(t *testing.T) {
	cause := errorsmod.Wrap(gerrc.ErrNotFound, "missing record")

	err := dymerrors.Join(gerrc.ErrInvalidArgument, cause)

	require.Equal(t, "missing record: not found: invalid argument", err.Error())
	require.NotContains(t, err.Error(), "\n")
	require.ErrorIs(t, err, gerrc.ErrInvalidArgument)
	require.ErrorIs(t, err, gerrc.ErrNotFound)
}

func TestJoinfDoesNotRenderCosmosErrorSourceLocation(t *testing.T) {
	cause := errorsmod.Wrap(gerrc.ErrNotFound, "missing record")

	err := dymerrors.Joinf(gerrc.ErrInvalidArgument, cause, "lookup %q", "name")

	require.Equal(t, "lookup \"name\": missing record: not found: invalid argument", err.Error())
	require.NotRegexp(t, regexp.MustCompile(`\[.*\.go.*\]`), err.Error())
	require.ErrorIs(t, err, cause)
}

func TestJoinNilCauseReturnsNil(t *testing.T) {
	require.NoError(t, dymerrors.Join(gerrc.ErrInvalidArgument, nil))
	require.NoError(t, dymerrors.Joinf(gerrc.ErrInvalidArgument, nil, "context"))
}

func TestJoinResultIsComparableWhenWrapped(t *testing.T) {
	first := dymerrors.Join(gerrc.ErrInvalidArgument, errors.New("first"))
	second := dymerrors.Join(gerrc.ErrInvalidArgument, errors.New("second"))
	wrapped := errorsmod.Wrap(first, "context")

	require.NotPanics(t, func() {
		require.False(t, errors.Is(wrapped, second))
	})
}

func TestJoinPreservesRegisteredABCIInfo(t *testing.T) {
	err := dymerrors.Join(gerrc.ErrInvalidArgument, errors.New("cause"))

	wantCodespace, wantCode, _ := errorsmod.ABCIInfo(gerrc.ErrInvalidArgument, false)
	gotCodespace, gotCode, _ := errorsmod.ABCIInfo(err, false)
	require.Equal(t, wantCodespace, gotCodespace)
	require.Equal(t, wantCode, gotCode)
}
