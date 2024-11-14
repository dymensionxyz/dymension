package types

import "fmt"

func DefaultGenesisState() GenesisState {
	return GenesisState{
		CanonicalClients: []CanonicalClient{},
	}
}

func (g GenesisState) Validate() error {
	for _, client := range g.CanonicalClients {
		if client.RollappId == "" {
			return fmt.Errorf("invalid rollapp id: %v", client)
		}
		if client.IbcClientId == "" {
			return fmt.Errorf("invalid ibc client id: %v", client)
		}
	}

	return nil
}
