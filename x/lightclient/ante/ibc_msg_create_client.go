package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

func (i IBCMessagesDecorator) HandleMsgCreateClient(ctx sdk.Context, msg *ibcclienttypes.MsgCreateClient) {
	// Unpack client state from message
	clientState, err := ibcclienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return
	}
	// Only tendermint client types can be set as canonical light client
	if clientState.ClientType() == exported.Tendermint {
		// Cast client state to tendermint client state - we need this to get the chain id and state height
		tmClientState, ok := clientState.(*ibctm.ClientState)
		if !ok {
			return
		}
		// Check if rollapp exists for given chain id
		rollappID := tmClientState.ChainId
		_, found := i.rollappKeeper.GetRollapp(ctx, rollappID)
		if !found {
			return
		}
		// Check if canonical client already exists for rollapp. Only one canonical client is allowed per rollapp
		_, found = i.lightClientKeeper.GetCanonicalClient(ctx, rollappID)
		if found {
			return
		}
		// Check if there are existing block descriptors for the given height of client state
		height := tmClientState.GetLatestHeight()
		stateInfo, err := i.rollappKeeper.FindStateInfoByHeight(ctx, rollappID, height.GetRevisionHeight())
		if err != nil {
			return
		}
		blockDescriptor := stateInfo.GetBDs().BD[height.GetRevisionHeight()-stateInfo.GetStartHeight()]
		// Unpack consensus state from message
		consensusState, err := ibcclienttypes.UnpackConsensusState(msg.ConsensusState)
		if err != nil {
			return
		}
		// Cast consensus state to tendermint consensus state - we need this to get the state root and timestamp and nextValHash
		tmConsensusState, ok := consensusState.(*ibctm.ConsensusState)
		if !ok {
			return
		}
		// Check if the consensus state is compatible with the block descriptor state
		err = types.StateCompatible(*tmConsensusState, blockDescriptor)
		if err != nil {
			return // In case of incompatibility, the client will be created but not set as canonical
		}

		// Generate client id and begin canonical light client registration by storing it in transient store.
		// Will be confirmed after the client is created in post handler.
		nextClientID := i.ibcKeeper.ClientKeeper.GenerateClientIdentifier(ctx, exported.Tendermint)
		i.lightClientKeeper.BeginCanonicalLightClientRegistration(ctx, rollappID, nextClientID)
	}
}
