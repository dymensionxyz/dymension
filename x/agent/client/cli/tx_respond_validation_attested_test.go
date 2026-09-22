package cli

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	clitest "github.com/cosmos/cosmos-sdk/testutil/cli"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"

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
	require.Equal(t, "raw-token", msg.Token)
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

func TestRespondValidationAttestedGenerateOnly(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)
	ctx := client.Context{}.WithCodec(cdc).WithTxConfig(authtx.NewTxConfig(cdc, authtx.DefaultSignModes))
	_, _, address := testdata.KeyTestPubAddr()
	args := []string{strings.Repeat("01", 32), "100", "ipfs://verdict", "", "tee", "header.payload.signature", "--from", address.String(), "--generate-only", "--offline", "--account-number", "0", "--sequence", "0"}
	out, err := clitest.ExecTestCLICmd(ctx, CmdRespondValidationAttested(), args)
	require.NoError(t, err)
	t.Logf("generated transaction: %s", out.String())
	decoded, err := ctx.TxConfig.TxJSONDecoder()(out.Bytes())
	require.NoError(t, err)
	msg, ok := decoded.GetMsgs()[0].(*types.MsgRespondValidationAttested)
	require.True(t, ok)
	require.Equal(t, "header.payload.signature", msg.Token)
	require.Contains(t, out.String(), `"token":"header.payload.signature"`)
	args[1] = "101"
	_, err = clitest.ExecTestCLICmd(ctx, CmdRespondValidationAttested(), args)
	require.ErrorIs(t, err, types.ErrInvalidValidationResponse)
	t.Logf("invalid verdict rejected: %s", err)
}
