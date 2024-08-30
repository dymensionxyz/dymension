package types

import fmt "fmt"

func DefaultGenesisState() GenesisState {
	return GenesisState{
		CanonicalClients:      []CanonicalClient{},
		ConsensusStateSigners: []ConsensusStateSigner{},
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
	for _, stateSigner := range g.ConsensusStateSigners {
		if stateSigner.IbcClientId == "" {
			return fmt.Errorf("invalid ibc client id: %v", stateSigner)
		}
		if stateSigner.Height == 0 {
			return fmt.Errorf("invalid height: %v", stateSigner)
		}
		if stateSigner.BlockValHash == "" {
			return fmt.Errorf("invalid signer: %v", stateSigner)
		}
	}
	return nil
}
