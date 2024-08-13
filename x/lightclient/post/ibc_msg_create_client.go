package post

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

func (i IBCMessagesDecorator) HandleMsgCreateClient(ctx sdk.Context, msg *ibcclienttypes.MsgCreateClient, success bool) {
	if success {
		clientState, err := ibcclienttypes.UnpackClientState(msg.ClientState)
		if err != nil {
			return
		}
		tendmermintClientState, ok := clientState.(*ibctm.ClientState)
		if !ok {
			return
		}
		rollappID := tendmermintClientState.ChainId
		nextClientID, registrationFound := i.lightClientKeeper.GetCanonicalClient(ctx, rollappID)
		if registrationFound {
			_, clientFound := i.ibcKeeper.ClientKeeper.GetClientState(ctx, nextClientID)
			if clientFound {
				i.lightClientKeeper.SetCanonicalClient(ctx, rollappID, nextClientID)
			}
			i.lightClientKeeper.ClearCanonicalLightClientRegistration(ctx, rollappID)
		}
	}
}
