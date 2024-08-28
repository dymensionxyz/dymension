package keeper

import (
	"errors"
	"time"

	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
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

// AfterUpdateState is called after a state update is made to a rollapp.
// This hook checks if the rollapp has a canonical IBC light client and if the Consensus state is compatible with the state update
// and punishes the sequencer if it is not
func (hook rollappHook) AfterUpdateState(
	ctx sdk.Context,
	rollappId string,
	stateInfo *rollapptypes.StateInfo,
) error {
	canonicalClient, found := hook.k.GetCanonicalClient(ctx, rollappId)
	if !found {
		canonicalClient = hook.k.GetProspectiveCanonicalClient(ctx, rollappId, stateInfo)
		if canonicalClient != "" {
			hook.k.SetCanonicalClient(ctx, rollappId, canonicalClient)
		}
		return nil
	}
	bds := stateInfo.GetBDs()
	for i, bd := range bds.GetBD() {
		// Check if any optimistic updates were made for the given height
		tmHeaderSigner, found := hook.k.GetConsensusStateSigner(ctx, canonicalClient, bd.GetHeight())
		if !found {
			continue
		}
		height := ibcclienttypes.NewHeight(ibcRevisionNumber, bd.GetHeight())
		consensusState, _ := hook.k.ibcClientKeeper.GetClientConsensusState(ctx, canonicalClient, height)
		// Cast consensus state to tendermint consensus state - we need this to check the state root and timestamp and nextValHash
		tmConsensusState, ok := consensusState.(*ibctm.ConsensusState)
		if !ok {
			return nil
		}
		ibcState := types.IBCState{
			Root:               tmConsensusState.GetRoot().GetHash(),
			ValidatorsHash:     tmHeaderSigner,
			NextValidatorsHash: tmConsensusState.NextValidatorsHash,
			Timestamp:          time.Unix(0, int64(tmConsensusState.GetTimestamp())),
		}
		sequencerPk, err := hook.k.GetSequencerPubKey(ctx, stateInfo.Sequencer)
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
			// The BD for (h+1) is missing, cannot verify if the nextvalhash matches
			if errors.Is(err, types.ErrNextBlockDescriptorMissing) {
				return err
			}
			// If the state is not compatible,
			// Take this state update as source of truth over the IBC update
			// Punish the block proposer of the IBC signed header
			sequencerAddr, err := getAddress(tmHeaderSigner)
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

// getAddress converts a tendermint public key to a bech32 address
func getAddress(tmPubkeyBz []byte) (string, error) {
	var tmpk tmprotocrypto.PublicKey
	err := tmpk.Unmarshal(tmPubkeyBz)
	if err != nil {
		return "", err
	}
	pubkey, err := cryptocodec.FromTmProtoPublicKey(tmpk)
	if err != nil {
		return "", err
	}
	acc := sdk.AccAddress(pubkey.Address().Bytes())
	return acc.String(), nil
}
