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

func TestMsgUpdateAgentSpendPolicyValidateBasic_MaxPerTransfer(t *testing.T) {
	_, _, owner := testdata.KeyTestPubAddr()

	tests := []struct {
		name           string
		spendDenom     string
		maxPerTransfer math.Int
		wantErr        bool
	}{
		{name: "unset", spendDenom: "adym"},
		{name: "zero", spendDenom: "adym", maxPerTransfer: math.ZeroInt()},
		{name: "below window limit", spendDenom: "adym", maxPerTransfer: math.NewInt(10)},
		{name: "above window limit never binds", spendDenom: "adym", maxPerTransfer: math.NewInt(1000)},
		{name: "negative", spendDenom: "adym", maxPerTransfer: math.NewInt(-1), wantErr: true},
		{name: "set without spend denom", maxPerTransfer: math.NewInt(10), wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgUpdateAgentSpendPolicy(owner.String(), "a1", "", math.Int{}, 0)
			if tc.spendDenom != "" {
				msg = types.NewMsgUpdateAgentSpendPolicy(owner.String(), "a1", tc.spendDenom, math.NewInt(100), 10)
			}
			msg.SpendMaxPerTransfer = tc.maxPerTransfer
			err := msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
