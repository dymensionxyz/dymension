package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var _ rollapptypes.RollappHooks = rollappHook{}

// Hooks wrapper struct for rollapp keeper.
type rollappHook struct {
	rollapptypes.StubRollappCreatedHooks
	k Keeper
}

// RollappHooks returns the wrapper struct.
func (k Keeper) RollappHooks() rollapptypes.RollappHooks {
	return rollappHook{k: k}
}

func (hook rollappHook) AfterUpdateState(ctx sdk.Context, rollappId string, stateInfo *rollapptypes.StateInfo) error {
	canonicalClient, found := hook.k.GetCanonicalClient(ctx, rollappId)
	if !found {
		return nil
	}
	bds := stateInfo.GetBDs()
	for _, bd := range bds.GetBD() {
		// Check if any optimistic updates were made for the given height
		signer, found := hook.k.GetConsensusStateSigner(ctx, canonicalClient, bd.GetHeight())
		if !found {
			continue
		}
		height := ibcclienttypes.NewHeight(1, bd.GetHeight())
		consensusState, found := hook.k.ibcKeeper.ClientKeeper.GetClientConsensusState(ctx, canonicalClient, height)
		if !found {
			return nil
		}
		// Cast consensus state to tendermint consensus state - we need this to check the state root and timestamp and nextValHash
		tmConsensusState, ok := consensusState.(*ibctm.ConsensusState)
		if !ok {
			return nil
		}

		err := types.StateCompatible(*tmConsensusState, bd)
		if err != nil {
			// If the state is not compatible,
			// Take this state update as source of truth over the IBC update
			// Punish the block proposer of the IBC signed header
			sequencerAddr := signer // todo: signer addr is sent from tm. so will be valconsaddr(?). check and then transform to valid address
			err = hook.k.sequencerKeeper.SlashAndJailFraud(ctx, sequencerAddr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
