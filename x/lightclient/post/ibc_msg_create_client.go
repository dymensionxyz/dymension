package post

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

func (i IBCMessagesDecorator) HandleMsgCreateClient(ctx sdk.Context, msg *ibcclienttypes.MsgCreateClient, success bool) {
	clientState, err := ibcclienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return
	}
	// Parse client state to tendermint client state to get the chain id
	tmClientState, ok := clientState.(*ibctm.ClientState)
	if !ok {
		return
	}
	rollappID := tmClientState.ChainId
	// If tx failed, no need to proceed with canonical client registration
	if success {
		// Check if a client registration is in progress
		nextClientID, registrationFound := i.lightClientKeeper.GetCanonicalLightClientRegistration(ctx, rollappID)
		if registrationFound {
			// Check if the client was successfully created with given clientID
			_, clientFound := i.ibcClientKeeper.GetClientState(ctx, nextClientID)
			if clientFound {
				// Set the client as the canonical client for the rollapp
				i.lightClientKeeper.SetCanonicalClient(ctx, rollappID, nextClientID)
			}

		}
	}
	// Always clear the registration after the tx as the transient store is shared among all txs in the block
	i.lightClientKeeper.ClearCanonicalLightClientRegistration(ctx, rollappID)
}
