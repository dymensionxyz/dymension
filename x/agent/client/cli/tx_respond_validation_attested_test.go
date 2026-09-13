package cli

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/stretchr/testify/require"
)

func TestRespondValidationAttestedCommand(t *testing.T) {
	cmd, _, err := GetTxCmd().Find([]string{"respond-validation-attested"})
	require.NoError(t, err)
	require.Equal(t, "respond-validation-attested", cmd.Name())
	args := []string{strings.Repeat("01", 32), "100", "ipfs://verdict", "", "tee", "raw-token"}
	require.NoError(t, cmd.Args(cmd, args))
	require.Error(t, cmd.Args(cmd, args[:5]))
	require.Error(t, cmd.Args(cmd, append(args, "extra")))
	_, _, address := testdata.KeyTestPubAddr()
	msg, err := newMsgRespondValidationAttested(address.String(), args)
	require.NoError(t, err)
	require.Equal(t, address.String(), msg.Responder)
	require.Equal(t, []byte("raw-token"), msg.Token)
	require.Equal(t, uint32(100), msg.Response)
	require.Equal(t, "ipfs://verdict", msg.ResponseUri)
	require.Equal(t, "tee", msg.Tag)
	require.Empty(t, msg.ResponseHash)
	for _, tc := range []struct {
		index int
		value string
	}{{0, "zz"}, {0, "01"}, {1, "-1"}, {1, "4294967296"}, {1, "101"}, {3, "zz"}, {3, "01"}, {5, ""}} {
		bad := append([]string(nil), args...)
		bad[tc.index] = tc.value
		_, err := newMsgRespondValidationAttested(address.String(), bad)
		require.Error(t, err)
	}
}
