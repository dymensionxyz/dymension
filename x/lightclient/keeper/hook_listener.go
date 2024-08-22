package keeper

import (
	"time"

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

func (hook rollappHook) AfterUpdateState(
	ctx sdk.Context,
	rollappId string,
	stateInfo *rollapptypes.StateInfo,
	isFirstStateUpdate bool,
	previousStateHasTimestamp bool,
) error {
	canonicalClient, found := hook.k.GetCanonicalClient(ctx, rollappId)
	if !found {
		return nil
	}
	bds := stateInfo.GetBDs()
	for i, bd := range bds.GetBD() {
		// Check if any optimistic updates were made for the given height
		tmHeaderSigner, found := hook.k.GetConsensusStateSigner(ctx, canonicalClient, bd.GetHeight())
		if !found {
			continue
		}
		height := ibcclienttypes.NewHeight(1, bd.GetHeight())
		consensusState, _ := hook.k.ibcClientKeeper.GetClientConsensusState(ctx, canonicalClient, height)
		// Cast consensus state to tendermint consensus state - we need this to check the state root and timestamp and nextValHash
		tmConsensusState, ok := consensusState.(*ibctm.ConsensusState)
		if !ok {
			return nil
		}

		// Convert timestamp from nanoseconds to time.Time
		timestamp := time.Unix(0, int64(tmConsensusState.GetTimestamp()))

		ibcState := types.IBCState{
			Root:               tmConsensusState.GetRoot().GetHash(),
			Height:             bd.GetHeight(),
			Validator:          []byte(tmHeaderSigner),
			NextValidatorsHash: tmConsensusState.NextValidatorsHash,
			Timestamp:          timestamp,
		}
		sequencerPk, err := hook.k.GetTmPubkeyAsBytes(ctx, stateInfo.Sequencer)
		if err != nil {
			return err
		}
		rollappState := types.RollappState{
			BlockSequencer:  sequencerPk,
			BlockDescriptor: bd,
		}
		// check if bd for next block exists in same state info
		if i+1 < len(bds.GetBD()) {
			rollappState.NextBlockSequencer = sequencerPk
			rollappState.NextBlockDescriptor = bds.GetBD()[i+1]
		}
		err = types.CheckCompatibility(ibcState, rollappState)
		if err != nil {
			// Only require timestamp on BD if first ever update, or the previous update had BD
			if err == types.ErrTimestampNotFound && !isFirstStateUpdate && !previousStateHasTimestamp {
				continue
			}

			// The BD for (h+1) is missing, cannot verify if the nextvalhash matches
			if err == types.ErrNextBlockDescriptorMissing {
				return err
			}

			// If the state is not compatible,
			// Take this state update as source of truth over the IBC update
			// Punish the block proposer of the IBC signed header
			sequencerAddr, err := hook.k.getAddress([]byte(tmHeaderSigner))
			if err != nil {
				return err
			}
			err = hook.k.sequencerKeeper.JailSequencerOnFraud(ctx, sequencerAddr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
