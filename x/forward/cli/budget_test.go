package cli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"cosmossdk.io/math"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	forwardtypes "github.com/dymensionxyz/dymension/v3/x/forward/types"
	ibcompletiontypes "github.com/dymensionxyz/dymension/v3/x/ibc_completion/types"
	"github.com/stretchr/testify/require"
)

func TestHyperlaneBudgetFlags(t *testing.T) {
	for _, tc := range []struct {
		name, floor   string
		full, wantErr bool
	}{
		{"legacy defaults", "0", false, false},
		{"full budget", "80", true, false},
		{"invalid floor", "abc", true, true},
		{"negative floor", "-1", true, true},
		{"inactive positive floor", "80", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := CmdCreateMemo()
			require.NoError(t, cmd.Flags().Set("min-amount", tc.floor))
			if tc.full {
				require.NoError(t, cmd.Flags().Set("use-full-budget", "true"))
			}
			params, err := parseHyperlaneFlags(cmd)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.full, params.UseFullBudget)
			floor, ok := math.NewIntFromString(tc.floor)
			require.True(t, ok)
			require.Equal(t, floor, params.MinAmount)
		})
	}
}

func captureBudgetOutput(t *testing.T, run func() error) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stdout")
	require.NoError(t, err)
	original := os.Stdout
	os.Stdout = f
	t.Cleanup(func() { os.Stdout = original; require.NoError(t, f.Close()) })
	require.NoError(t, run())
	os.Stdout = original
	bz, err := os.ReadFile(f.Name())
	require.NoError(t, err)
	return strings.TrimSpace(string(bz))
}

func TestCreateAndDecodeFullBudgetMemo(t *testing.T) {
	for _, src := range []string{SrcIBC, SrcEIBC, SrcHL} {
		t.Run(src, func(t *testing.T) {
			cmd := CmdCreateMemo()
			cmd.SetArgs([]string{"--src=" + src, "--dst=hl", "--token-id=0x0101010101010101010101010101010101010101010101010101010101010101", "--dst-domain=1", "--funds-recipient-dst=0x0202020202020202020202020202020202020202020202020202020202020202", "--max-fee=10adym", "--use-full-budget", "--min-amount=80", "--eibc-fee=10"})
			output := captureBudgetOutput(t, cmd.Execute)
			var hookData []byte
			switch src {
			case SrcIBC:
				var memo ibcompletiontypes.Memo
				require.NoError(t, json.Unmarshal([]byte(output), &memo))
				var call commontypes.CompletionHookCall
				require.NoError(t, proto.Unmarshal(memo.OnCompletionHook, &call))
				hookData = call.Data
			case SrcEIBC:
				memo, err := delayedacktypes.ParseMemo(output)
				require.NoError(t, err)
				call, err := memo.EIBC.GetCompletionHook()
				require.NoError(t, err)
				hookData = call.Data
			case SrcHL:
				bz, err := hyperutil.DecodeEthHex(output)
				require.NoError(t, err)
				metadata, err := forwardtypes.UnpackHLMetadata(bz)
				require.NoError(t, err)
				hookData = metadata.HookForwardToHl
			}
			hook, err := forwardtypes.UnpackForwardToHL(hookData)
			require.NoError(t, err)
			require.True(t, hook.UseFullBudget)
			require.Equal(t, math.NewInt(80), hook.MinAmount)
			message, err := MakeForwardToHLHyperlaneMessage(1, 1, hyperutil.HexAddress{}, 2, hyperutil.HexAddress{}, sdk.AccAddress(make([]byte, 20)), math.NewInt(90), hook)
			require.NoError(t, err)
			decode := CmdDecodeHL()
			decode.SetArgs([]string{hyperutil.EncodeEthHex(message.Bytes())})
			displayed := captureBudgetOutput(t, decode.Execute)
			require.Contains(t, displayed, "Use Full Budget:    true")
			require.Contains(t, displayed, "Min Amount:         80")
		})
	}
}

func TestLegacyBudgetFloorDisplay(t *testing.T) {
	hook := forwardtypes.NewHookForwardToHL(hyperutil.HexAddress{}, 1, hyperutil.HexAddress{}, math.NewInt(100), sdk.NewInt64Coin("adym", 10), math.ZeroInt(), nil, "", false, math.NewInt(80))
	output := captureBudgetOutput(t, func() error { printForwardToHL(hook); return nil })
	require.Contains(t, output, "Min Amount:         80 (inactive; requires use_full_budget)")
}
