package cli

import (
	"encoding/json"
	"fmt"
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
	"github.com/spf13/cobra"
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

func TestIBCMinimumFlags(t *testing.T) {
	for _, floor := range []string{"0", "80", "-1", "abc", ""} {
		t.Run(floor, func(t *testing.T) {
			cmd := CmdCreateMemo()
			require.NoError(t, cmd.Flags().Set(FlagDst, DstIBC))
			require.NoError(t, cmd.Flags().Set(FlagMinAmount, floor))
			params, err := parseIBCFlags(cmd)
			if floor == "-1" || floor == "abc" || floor == "" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, floor, params.MinAmount.String())
			_, err = parseHyperlaneFlags(cmd)
			require.NoError(t, err)
		})
	}
}

func TestComposeIBCMinimum(t *testing.T) {
	for _, src := range []string{SrcIBC, SrcEIBC, SrcHL, SrcKaspa} {
		for _, message := range []bool{false, true} {
			if message && (src == SrcIBC || src == SrcEIBC) || !message && src == SrcKaspa {
				continue
			}
			for _, floor := range []string{"", "0", "80"} {
				t.Run(src+"/"+fmt.Sprint(message)+"/"+floor, func(t *testing.T) {
					cmd := CmdCreateMemo()
					args := []string{"--src=" + src, "--dst=ibc", "--channel=channel-0", "--funds-recipient-dst=osmo1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5helwsw"}
					if message {
						cmd = CmdCreateHLMessage()
						args = append(args, "--amount=100", "--token-id=0x0101010101010101010101010101010101010101010101010101010101010101", "--funds-recipient-hub="+sdk.AccAddress(make([]byte, 20)).String())
					} else {
						args = append(args, "--eibc-fee=10")
					}
					if floor != "" {
						args = append(args, "--min-amount="+floor)
					}
					cmd.SetArgs(args)
					output := captureBudgetOutput(t, cmd.Execute)
					var hook *forwardtypes.HookForwardToIBC
					if message {
						bz, err := hyperutil.DecodeEthHex(output)
						require.NoError(t, err)
						decoded, err := parseHL(bz)
						require.NoError(t, err)
						hook = decoded.forwardToIBC
						require.Equal(t, sdk.AccAddress(make([]byte, 20)), decoded.warpPL.GetCosmosAccount())
					} else {
						var data []byte
						switch src {
						case SrcIBC:
							var memo ibcompletiontypes.Memo
							require.NoError(t, json.Unmarshal([]byte(output), &memo))
							var call commontypes.CompletionHookCall
							require.NoError(t, proto.Unmarshal(memo.OnCompletionHook, &call))
							data = call.Data
						case SrcEIBC:
							memo, err := delayedacktypes.ParseMemo(output)
							require.NoError(t, err)
							call, err := memo.EIBC.GetCompletionHook()
							require.NoError(t, err)
							data = call.Data
						case SrcHL:
							bz, err := hyperutil.DecodeEthHex(output)
							require.NoError(t, err)
							metadata, err := forwardtypes.UnpackHLMetadata(bz)
							require.NoError(t, err)
							data = metadata.HookForwardToIbc
						}
						var err error
						hook, err = forwardtypes.UnpackForwardToIBC(data)
						require.NoError(t, err)
					}
					require.NotNil(t, hook)
					require.Equal(t, "osmo1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5helwsw", hook.Transfer.Receiver)
					expected := floor
					if expected == "" {
						expected = "0"
					}
					require.Equal(t, expected, hook.MinAmount.String())
					if message {
						decode := CmdDecodeHL()
						decode.SetArgs([]string{output})
						displayed := captureBudgetOutput(t, decode.Execute)
						require.Contains(t, displayed, "Min Amount:        "+expected)
					}
				})
			}
		}
	}
}

func TestIBCMinimumDisplay(t *testing.T) {
	for _, tc := range []struct {
		name  string
		floor math.Int
		want  string
	}{
		{"absent", math.Int{}, "0"}, {"zero", math.ZeroInt(), "0"}, {"positive", math.NewInt(199), "199"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hook := forwardtypes.NewHookForwardToIBC("channel-0", "destination", 1, tc.floor)
			output := captureBudgetOutput(t, func() error { printForwardToIBC(hook); return nil })
			require.Contains(t, output, "Min Amount:        "+tc.want)
		})
	}
}

func TestIBCMinimumFlagRegistration(t *testing.T) {
	for _, registrars := range [][]func(*cobra.Command){
		{addIBCFlags}, {addIBCFlags, addHyperlaneFlags}, {addHyperlaneFlags, addIBCFlags},
	} {
		cmd := &cobra.Command{}
		for _, register := range registrars {
			register(cmd)
		}
		params, err := parseIBCFlags(cmd)
		require.NoError(t, err)
		require.True(t, params.MinAmount.IsZero())
	}
}

func TestHLToIBCRequiresHubRecipient(t *testing.T) {
	for _, hub := range []string{"", "not-an-address", "osmo1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5helwsw"} {
		t.Run(hub, func(t *testing.T) {
			cmd := CmdCreateHLMessage()
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			cmd.SetArgs([]string{"--src=hl", "--dst=ibc", "--channel=channel-0", "--amount=100", "--token-id=0x0101010101010101010101010101010101010101010101010101010101010101", "--funds-recipient-dst=" + sdk.AccAddress(make([]byte, 20)).String(), "--funds-recipient-hub=" + hub, "--min-amount=80"})
			require.ErrorContains(t, cmd.Execute(), "hub recipient")
		})
	}
}
