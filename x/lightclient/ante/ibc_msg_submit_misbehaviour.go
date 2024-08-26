package ante

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

func (i IBCMessagesDecorator) HandleMsgSubmitMisbehaviour(ctx sdk.Context, msg *ibcclienttypes.MsgSubmitMisbehaviour) error {
	clientState, found := i.ibcClientKeeper.GetClientState(ctx, msg.ClientId)
	if !found {
		return nil
	}
	// Cast client state to tendermint client state - we need this to get the chain id
	tendmermintClientState, ok := clientState.(*ibctm.ClientState)
	if !ok {
		return nil
	}
	// Check if the client is the canonical client for a rollapp
	rollappID := tendmermintClientState.ChainId
	canonicalClient, found := i.lightClientKeeper.GetCanonicalClient(ctx, rollappID)
	if canonicalClient == msg.ClientId {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidClient, "cannot submit misbehavour for a canonical client")
	}
	return nil
}
