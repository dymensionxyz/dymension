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
		if err := a.ValidateSpendState(); err != nil {
			return fmt.Errorf("agent %s: %w", a.Id, err)
		}
		agentIDs[a.Id] = struct{}{}
	}

	seen := make(map[string]struct{}, len(g.Escrows))
	for _, e := range g.Escrows {
		if _, ok := agentIDs[e.AgentId]; !ok {
			// funds without a registered owner would be trapped in the module account
			return fmt.Errorf("escrow for unknown agent: %s", e.AgentId)
		}
		if _, dup := seen[e.AgentId]; dup {
			return fmt.Errorf("duplicate escrow entry: %s", e.AgentId)
		}
		seen[e.AgentId] = struct{}{}
		if err := e.Balance.Validate(); err != nil {
			return fmt.Errorf("escrow balance for agent %s: %w", e.AgentId, err)
		}
		if e.Balance.IsZero() {
			return fmt.Errorf("empty escrow balance for agent: %s", e.AgentId)
		}
	}
	return nil
}
