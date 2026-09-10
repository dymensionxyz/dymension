package keeper

import (
	"errors"
	"fmt"
	"testing"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/require"
)

func TestClassifyGaugeLookupError(t *testing.T) {
	missing := fmt.Errorf("gauge with ID 7 does not exist")
	err := classifyGaugeLookupError(7, missing)
	codespace, code, _ := errorsmod.ABCIInfo(err, false)
	require.Equal(t, gerrc.DefaultCodespace, codespace)
	require.Equal(t, uint32(4), code)

	corrupt := errors.New("cannot decode gauge")
	err = classifyGaugeLookupError(7, corrupt)
	require.ErrorIs(t, err, corrupt)
	codespace, code, _ = errorsmod.ABCIInfo(err, false)
	require.Equal(t, errorsmod.UndefinedCodespace, codespace)
	require.Equal(t, uint32(1), code)
}
