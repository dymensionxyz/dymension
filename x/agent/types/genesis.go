package types

import "fmt"

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

func (g GenesisState) Validate() error {
	if err := g.Params.Validate(); err != nil {
		return err
	}
	agentIDs := make(map[string]struct{}, len(g.Agents))
	for _, a := range g.Agents {
		if _, dup := agentIDs[a.Id]; dup {
			return fmt.Errorf("duplicate agent entry: %s", a.Id)
		}
		if err := a.ValidateSpendState(); err != nil {
			return fmt.Errorf("agent %s: %w", a.Id, err)
		}
		agentIDs[a.Id] = struct{}{}
	}

	escrowSeen := make(map[string]struct{}, len(g.Escrows))
	for _, e := range g.Escrows {
		if _, ok := agentIDs[e.AgentId]; !ok {
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
	return nil
}
