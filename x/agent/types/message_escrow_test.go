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
			msg := types.NewMsgUpdateAgentSpendPolicy(owner.String(), "a1", "adym", math.NewInt(100), 10, 0)
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

func TestValidateSpendPolicy_SpendExpiryHeight(t *testing.T) {
	_, _, owner := testdata.KeyTestPubAddr()

	tests := []struct {
		name    string
		denom   string
		limit   math.Int
		window  uint64
		expiry  uint64
		wantErr bool
	}{
		{name: "enabled without expiry", denom: "adym", limit: math.NewInt(100), window: 10},
		{name: "enabled with expiry", denom: "adym", limit: math.NewInt(100), window: 10, expiry: 1},
		{name: "enabled with max expiry", denom: "adym", limit: math.NewInt(100), window: 10, expiry: ^uint64(0)},
		{name: "disabled without expiry", limit: math.ZeroInt()},
		{name: "disabled with expiry", limit: math.ZeroInt(), expiry: 1, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgUpdateAgentSpendPolicy(owner.String(), "a1", tc.denom, tc.limit, tc.window, tc.expiry)
			err := msg.ValidateBasic()
			if tc.wantErr {
				require.ErrorContains(t, err, "spend expiry set without spend denom")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
