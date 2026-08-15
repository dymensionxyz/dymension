package types_test

import (
	"bytes"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
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
