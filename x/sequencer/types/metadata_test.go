package types

import (
	"net/url"
	"testing"

	errorsmod "cosmossdk.io/errors"
	"github.com/stretchr/testify/require"
)

func TestValidateURLsPreservesErrorChain(t *testing.T) {
	err := validateURLs([]string{"https://example.com/%"})

	require.ErrorIs(t, err, ErrInvalidURL)
	var urlErr *url.Error
	require.ErrorAs(t, err, &urlErr)

	wantCodespace, wantCode, _ := errorsmod.ABCIInfo(ErrInvalidURL, false)
	gotCodespace, gotCode, _ := errorsmod.ABCIInfo(err, false)
	require.Equal(t, wantCodespace, gotCodespace)
	require.Equal(t, wantCode, gotCode)
}
