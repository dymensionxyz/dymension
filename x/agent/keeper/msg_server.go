package keeper

import (
	"context"
	"crypto/sha256"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

var _ types.MsgServer = msgServer{}

// SubmitAttestedAction verifies a TEE attestation token against the agent's
// policy, bound by a per-action nonce, and appends an immutable entry to the
// agent's action log.
func (k msgServer) SubmitAttestedAction(goCtx context.Context, msg *types.MsgSubmitAttestedAction) (*types.MsgSubmitAttestedActionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "get params")
	}
	if uint64(len(msg.Payload)) > params.MaxActionBytes {
		return nil, gerrc.ErrInvalidArgument.Wrapf("payload exceeds max action bytes: got %d, max %d", len(msg.Payload), params.MaxActionBytes)
	}

	agent, found := k.GetAgent(ctx, msg.AgentId)
	if !found {
		return nil, gerrc.ErrNotFound.Wrapf("agent: %s", msg.AgentId)
	}
	if !agent.Active {
		return nil, gerrc.ErrFailedPrecondition.Wrapf("agent is not active: %s", msg.AgentId)
	}

	// Promote a matured pending policy so verification uses the effective policy;
	// the SetAgent below persists the promotion in the same write.
	if agent.PendingPolicy != nil && ctx.BlockHeight() >= agent.PendingPolicyHeight {
		agent.Policy = *agent.PendingPolicy
		agent.PendingPolicy = nil
		agent.PendingPolicyHeight = 0
	}

	seq := agent.ActionSeq
	nonce := types.ActionNonce(msg.AgentId, msg.Payload, seq)

	// On any failure the tx rolls back, so no state changes. Replay protection
	// is structural: action_seq below advances, so re-submitting the same
	// (payload, token) re-derives a different nonce and the verifier rejects.
	if err := k.verifier.Verify(ctx, agent.Policy, nonce, msg.Token); err != nil {
		return nil, errorsmod.Wrap(err, "verify attestation")
	}

	payloadHash := sha256.Sum256(msg.Payload)
	entry := types.ActionLogEntry{
		AgentId:     msg.AgentId,
		Seq:         seq,
		Payload:     msg.Payload,
		PayloadHash: payloadHash[:],
		Height:      ctx.BlockHeight(),
		Time:        ctx.BlockTime(),
	}
	if err := k.setActionLogEntry(ctx, entry); err != nil {
		return nil, errorsmod.Wrap(err, "append action log entry")
	}

	agent.ActionSeq = seq + 1
	if err := k.SetAgent(ctx, agent); err != nil {
		return nil, errorsmod.Wrap(err, "set agent")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAttestedAction,
			sdk.NewAttribute(types.AttributeKeyAgentID, msg.AgentId),
			sdk.NewAttribute(types.AttributeKeySeq, fmt.Sprintf("%d", seq)),
			sdk.NewAttribute(types.AttributeKeyPayloadHash, fmt.Sprintf("%x", payloadHash[:])),
			sdk.NewAttribute(types.AttributeKeySubmitter, msg.Submitter),
		),
	)

	return &types.MsgSubmitAttestedActionResponse{Seq: seq}, nil
}
