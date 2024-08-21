package ante

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

func (i IBCMessagesDecorator) HandleMsgCreateClient(ctx sdk.Context, msg *ibcclienttypes.MsgCreateClient) {
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

		consensusState, err := ibcclienttypes.UnpackConsensusState(msg.ConsensusState)
		if err != nil {
			return
		}
		// Cast consensus state to tendermint consensus state - we need this to get the state root and timestamp and nextValHash
		tmConsensusState, ok := consensusState.(*ibctm.ConsensusState)
		if !ok {
			return
		}
		// Convert timestamp from nanoseconds to time.Time
		timestamp := time.Unix(0, int64(tmConsensusState.GetTimestamp()))

		ibcState := types.IBCState{
			Root:               tmConsensusState.GetRoot().GetHash(),
			Height:             tmClientState.GetLatestHeight().GetRevisionHeight(),
			Validator:          []byte{}, // This info is available in the tendermint consensus state as no header has been shared yet
			NextValidatorsHash: tmConsensusState.NextValidatorsHash,
			Timestamp:          timestamp,
		}
		sequencerPk, err := i.lightClientKeeper.GetTmPubkeyAsBytes(ctx, stateInfo.Sequencer)
		if err != nil {
			return
		}
		rollappState := types.RollappState{
			BlockSequencer:  sequencerPk,
			BlockDescriptor: blockDescriptor,
		}
		// Check if bd for next block exists and is part of same state info
		nextHeight := height.GetRevisionHeight() + 1
		if stateInfo.StartHeight+stateInfo.NumBlocks >= nextHeight {
			rollappState.NextBlockDescriptor = stateInfo.GetBDs().BD[nextHeight-stateInfo.StartHeight]
			rollappState.NextBlockSequencer = sequencerPk
		} else {
			// nextBD doesnt exist in same stateInfo. So lookup in the next StateInfo
			currentStateInfoIndex := stateInfo.GetIndex().Index
			nextStateInfo, found := i.rollappKeeper.GetStateInfo(ctx, rollappID, currentStateInfoIndex+1)
			if !found {
				return // There is no BD for h+1, so we can't verify the next block valhash. So we cant mark this client as canonical
			} else {
				nextSequencerPk, err := i.lightClientKeeper.GetTmPubkeyAsBytes(ctx, nextStateInfo.Sequencer)
				if err != nil {
					return
				}
				rollappState.NextBlockSequencer = nextSequencerPk
				rollappState.NextBlockDescriptor = nextStateInfo.GetBDs().BD[0]
			}
		}
		// Check if the consensus state is compatible with the block descriptor state
		err = types.CheckCompatibility(ibcState, rollappState)
		if err != nil {
			return // In case of incompatibility, the client will be created but not set as canonical
		}

		// Ensure the light client params conform to expected values
		if !types.IsCanonicalClientParamsValid(tmClientState) {
			return // In case of invalid params, the client will be created but not set as canonical
		}
		// Generate client id and begin canonical light client registration by storing it in transient store.
		// Will be confirmed after the client is created in post handler.
		nextClientID := i.ibcClientKeeper.GenerateClientIdentifier(ctx, exported.Tendermint)
		i.lightClientKeeper.BeginCanonicalLightClientRegistration(ctx, rollappID, nextClientID)
	}
}
