package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

func (g GenesisState) Validate() error {
	if err := g.Params.Validate(); err != nil {
		return err
	}
	agents := make(map[string]Agent, len(g.Agents))
	for _, a := range g.Agents {
		if _, dup := agents[a.Id]; dup {
			return fmt.Errorf("duplicate agent entry: %s", a.Id)
		}
		if err := a.ValidateSpendState(g.Params.SpendRecipientAllowlistMax); err != nil {
			return fmt.Errorf("agent %s: %w", a.Id, err)
		}
		agents[a.Id] = a
	}

	escrowSeen := make(map[string]struct{}, len(g.Escrows))
	for _, e := range g.Escrows {
		if _, ok := agents[e.AgentId]; !ok {
			// funds without a registered owner would be trapped in the module account
			return fmt.Errorf("escrow for unknown agent: %s", e.AgentId)
		}
		if _, dup := escrowSeen[e.AgentId]; dup {
			return fmt.Errorf("duplicate escrow entry: %s", e.AgentId)
		}
		escrowSeen[e.AgentId] = struct{}{}
		if err := e.Balance.Validate(); err != nil {
			return fmt.Errorf("escrow balance for agent %s: %w", e.AgentId, err)
		}
		if e.Balance.IsZero() {
			return fmt.Errorf("empty escrow balance for agent: %s", e.AgentId)
		}
	}
	for _, fp := range g.RevokedPolicies {
		if err := ValidateFingerprint(fp); err != nil {
			return err
		}
	}
	feedbackSeen := make(map[string]struct{}, len(g.Feedbacks))
	for _, f := range g.Feedbacks {
		if f.Score > MaxFeedbackScore {
			return fmt.Errorf("feedback score %d exceeds max %d: agent %s client %s", f.Score, MaxFeedbackScore, f.AgentId, f.Client)
		}
		key := f.AgentId + "/" + f.Client
		if _, ok := feedbackSeen[key]; ok {
			return fmt.Errorf("duplicate feedback: agent %s client %s", f.AgentId, f.Client)
		}
		feedbackSeen[key] = struct{}{}
	}
	requests := make(map[string]ValidationRequest, len(g.ValidationRequests))
	for _, v := range g.ValidationRequests {
		if len(v.RequestHash) != 32 {
			return ErrInvalidValidationHash
		}
		if _, ok := requests[string(v.RequestHash)]; ok {
			return ErrValidationRequestExists
		}
		requests[string(v.RequestHash)] = v
	}
	responses := make(map[string]struct{}, len(g.ValidationResponses))
	responseCounts := make(map[string]uint64, len(g.ValidationRequests))
	for _, v := range g.ValidationResponses {
		request, ok := requests[string(v.RequestHash)]
		if !ok {
			return ErrValidationRequestNotFound
		}
		if v.Seq >= request.ResponseCount {
			return fmt.Errorf("non-contiguous validation response sequence: hash %x seq %d count %d", v.RequestHash, v.Seq, request.ResponseCount)
		}
		responseKey := fmt.Sprintf("%x/%d", v.RequestHash, v.Seq)
		if _, ok := responses[responseKey]; ok {
			return fmt.Errorf("duplicate validation response: %s", responseKey)
		}
		responses[responseKey] = struct{}{}
		responseCounts[string(v.RequestHash)]++
	}
	for hash, request := range requests {
		if responseCounts[hash] != request.ResponseCount {
			return fmt.Errorf("validation response count mismatch: hash %x stored %d actual %d", request.RequestHash, request.ResponseCount, responseCounts[hash])
		}
	}
	return validateRecipientSpends(g.RecipientSpends, agents)
}

// validateRecipientSpends accepts only records the runtime could have written:
// for a capped agent, an allowlisted recipient, a bucket-aligned window and a
// total within the cap.
func validateRecipientSpends(spends []RecipientSpend, agents map[string]Agent) error {
	seen := make(map[[2]string]struct{}, len(spends))
	for _, r := range spends {
		agent, ok := agents[r.AgentId]
		if !ok {
			return fmt.Errorf("recipient spend for unknown agent: %s", r.AgentId)
		}
		if !agent.RecipientSpendEnforced() {
			return fmt.Errorf("recipient spend for agent without per-recipient limit: %s", r.AgentId)
		}
		recipient, err := sdk.AccAddressFromBech32(r.Recipient)
		if err != nil {
			return fmt.Errorf("recipient spend address: agent %s: %w", r.AgentId, err)
		}
		if !agent.SpendRecipientAllowed(r.Recipient) {
			return fmt.Errorf("recipient spend for non-allowlisted recipient: agent %s recipient %s", r.AgentId, r.Recipient)
		}
		key := [2]string{r.AgentId, string(recipient)}
		if _, dup := seen[key]; dup {
			return fmt.Errorf("duplicate recipient spend: agent %s recipient %s", r.AgentId, r.Recipient)
		}
		seen[key] = struct{}{}
		if r.Window.WindowStartHeight%agent.SpendWindowBlocks != 0 {
			return fmt.Errorf("recipient spend window start height not bucket-aligned: agent %s recipient %s", r.AgentId, r.Recipient)
		}
		if r.Window.Spent.IsNil() || r.Window.Spent.IsNegative() || r.Window.Spent.GT(agent.SpendLimitPerRecipientPerWindow) {
			return fmt.Errorf("recipient spend out of bounds: agent %s recipient %s", r.AgentId, r.Recipient)
		}
	}
	return nil
}
