package types_test

import (
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func TestMsgUpdateAgentSpendPolicyValidateBasic_RecipientAllowlist(t *testing.T) {
	_, _, owner := testdata.KeyTestPubAddr()
	_, _, recipient := testdata.KeyTestPubAddr()

	tests := []struct {
		name      string
		allowlist []string
		wantErr   bool
	}{
		{name: "empty is unrestricted"},
		{name: "valid recipient", allowlist: []string{recipient.String()}},
		{name: "invalid recipient", allowlist: []string{"invalid"}, wantErr: true},
		{name: "duplicate recipient", allowlist: []string{recipient.String(), recipient.String()}, wantErr: true},
		{name: "equivalent case recipient", allowlist: []string{recipient.String(), strings.ToUpper(recipient.String())}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgUpdateAgentSpendPolicy(owner.String(), "a1", "adym", math.NewInt(100), 10)
			msg.RecipientAllowlist = tc.allowlist
			err := msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgUpdateAgentSpendPolicyValidateBasic_RecipientCap(t *testing.T) {
	_, _, owner := testdata.KeyTestPubAddr()
	_, _, recipient := testdata.KeyTestPubAddr()

	tests := []struct {
		name      string
		denom     string
		limit     math.Int
		window    uint64
		cap       math.Int
		allowlist []string
		wantErr   bool
	}{
		{name: "unset is disabled", denom: "adym", limit: math.NewInt(100), window: 10, cap: math.Int{}},
		{name: "zero is disabled without allowlist", denom: "adym", limit: math.NewInt(100), window: 10, cap: math.ZeroInt()},
		{name: "cap with allowlist", denom: "adym", limit: math.NewInt(100), window: 10, cap: math.NewInt(40), allowlist: []string{recipient.String()}},
		{name: "cap equal to aggregate", denom: "adym", limit: math.NewInt(100), window: 10, cap: math.NewInt(100), allowlist: []string{recipient.String()}},
		{name: "cap without allowlist", denom: "adym", limit: math.NewInt(100), window: 10, cap: math.NewInt(40), wantErr: true},
		{name: "cap above aggregate", denom: "adym", limit: math.NewInt(100), window: 10, cap: math.NewInt(101), allowlist: []string{recipient.String()}, wantErr: true},
		{name: "negative cap", denom: "adym", limit: math.NewInt(100), window: 10, cap: math.NewInt(-1), allowlist: []string{recipient.String()}, wantErr: true},
		{name: "cap with spending disabled", limit: math.ZeroInt(), cap: math.NewInt(1), allowlist: []string{recipient.String()}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgUpdateAgentSpendPolicy(owner.String(), "a1", tc.denom, tc.limit, tc.window)
			msg.RecipientAllowlist = tc.allowlist
			msg.SpendLimitPerRecipientPerWindow = tc.cap
			err := msg.ValidateBasic()
			if tc.wantErr {
				require.ErrorIs(t, err, types.ErrInvalidSpendPolicy)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
