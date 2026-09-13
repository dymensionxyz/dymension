package keeper_test

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/stretchr/testify/require"
)

func setupAttestedValidation(t *testing.T) (sdk.Context, *keeper.Keeper, types.MsgServer, *fakeVerifier, *types.MsgRespondValidationAttested) {
	t.Helper()
	ctx, k, v := setup(t)
	p := types.DefaultParams()
	p.ValidationRequestFee.Amount = math.ZeroInt()
	require.NoError(t, k.SetParams(ctx, p))
	validatorOwner := owner(t)
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: "validator", Owner: validatorOwner, Active: true, Policy: policyA(), ActionSeq: 3}))
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: "subject", Owner: owner(t), Active: true, Policy: policyB(), ActionSeq: 2}))
	ms := keeper.NewMsgServerImpl(*k)
	hash := bytes.Repeat([]byte{1}, 32)
	_, err := ms.RequestValidation(ctx, types.NewMsgRequestValidation(owner(t), "validator", "subject", 1, hash, ""))
	require.NoError(t, err)
	msg := &types.MsgRespondValidationAttested{Responder: validatorOwner, RequestHash: hash, Response: 100, ResponseUri: "ipfs://verdict", ResponseHash: bytes.Repeat([]byte{2}, 32), Tag: "tee"}
	payload := types.AttestedValidationBytes(msg.RequestHash, msg.Response, msg.ResponseHash, msg.ResponseUri, msg.Tag)
	msg.Token = []byte(types.ValidationNonce("validator", payload, 3))
	return ctx, k, ms, v, msg
}

func TestAttestedValidationRecordsLogQueryAndGenesis(t *testing.T) {
	ctx, k, ms, v, msg := setupAttestedValidation(t)
	result, err := ms.RespondValidationAttested(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, uint64(0), result.Seq)
	require.Equal(t, 1, v.calls)
	require.Equal(t, policyA(), v.gotPolicy)
	events := ctx.EventManager().Events()
	event := events[len(events)-1]
	require.Equal(t, "dymensionxyz.dymension.agent.EventRespondValidation", event.Type)
	attributes := map[string]string{}
	for _, attr := range event.Attributes {
		attributes[attr.Key] = attr.Value
	}
	require.Equal(t, "true", attributes["attested"])
	response, found := k.GetValidationResponse(ctx, msg.RequestHash, 0)
	require.True(t, found)
	require.True(t, response.Attested)
	require.Equal(t, msg.ResponseHash, response.ResponseHash)
	validator, _ := k.GetAgent(ctx, "validator")
	require.Equal(t, uint64(4), validator.ActionSeq)
	subject, _ := k.GetAgent(ctx, "subject")
	require.Equal(t, uint64(2), subject.ActionSeq)
	log, found := k.GetActionLogEntry(ctx, "validator", 3)
	require.True(t, found)
	payload := types.AttestedValidationBytes(msg.RequestHash, msg.Response, msg.ResponseHash, msg.ResponseUri, msg.Tag)
	require.Equal(t, payload, log.Payload)
	hash := sha256.Sum256(payload)
	require.Equal(t, hash[:], log.PayloadHash)
	query, err := k.ValidationResponses(ctx, &types.QueryValidationResponsesRequest{RequestHash: msg.RequestHash})
	require.NoError(t, err)
	require.True(t, query.ValidationResponses[0].Attested)
	exported := keeper.ExportGenesis(ctx, k)
	require.NoError(t, exported.Validate())
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	json, err := cdc.MarshalJSON(exported)
	require.NoError(t, err)
	var fromJSON types.GenesisState
	require.NoError(t, cdc.UnmarshalJSON(json, &fromJSON))
	require.True(t, fromJSON.ValidationResponses[0].Attested)
	encoded, err := exported.Marshal()
	require.NoError(t, err)
	var decoded types.GenesisState
	require.NoError(t, decoded.Unmarshal(encoded))
	ctx2, k2, _ := setup(t)
	keeper.InitGenesis(ctx2, k2, decoded)
	restored, found := k2.GetValidationResponse(ctx2, msg.RequestHash, 0)
	require.True(t, found)
	require.Equal(t, response, restored)
	_, err = ms.RespondValidationAttested(ctx, msg)
	require.ErrorContains(t, err, "nonce mismatch")
	req, _ := k.GetValidationRequest(ctx, msg.RequestHash)
	require.Equal(t, uint64(1), req.ResponseCount)
}

func TestAttestedValidationRejectsWithoutStateChange(t *testing.T) {
	cases := []string{"invalid token", "action token", "transfer token", "revoked validator", "revoked subject", "oversize", "unauthorized", "inactive validator", "inactive subject", "same owner", "missing request", "bad score", "bad hash", "empty token", "long tag", "long uri"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, k, ms, v, msg := setupAttestedValidation(t)
			payload := types.AttestedValidationBytes(msg.RequestHash, msg.Response, msg.ResponseHash, msg.ResponseUri, msg.Tag)
			switch name {
			case "invalid token":
				msg.Token = []byte("bad")
			case "action token":
				msg.Token = []byte(types.ActionNonce("validator", payload, 3))
			case "transfer token":
				msg.Token = []byte(types.TransferNonce("validator", payload, 3))
			case "revoked validator", "revoked subject":
				id := "validator"
				if name == "revoked subject" {
					id = "subject"
				}
				agent, _ := k.GetAgent(ctx, id)
				fp, err := types.PolicyFingerprint(agent.Policy)
				require.NoError(t, err)
				require.NoError(t, k.SetRevoked(ctx, fp))
			case "oversize":
				p, err := k.GetParams(ctx)
				require.NoError(t, err)
				p.MaxActionBytes = uint64(len(payload) - 1)
				require.NoError(t, k.SetParams(ctx, p))
			case "unauthorized":
				msg.Responder = owner(t)
			case "inactive validator", "inactive subject":
				id := "validator"
				if name == "inactive subject" {
					id = "subject"
				}
				agent, _ := k.GetAgent(ctx, id)
				agent.Active = false
				require.NoError(t, k.SetAgent(ctx, agent))
			case "same owner":
				agent, _ := k.GetAgent(ctx, "subject")
				agent.Owner = msg.Responder
				require.NoError(t, k.SetAgent(ctx, agent))
			case "missing request":
				msg.RequestHash = bytes.Repeat([]byte{9}, 32)
			case "bad score":
				msg.Response = 101
			case "bad hash":
				msg.ResponseHash = []byte{1}
			case "empty token":
				msg.Token = nil
			case "long tag":
				msg.Tag = string(bytes.Repeat([]byte{'x'}, int(types.DefaultValidationTagMaxBytes)+1))
			case "long uri":
				msg.ResponseUri = string(bytes.Repeat([]byte{'x'}, int(types.DefaultValidationUriMaxBytes)+1))
			}
			before := keeper.ExportGenesis(ctx, k)
			events := len(ctx.EventManager().Events())
			_, err := ms.RespondValidationAttested(ctx, msg)
			require.Error(t, err)
			require.Equal(t, before, keeper.ExportGenesis(ctx, k))
			require.Len(t, ctx.EventManager().Events(), events)
			if name != "invalid token" && name != "action token" && name != "transfer token" {
				require.Zero(t, v.calls)
			}
		})
	}
}

func TestAttestedValidationSharesCapWithPlainResponses(t *testing.T) {
	ctx, k, ms, _, msg := setupAttestedValidation(t)
	p, err := k.GetParams(ctx)
	require.NoError(t, err)
	p.ValidationMaxResponsesPerRequest = 2
	require.NoError(t, k.SetParams(ctx, p))
	_, err = ms.RespondValidation(ctx, types.NewMsgRespondValidation(msg.Responder, msg.RequestHash, 50, "", nil, ""))
	require.NoError(t, err)
	plain, _ := k.GetValidationResponse(ctx, msg.RequestHash, 0)
	require.False(t, plain.Attested)
	result, err := ms.RespondValidationAttested(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, uint64(1), result.Seq)
	_, err = ms.RespondValidationAttested(ctx, msg)
	require.ErrorIs(t, err, types.ErrTooManyValidationResponses)
	validator, _ := k.GetAgent(ctx, "validator")
	require.Equal(t, uint64(4), validator.ActionSeq)
}

func TestAttestedValidationTransactionRollback(t *testing.T) {
	ctx, k, ms, _, msg := setupAttestedValidation(t)
	before := keeper.ExportGenesis(ctx, k)
	cacheCtx, _ := ctx.CacheContext()
	_, err := ms.RespondValidationAttested(cacheCtx, msg)
	require.NoError(t, err)
	// A later message failure discards this transaction cache in BaseApp.
	require.Equal(t, before, keeper.ExportGenesis(ctx, k))
}

func TestAttestedValidationPromotesPolicyAndAcceptsExactPayloadLimit(t *testing.T) {
	ctx, k, ms, v, msg := setupAttestedValidation(t)
	validator, _ := k.GetAgent(ctx, "validator")
	next := policyB()
	validator.PendingPolicy = &next
	validator.PendingPolicyHeight = ctx.BlockHeight()
	require.NoError(t, k.SetAgent(ctx, validator))
	payload := types.AttestedValidationBytes(msg.RequestHash, msg.Response, msg.ResponseHash, msg.ResponseUri, msg.Tag)
	p, err := k.GetParams(ctx)
	require.NoError(t, err)
	p.MaxActionBytes = uint64(len(payload))
	p.ValidationMaxResponsesPerRequest = 0
	require.NoError(t, k.SetParams(ctx, p))
	for seq := uint64(3); seq < 5; seq++ {
		msg.Token = []byte(types.ValidationNonce("validator", payload, seq))
		_, err = ms.RespondValidationAttested(ctx, msg)
		require.NoError(t, err)
	}
	require.Equal(t, next, v.gotPolicy)
	validator, _ = k.GetAgent(ctx, "validator")
	require.Equal(t, next, validator.Policy)
	require.Nil(t, validator.PendingPolicy)
	require.Equal(t, uint64(5), validator.ActionSeq)
}
