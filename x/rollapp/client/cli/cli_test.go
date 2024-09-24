package cli_test

import (
	"os"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/client/cli"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var (
	metadata = `
	{
		"website": "https://dymension.xyz/",
		"description": "This is a description of the Rollapp.",
		"logo_data_uri": "data:image/jpeg;base64,/000",
		"token_logo_uri": "data:image/jpeg;base64,/000",
		"telegram": "https://t.me/example",
		"x": "https://x.com/dymension"
	}
	`

	nativeDenom = `
	{
		"display": "dummyDisplay",
		"base": "dummyBase",
		"exponent": 10
	}
	`
)

func TestGetTxCmd(t *testing.T) {
	cmd := cli.GetTxCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, types.ModuleName, cmd.Use)
	assert.True(t, cmd.HasSubCommands())

	cmd = cli.CmdCreateRollapp()
	assert.NotNil(t, cmd)
	assert.True(t, strings.HasPrefix(cmd.Use, "create"))
	assert.True(t, cmd.Flags().HasFlags())
}

func TestCmdCreateIRO(t *testing.T) {
	addr := sdk.AccAddress("testAddress").String()

	// Create a temporary file for metadata
	metadataFile, err := os.CreateTemp("", "metadata*.json")
	assert.NoError(t, err)
	defer os.Remove(metadataFile.Name()) // nolint:errcheck
	_, err = metadataFile.WriteString(metadata)
	assert.NoError(t, err)

	// Create a temporary file for native denom
	nativeDenomFile, err := os.CreateTemp("", "nativeDenom*.json")
	assert.NoError(t, err)
	defer os.Remove(nativeDenomFile.Name()) // nolint:errcheck
	_, err = nativeDenomFile.WriteString(nativeDenom)
	assert.NoError(t, err)

	testCases := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{
			"valid minimal args",
			[]string{"testRollappId", "alias", "EVM", "--from", addr},
			"",
		},
		{
			"with metadata",
			[]string{"testRollappId", "alias", "EVM", "--metadata", metadataFile.Name(), "--from", addr},
			"",
		},
		{
			"with native denom",
			[]string{"testRollappId", "alias", "EVM", "--native-denom", nativeDenomFile.Name(), "--from", addr},
			"",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.CmdCreateRollapp()
			cmd.SetArgs(tc.args)
			err := cmd.Execute()
			if tc.errMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.Contains(t, err.Error(), "key not found") // we expect this error because we are not setting the key. anyway it means we passed validation
			}
		})
	}
}
