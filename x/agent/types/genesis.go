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
	for _, fp := range g.RevokedPolicies {
		if err := ValidateFingerprint(fp); err != nil {
			return err
		}
	}
	seen := make(map[string]struct{}, len(g.Feedbacks))
	for _, f := range g.Feedbacks {
		if f.Score > MaxFeedbackScore {
			return fmt.Errorf("feedback score %d exceeds max %d: agent %s client %s", f.Score, MaxFeedbackScore, f.AgentId, f.Client)
		}
		key := f.AgentId + "/" + f.Client
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate feedback: agent %s client %s", f.AgentId, f.Client)
		}
		seen[key] = struct{}{}
	}
	return nil
}
