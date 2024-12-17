package types

import "fmt"

func DefaultGenesisState() GenesisState {
	return GenesisState{
		CanonicalClients: []CanonicalClient{},
	}
}

func (g GenesisState) Validate() error {

	clientSet := make(map[string]struct{})
	for _, client := range g.CanonicalClients {
		if client.RollappId == "" {
			return fmt.Errorf("invalid rollapp_id: %v", client)
		}
		if client.IbcClientId == "" {
			return fmt.Errorf("invalid ibc_client_id: %v", client)
		}

		clientKey := client.RollappId + ":" + client.IbcClientId
		if _, exists := clientSet[clientKey]; exists {
			return fmt.Errorf("duplicate canonical client found: %v", client)
		}
		clientSet[clientKey] = struct{}{}
	}

	signerSet := make(map[string]struct{})
	for _, signer := range g.HeaderSigners {
		if signer.SequencerAddress == "" {
			return fmt.Errorf("invalid sequencer address: %v", signer)
		}
		if signer.ClientId == "" {
			return fmt.Errorf("invalid client id: %v", signer)
		}

		signerKey := signer.SequencerAddress + ":" + fmt.Sprint(signer.Height)
		if _, exists := signerSet[signerKey]; exists {
			return fmt.Errorf("duplicate signer entry found: %v", signer)
		}
		signerSet[signerKey] = struct{}{}
	}

	return nil
}
