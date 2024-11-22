package uinv

import (
	"errors"
	"testing"

	errorsmod "cosmossdk.io/errors"
	"github.com/stretchr/testify/require"
)

func TestSanityCheckErrorTypes(t *testing.T) {
	baseErr := errors.New("base")
	var nilErr error

	t.Run("breaking", func(t *testing.T) {
		require.True(t, errorsmod.IsOf(Breaking(baseErr), ErrBroken))
		require.False(t, errorsmod.IsOf(Breaking(nilErr), ErrBroken))
	})

	t.Run("join", func(t *testing.T) {
		joinedBase := errors.Join(baseErr, baseErr)
		joinedNil := errors.Join(nil, nil)
		require.True(t, errorsmod.IsOf(Breaking(joinedBase), ErrBroken))
		require.False(t, errorsmod.IsOf(Breaking(joinedNil), ErrBroken))
	})
}
