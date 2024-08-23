package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

func (i IBCMessagesDecorator) HandleMsgUpdateClient(ctx sdk.Context, msg *ibcclienttypes.MsgUpdateClient) error {
	clientState, found := i.ibcClientKeeper.GetClientState(ctx, msg.ClientId)
	if !found {
		return nil
	}
	// Only continue if the client is a tendermint client as rollapp only supports tendermint clients as canonical clients
	if clientState.ClientType() == exported.Tendermint {
		// Cast client state to tendermint client state - we need this to get the chain id(rollapp id)
		tmClientState, ok := clientState.(*ibctm.ClientState)
		if !ok {
			return nil
		}
		// Check if the client is the canonical client for the rollapp
		rollappID := tmClientState.ChainId
		canonicalClient, found := i.lightClientKeeper.GetCanonicalClient(ctx, rollappID)
		if !found || canonicalClient != msg.ClientId {
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
		height := clientState.GetLatestHeight()
		stateInfo, err := i.rollappKeeper.FindStateInfoByHeight(ctx, rollappID, height.GetRevisionHeight())
		if err != nil {
			// No BDs found for given height.
			// Will accept the update optimistically
			// But also save the blockProposer address with the height for future verification
			blockProposer := header.Header.ProposerAddress
			i.lightClientKeeper.SetConsensusStateSigner(ctx, msg.ClientId, height.GetRevisionHeight(), string(blockProposer))
			return nil
		}
		bd := stateInfo.GetBDs().BD[height.GetRevisionHeight()-stateInfo.GetStartHeight()]

		ibcState := types.IBCState{
			Root:               header.Header.GetAppHash(),
			Height:             height.GetRevisionHeight(),
			Validator:          header.Header.ProposerAddress,
			NextValidatorsHash: header.Header.NextValidatorsHash,
			Timestamp:          header.Header.Time,
		}
		sequencerPubKey, err := i.lightClientKeeper.GetTmPubkeyAsBytes(ctx, stateInfo.Sequencer)
		if err != nil {
			return err
		}
		rollappState := types.RollappState{
			BlockSequencer:  sequencerPubKey,
			BlockDescriptor: bd,
		}
		// Check that BD for next block exists in same stateinfo
		if height.GetRevisionHeight()-stateInfo.GetStartHeight()+1 < uint64(len(stateInfo.GetBDs().BD)) {
			rollappState.NextBlockSequencer = sequencerPubKey
			rollappState.NextBlockDescriptor = stateInfo.GetBDs().BD[height.GetRevisionHeight()-stateInfo.GetStartHeight()+1]
		} else {
			// next BD does not exist in this state info, check the next state info
			nextStateInfo, found := i.rollappKeeper.GetStateInfo(ctx, rollappID, stateInfo.GetIndex().Index+1)
			if found {
				nextSequencerPk, err := i.lightClientKeeper.GetTmPubkeyAsBytes(ctx, nextStateInfo.Sequencer)
				if err != nil {
					return err
				}
				rollappState.NextBlockSequencer = nextSequencerPk
				rollappState.NextBlockDescriptor = nextStateInfo.GetBDs().BD[0]
			} else {
				// if next state info does not exist, then we can't verify the next block valhash.
				// Will accept the update optimistically
				// But also save the blockProposer address with the height for future verification
				blockProposer := header.Header.ProposerAddress
				i.lightClientKeeper.SetConsensusStateSigner(ctx, msg.ClientId, height.GetRevisionHeight(), string(blockProposer))
				return nil
			}
		}
		// Ensure that the ibc header is compatible with the existing rollapp state
		// If its not, we error and prevent the MsgUpdateClient from being processed
		err = types.CheckCompatibility(ibcState, rollappState)
		if err != nil {
			return err
		}
	}
	return nil
}
