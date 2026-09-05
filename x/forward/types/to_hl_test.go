package types

import (
	"encoding/binary"
	"encoding/json"
	"math/big"
	"testing"

	"cosmossdk.io/math"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	ibcompletiontypes "github.com/dymensionxyz/dymension/v3/x/ibc_completion/types"
	"github.com/stretchr/testify/require"
)

func TestResolveHLForwardAmount(t *testing.T) {
	for _, tc := range []struct {
		name                string
		full                bool
		budget, fee, amount int64
		floor               math.Int
		denom               string
		want                int64
		err                 string
	}{
		{name: "legacy", budget: 100, fee: 10, amount: 80, want: 80},
		{name: "legacy exact budget", budget: 100, fee: 10, amount: 90, want: 90},
		{name: "legacy insufficient", budget: 100, fee: 10, amount: 100, err: "max cost (fee + amount)exceeds max budget 110 > 100"},
		{name: "legacy ignores floor", budget: 100, fee: 10, amount: 80, floor: math.NewInt(99), want: 80},
		{name: "full budget after eibc fee", full: true, budget: 90, fee: 10, amount: 100, want: 80},
		{name: "full budget eliminates surplus", full: true, budget: 100, fee: 10, amount: 20, want: 90},
		{name: "equal fee", full: true, budget: 10, fee: 10, err: "does not cover max fee"},
		{name: "below fee", full: true, budget: 9, fee: 10, err: "does not cover max fee"},
		{name: "zero budget", full: true, err: "does not cover max fee"},
		{name: "zero fee", full: true, budget: 100, want: 100},
		{name: "zero floor", full: true, budget: 100, fee: 10, floor: math.ZeroInt(), want: 90},
		{name: "floor violated", full: true, budget: 100, fee: 10, floor: math.NewInt(91), err: "below min_amount"},
		{name: "floor equal", full: true, budget: 100, fee: 10, floor: math.NewInt(90), want: 90},
		{name: "floor exceeded", full: true, budget: 100, fee: 10, floor: math.NewInt(89), want: 90},
		{name: "legacy wrong denom", budget: 100, fee: 10, denom: "other", err: "max fee denom does not match allowed denom"},
		{name: "full wrong denom", full: true, budget: 100, fee: 10, denom: "other", err: "max fee denom does not match allowed denom"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			denom := tc.denom
			if denom == "" {
				denom = "adym"
			}
			hook := &HookForwardToHL{HyperlaneTransfer: &warptypes.MsgRemoteTransfer{
				Amount: math.NewInt(tc.amount), MaxFee: sdk.NewInt64Coin(denom, tc.fee),
			}, UseFullBudget: tc.full, MinAmount: tc.floor}
			got, err := ResolveHLForwardAmount(sdk.NewInt64Coin("adym", tc.budget), hook)
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, math.NewInt(tc.want), got)
		})
	}
}

// Legacy wire payloads contain only field 1; do not marshal the new hook type to
// construct this fixture, because that would also encode the new floor field.
func TestLegacyHLForwardPayload(t *testing.T) {
	mt := &warptypes.MsgRemoteTransfer{Amount: math.NewInt(80), MaxFee: sdk.NewInt64Coin("adym", 10)}
	transfer, err := mt.Marshal()
	require.NoError(t, err)
	legacy := append([]byte{0x0a}, binary.AppendUvarint(nil, uint64(len(transfer)))...)
	legacy = append(legacy, transfer...)
	hook, err := UnpackForwardToHL(legacy)
	require.NoError(t, err)
	require.False(t, hook.UseFullBudget)
	require.True(t, hook.MinAmount.IsNil() || hook.MinAmount.IsZero())
	amount, err := ResolveHLForwardAmount(sdk.NewInt64Coin("adym", 100), hook)
	require.NoError(t, err)
	before, err := mt.Amount.Marshal()
	require.NoError(t, err)
	after, err := amount.Marshal()
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func TestHLForwardBudgetMemoRoundTrips(t *testing.T) {
	hook := NewHookForwardToHL(hyperutil.HexAddress{}, 1, hyperutil.HexAddress{}, math.NewInt(100), sdk.NewInt64Coin("adym", 10), math.ZeroInt(), nil, "", true, math.NewInt(80))
	for _, path := range []string{"rollapp", "ibc", "hyperlane"} {
		t.Run(path, func(t *testing.T) {
			var payload []byte
			switch path {
			case "rollapp":
				memo, err := MakeRolForwardToHLMemoString("10", hook)
				require.NoError(t, err)
				decoded, err := delayedacktypes.ParseMemo(memo)
				require.NoError(t, err)
				call, err := decoded.EIBC.GetCompletionHook()
				require.NoError(t, err)
				payload = call.Data
			case "ibc":
				memo, err := MakeIBCForwardToHLMemoString(hook)
				require.NoError(t, err)
				var decoded ibcompletiontypes.Memo
				require.NoError(t, json.Unmarshal([]byte(memo), &decoded))
				var call commontypes.CompletionHookCall
				require.NoError(t, proto.Unmarshal(decoded.OnCompletionHook, &call))
				payload = call.Data
			case "hyperlane":
				bz, err := MakeHLForwardToHLMetadata(hook)
				require.NoError(t, err)
				metadata, err := UnpackHLMetadata(bz)
				require.NoError(t, err)
				payload = metadata.HookForwardToHl
			}
			decoded, err := UnpackForwardToHL(payload)
			require.NoError(t, err)
			require.True(t, decoded.UseFullBudget)
			require.Equal(t, math.NewInt(80), decoded.MinAmount)
		})
	}
}

func TestResolveHLForwardAmountRejectsInvalidFee(t *testing.T) {
	for _, fee := range []math.Int{{}, math.NewInt(-1), math.NewIntFromBigInt(new(big.Int).Neg(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))))} {
		hook := &HookForwardToHL{UseFullBudget: true, HyperlaneTransfer: &warptypes.MsgRemoteTransfer{Amount: math.ZeroInt(), MaxFee: sdk.Coin{Denom: "adym", Amount: fee}}}
		require.NotPanics(t, func() {
			_, err := ResolveHLForwardAmount(sdk.NewInt64Coin("adym", 1), hook)
			require.Error(t, err)
		})
	}
}
