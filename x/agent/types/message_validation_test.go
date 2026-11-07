package types_test

import (
	"bytes"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func TestValidationMessagesValidateBasic(t *testing.T) {
	_, _, account := testdata.KeyTestPubAddr()
	addr := account.String()
	hash := bytes.Repeat([]byte{1}, 32)
	require.NoError(t, (&types.MsgRequestValidation{Requester: addr, ValidatorId: "validator", AgentId: "subject", RequestHash: hash}).ValidateBasic())
	require.ErrorIs(t, (&types.MsgRequestValidation{Requester: addr, ValidatorId: "validator", AgentId: "subject", RequestHash: hash[:31]}).ValidateBasic(), types.ErrInvalidValidationHash)
	require.NoError(t, (&types.MsgRespondValidation{Responder: addr, RequestHash: hash, Response: 100}).ValidateBasic())
	require.ErrorIs(t, (&types.MsgRespondValidation{Responder: addr, RequestHash: hash, Response: 101}).ValidateBasic(), types.ErrInvalidValidationResponse)
	require.ErrorIs(t, (&types.MsgRespondValidation{Responder: addr, RequestHash: hash, ResponseHash: hash[:31]}).ValidateBasic(), types.ErrInvalidValidationHash)
}

func TestRequestValidationBountyValidateBasic(t *testing.T) {
	_, _, account := testdata.KeyTestPubAddr()
	hash := bytes.Repeat([]byte{1}, 32)
	for name, tc := range map[string]struct {
		bounty sdk.Coin
		valid  bool
	}{
		"unset":     {sdk.Coin{}, true},
		"zero":      {sdk.NewInt64Coin("adym", 0), true},
		"positive":  {sdk.NewInt64Coin("adym", 7), true},
		"negative":  {sdk.Coin{Denom: "adym", Amount: math.NewInt(-1)}, false},
		"bad denom": {sdk.Coin{Denom: "1x", Amount: math.NewInt(7)}, false},
		"no denom":  {sdk.Coin{Amount: math.NewInt(7)}, false},
	} {
		t.Run(name, func(t *testing.T) {
			msg := types.NewMsgRequestValidation(account.String(), "validator", "subject", 0, hash, "")
			msg.Bounty = tc.bounty
			if tc.valid {
				require.NoError(t, msg.ValidateBasic())
			} else {
				require.Error(t, msg.ValidateBasic())
			}
		})
	}
}

func TestReclaimValidationBountyValidateBasic(t *testing.T) {
	_, _, account := testdata.KeyTestPubAddr()
	hash := bytes.Repeat([]byte{1}, 32)
	require.NoError(t, types.NewMsgReclaimValidationBounty(account.String(), hash).ValidateBasic())
	require.ErrorIs(t, types.NewMsgReclaimValidationBounty(account.String(), hash[:31]).ValidateBasic(), types.ErrInvalidValidationHash)
	require.Error(t, types.NewMsgReclaimValidationBounty("bad", hash).ValidateBasic())
}
