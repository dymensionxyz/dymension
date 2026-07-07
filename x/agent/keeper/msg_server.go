package keeper

import (
	"context"
	"crypto/sha256"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

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

// SubmitAttestedTransfer pays out from the agent's escrow to a recipient,
// authorized by the same TEE attestation + nonce machinery as
// SubmitAttestedAction. The nonce commits to the exact (recipient, denom,
// amount, memo), so the enclave — not the submitter — authorizes the payment.
func (k msgServer) SubmitAttestedTransfer(goCtx context.Context, msg *types.MsgSubmitAttestedTransfer) (*types.MsgSubmitAttestedTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "get params")
	}
	if uint64(len(msg.Memo)) > params.MaxActionBytes {
		return nil, gerrc.ErrInvalidArgument.Wrapf("memo exceeds max action bytes: got %d, max %d", len(msg.Memo), params.MaxActionBytes)
	}

	agent, found := k.GetAgent(ctx, msg.AgentId)
	if !found {
		return nil, gerrc.ErrNotFound.Wrapf("agent: %s", msg.AgentId)
	}
	if !agent.Active {
		return nil, gerrc.ErrFailedPrecondition.Wrapf("agent is not active: %s", msg.AgentId)
	}
	if !agent.SpendEnabled() {
		return nil, errorsmod.Wrap(types.ErrSpendingDisabled, msg.AgentId)
	}

	// Promote a matured pending policy so verification uses the effective policy;
	// the SetAgent below persists the promotion in the same write.
	if agent.PendingPolicy != nil && ctx.BlockHeight() >= agent.PendingPolicyHeight {
		agent.Policy = *agent.PendingPolicy
		agent.PendingPolicy = nil
		agent.PendingPolicyHeight = 0
	}

	seq := agent.ActionSeq
	payload := types.AttestedTransferBytes(msg.Recipient, agent.SpendDenom, msg.Amount, msg.Memo)
	nonce := types.ActionNonce(msg.AgentId, payload, seq)

	// On any failure the tx rolls back, so no state changes. Replay protection
	// is structural, as in SubmitAttestedAction: action_seq below advances, so
	// re-submitting the same transfer re-derives a different nonce.
	if err := k.verifier.Verify(ctx, agent.Policy, nonce, msg.Token); err != nil {
		return nil, errorsmod.Wrap(err, "verify attestation")
	}

	height := uint64(ctx.BlockHeight()) //nolint:gosec // block height is never negative
	if !agent.SpendAllows(height, msg.Amount) {
		return nil, errorsmod.Wrapf(types.ErrSpendBudgetExceeded, "amount %s, remaining %s", msg.Amount, agent.RemainingWindowBudget(height))
	}

	payout := sdk.NewCoins(sdk.NewCoin(agent.SpendDenom, msg.Amount))
	balance, negative := k.GetEscrowBalance(ctx, msg.AgentId).SafeSub(payout...)
	if negative {
		return nil, errorsmod.Wrap(types.ErrInsufficientEscrow, payout.String())
	}
	if err := k.setEscrowBalance(ctx, msg.AgentId, balance); err != nil {
		return nil, errorsmod.Wrap(err, "set escrow balance")
	}
	recipient := sdk.MustAccAddressFromBech32(msg.Recipient)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, payout); err != nil {
		return nil, errorsmod.Wrap(err, "send coins to recipient")
	}

	payloadHash := sha256.Sum256(payload)
	entry := types.ActionLogEntry{
		AgentId:     msg.AgentId,
		Seq:         seq,
		Payload:     payload,
		PayloadHash: payloadHash[:],
		Height:      ctx.BlockHeight(),
		Time:        ctx.BlockTime(),
	}
	if err := k.setActionLogEntry(ctx, entry); err != nil {
		return nil, errorsmod.Wrap(err, "append action log entry")
	}

	agent.RecordSpend(height, msg.Amount)
	agent.ActionSeq = seq + 1
	if err := k.SetAgent(ctx, agent); err != nil {
		return nil, errorsmod.Wrap(err, "set agent")
	}

	if err := uevent.EmitTypedEvent(ctx, &types.EventAttestedTransfer{
		AgentId:   msg.AgentId,
		Seq:       seq,
		Recipient: msg.Recipient,
		Amount:    sdk.NewCoin(agent.SpendDenom, msg.Amount),
	}); err != nil {
		return nil, err
	}

	return &types.MsgSubmitAttestedTransferResponse{Seq: seq}, nil
}
