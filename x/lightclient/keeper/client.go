package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

const (
	ibcRevisionNumber = 1
)

// GetProspectiveCanonicalClient returns the client id of the first IBC client which can be set as the canonical client for the given rollapp.
// Returns empty string if no such client is found.
func (k Keeper) GetProspectiveCanonicalClient(ctx sdk.Context, rollappId string, stateInfo *rollapptypes.StateInfo) (clientID string) {
	clients := k.ibcClientKeeper.GetAllGenesisClients(ctx)
	for _, client := range clients {
		clientState, err := ibcclienttypes.UnpackClientState(client.ClientState)
		if err != nil {
			continue
		}
		// Cast client state to tendermint client state - we need this to get the chain id and state height
		tmClientState, ok := clientState.(*ibctm.ClientState)
		if !ok {
			continue
		}
		if tmClientState.ChainId != rollappId {
			continue
		}
		if !types.IsCanonicalClientParamsValid(tmClientState) {
			continue
		}
		sequencerPk, err := k.GetSequencerPubKey(ctx, stateInfo.Sequencer)
		if err != nil {
			continue
		}
		for _, bd := range stateInfo.GetBDs().BD {
			height := ibcclienttypes.NewHeight(ibcRevisionNumber, bd.GetHeight())
			consensusState, found := k.ibcClientKeeper.GetClientConsensusState(ctx, client.ClientId, height)
			if !found {
				continue
			}
			// Cast consensus state to tendermint consensus state - we need this to check the state root and timestamp and nextValHash
			tmConsensusState, ok := consensusState.(*ibctm.ConsensusState)
			if !ok {
				continue
			}
			ibcState := types.IBCState{
				Root:               tmConsensusState.GetRoot().GetHash(),
				NextValidatorsHash: tmConsensusState.NextValidatorsHash,
				Timestamp:          time.Unix(0, int64(tmConsensusState.GetTimestamp())),
			}
			rollappState := types.RollappState{
				BlockDescriptor: bd,
			}
			// Check if BD for next block exists in same stateinfo
			if stateInfo.ContainsHeight(bd.GetHeight() + 1) {
				rollappState.NextBlockSequencer = sequencerPk
			}
			err := types.CheckCompatibility(ibcState, rollappState)
			if err != nil {
				continue
			}
			clientID = client.GetClientId()
			return
		}
	}
	return
}
