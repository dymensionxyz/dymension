package types

import (
	"fmt"
)

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

func (g GenesisState) Validate() error {
	if err := g.Params.Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}
	ids := make(map[string]bool, len(g.Agents))
	for _, a := range g.Agents {
		if a.Id == "" {
			return fmt.Errorf("agent id cannot be empty")
		}
		if ids[a.Id] {
			return fmt.Errorf("duplicate agent id: %s", a.Id)
		}
		ids[a.Id] = true
	}
	for _, e := range g.ActionLog {
		if !ids[e.AgentId] {
			return fmt.Errorf("action log entry references unknown agent: %s", e.AgentId)
		}
	}
	return nil
}
