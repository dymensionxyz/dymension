package ante

import (
	"bytes"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

func (i IBCMessagesDecorator) HandleMsgUpdateClient(ctx sdk.Context, msg *ibcclienttypes.MsgUpdateClient) error {
	clientState, found := i.ibcKeeper.ClientKeeper.GetClientState(ctx, msg.ClientId)
	if !found {
		return nil
	}
	// Cast client state to tendermint client state - we need this to check the chain id
	tendmermintClientState, ok := clientState.(*ibctm.ClientState)
	if !ok {
		return nil
	}
	// Check if the the client is the canonical client for the rollapp
	rollappID := tendmermintClientState.ChainId
	canonicalClient, found := i.lightClientKeeper.GetCanonicalClient(ctx, rollappID)
	if !found || canonicalClient != msg.ClientId {
		return nil
	}
	// Check if there are existing block descriptors for the given height of client state - if not return and optimistically accept the update
	height := tendmermintClientState.GetLatestHeight()
	stateInfo, err := i.rollappKeeper.FindStateInfoByHeight(ctx, rollappID, height.GetRevisionHeight())
	if err != nil {
		return nil
	}
	blockDescriptor := stateInfo.GetBDs().BD[height.GetRevisionHeight()-stateInfo.GetStartHeight()]
	clientMessage, err := ibcclienttypes.UnpackClientMessage(msg.ClientMessage)
	if err != nil {
		return nil
	}
	header, ok := clientMessage.(*ibctm.Header)
	if !ok {
		return nil
	}
	// Check if block descriptor state root matches tendermint header app hash
	if !bytes.Equal(blockDescriptor.StateRoot, header.Header.AppHash) {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "block descriptor state root does not match tendermint header app hash")
	}
	// TODO: Check if block descriptor timestamp matches tendermint header timestamp
	// Check if the validator set hash matches the sequencer
	// if len(header.ValidatorSet.Validators) == 1 && header.ValidatorSet.Validators[0].Address {

	// }
	return nil
}
