package cli_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/cli"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTxCmd(t *testing.T) {
	cmd := cli.GetTxCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, types.ModuleName, cmd.Use)
	assert.True(t, cmd.HasSubCommands())

	cmd = cli.CmdCreateIRO()
	assert.NotNil(t, cmd)
	assert.True(t, strings.HasPrefix(cmd.Use, "create"))
	assert.True(t, cmd.Flags().HasFlags())

	cmd = cli.CmdBuy()
	assert.NotNil(t, cmd)
	assert.True(t, strings.HasPrefix(cmd.Use, "buy"))

	cmd = cli.CmdSell()
	assert.NotNil(t, cmd)
	assert.True(t, strings.HasPrefix(cmd.Use, "sell"))
}

func TestCmdCreateIRO(t *testing.T) {
	addr := sdk.AccAddress("testAddress").String()

	testCases := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{
			"valid args",
			[]string{"testRollappId", "1000000", "2006-01-02T15:04:05Z", "--curve", "1.2,0.4,0", "--from", addr},
			"",
		},
		{
			"valid args with incentives",
			[]string{"testRollappId", "1000000", "2006-01-02T15:04:05Z", "--curve", "1.2,0.4,0", "--from", addr, "--" + cli.FlagIncentivesEpochs, "10", "--" + cli.FlagIncentivesStartDurationAfterSettlement, "1h"},
			"",
		},
		{
			"missing rollappId",
			[]string{"1000000", "2006-01-02T15:04:05Z", "--curve", "1.2,0.4,0", "--from", addr},
			"accepts 3 arg",
		},
		{
			"missing allocation",
			[]string{"testRollappId", "1630000000", "--curve", "1.2,0.4,0", "--from", addr},
			"accepts 3 arg",
		},
		{
			"missing curve",
			[]string{"testRollappId", "1000000", "1630000000", "--from", addr},
			"curve",
		},
		{
			"invalid allocation",
			[]string{"testRollappId", "invalid", "1630000000", "--curve", "1.2,0.4,0", "--from", addr},
			"allocation amount",
		},
		{
			"invalid pre-launch time",
			[]string{"testRollappId", "1000000", "invalid", "--curve", "1.2,0.4,0", "--from", addr},
			"start time",
		},
		{
			"invalid curve",
			[]string{"testRollappId", "1000000", "1630000000", "--curve", "s,s,s", "--from", addr},
			"curve",
		},
		{
			"invalid incentives params - start",
			[]string{"testRollappId", "1000000", "1630000000", "--curve", "1.2,0.4,0", "--incentives-start", "invalid", "--from", addr},
			"incentives-start",
		},
		{
			"invalid incentives params - epochs",
			[]string{"testRollappId", "1000000", "1630000000", "--curve", "1.2,0.4,0", "--incentives-epochs", "-1", "--from", addr},
			"incentives-epochs",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.CmdCreateIRO()
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

func TestParseBondingCurve(t *testing.T) {
	tests := []struct {
		name     string
		curveStr string
		wantErr  bool
	}{
		{"Valid curve", "1.2,0.4,0", false},
		{"Valid curve Ints", "2,1,1", false},
		{"Invalid params count", "1.2,0.4", true},
		{"Invalid params count - too much", "1.2,0.4,0.1,0.2", true},
		{"Invalid M", "invalid,0.4,0", true},
		{"Invalid N", "1.2,invalid,0", true},
		{"Invalid C", "1.2,0.4,invalid", true},
		{"Negative values M", "-1.2,0.4,0", true},
		{"Negative values N", "1.2,-0.4,0", true},
		{"Negative values C", "1.2,0.4,-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			curve, err := cli.ParseBondingCurve(tt.curveStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, curve)
			}
		})
	}
}

func TestFlagSetCreatePlan(t *testing.T) {
	flagSet := cli.FlagSetCreatePlan()
	assert.NotNil(t, flagSet)

	assert.True(t, flagSet.HasAvailableFlags())
	assert.NotNil(t, flagSet.Lookup(cli.FlagBondingCurve))
	assert.NotNil(t, flagSet.Lookup(cli.FlagStartTime))
	assert.NotNil(t, flagSet.Lookup(cli.FlagIncentivesStartDurationAfterSettlement))
	assert.NotNil(t, flagSet.Lookup(cli.FlagIncentivesEpochs))
}
