package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

func (i IBCMessagesDecorator) HandleMsgUpdateClient(ctx sdk.Context, msg *ibcclienttypes.MsgUpdateClient) error {
	clientState, found := i.ibcClientKeeper.GetClientState(ctx, msg.ClientId)
	if !found {
		return nil
	}
	// Cast client state to tendermint client state - we need this to get the chain id(rollapp id)
	tmClientState, ok := clientState.(*ibctm.ClientState)
	if !ok {
		return nil
	}
	// Check if the client is the canonical client for the rollapp
	rollappID := tmClientState.ChainId
	canonicalClient, _ := i.lightClientKeeper.GetCanonicalClient(ctx, rollappID)
	if canonicalClient != msg.ClientId {
		return nil // The client is not a rollapp's canonical client. Continue with default behaviour.
	}
	clientMessage, err := ibcclienttypes.UnpackClientMessage(msg.ClientMessage)
	if err != nil {
		return nil
	}
	header, ok := clientMessage.(*ibctm.Header)
	if !ok {
		return nil
	}

	// Check if there are existing block descriptors for the given height of client state
	height := uint64(header.Header.Height)
	stateInfo, err := i.rollappKeeper.FindStateInfoByHeight(ctx, rollappID, height)
	if err != nil {
		// No BDs found for given height.
		// Will accept the update optimistically
		// But also save the blockProposer address with the height for future verification
		i.acceptUpdateOptimistically(ctx, msg.ClientId, header)
		return nil
	}
	bd, _ := stateInfo.GetBlockDescriptor(height)

	stateInfo, err = i.rollappKeeper.FindStateInfoByHeight(ctx, rollappID, height+1)
	if err != nil {
		// No BDs found for next height.
		// Will accept the update optimistically
		// But also save the blockProposer address with the height for future verification
		i.acceptUpdateOptimistically(ctx, msg.ClientId, header)
		return nil
	}
	sequencerPubKey, err := i.lightClientKeeper.GetSequencerPubKey(ctx, stateInfo.Sequencer)
	if err != nil {
		return err
	}
	rollappState := types.RollappState{
		BlockDescriptor:    bd,
		NextBlockSequencer: sequencerPubKey,
	}
	// Ensure that the ibc header is compatible with the existing rollapp state
	// If it's not, we error and prevent the MsgUpdateClient from being processed
	err = types.CheckCompatibility(*header.ConsensusState(), rollappState)
	if err != nil {
		return err
	}

	return nil
}

func (i IBCMessagesDecorator) acceptUpdateOptimistically(ctx sdk.Context, clientID string, header *ibctm.Header) {
	i.lightClientKeeper.SetConsensusStateValHash(ctx, clientID, uint64(header.Header.Height), header.Header.ValidatorsHash)
}
