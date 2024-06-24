package keeper

import (
	"bytes"
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v6/modules/core/keeper"
	tendermint "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

/*
Need to think about
- sequencer change
- valsets of size more than 1
- adjacent/non-adjacent updates
- chain id / client id
- pruning
	cons states will be pruned when an old header no longer falls under trusting period, based on timestamp

A note on headers:
	In a header with height H:
		- The AppHash is for height H-1
		- The ValidatorsHash is for height H
		- The NextValidatorsHash is for height H+1
	Crucially, that means when the APP returns the 'next validator set' X to ABCI at height H:
		- Header H+1 contains the Root of H
		- Header H+1 contains the NextValidatorsHash (hash of X)
		- Header H+2 contains ValidatorsHash (hash of X)
	That means,
	See
		- https://github.com/tendermint/tendermint/blob/v0.34.x/spec/core/data_structures.md#header
		- https://github.com/cometbft/cometbft/blob/main/spec/abci/abci%2B%2B_methods.md#finalizeblock

Walkthrough:
	MsgCreateClient
		It just contains the client and cons state.
		In the TM case the relevant parts are
			- ChainID
			- TrustLevel(?)
			- Height
			- NextValidatorsHash
			- AppHash
	Update Client
		Contains:
			- A header
			- The set of validators who signed the header
			- A full set of trusted validators and a trusted height
		Steps
			Make sure the trusted NextValidatorsHash = the hash of the so-called trusted validators
			Make sure height is strictly increasing
			Make sure the hash of the validator set = the validatorsHash in the signed header
			Adjacent:
				Make sure the validatorsHash in the signed header = trusted NextValidatorsHash
			Non adjacent:
				Make sure +1/3 of the validator set in the trusted set signed the signed header
				In this way, at least one correct validator signed it
			Make sure +2/3 of the validator set signed the signed header
			Assuming above all OK, we store the timestamp, appHash and NextValidatorsHash

What do we need to do?
	Make sure the sequencer sent the right header / state root
	Make sure the light client is actually from the sequencer


How could it look?
	On update client




*/

type lcv struct {
	ibckeeper *ibckeeper.Keeper
}

type clientUpdateEventRaw struct {
	misbehavior     bool
	clientID        string
	clientType      string //  exported.ClientState.ClientType()
	consensusHeight string //  header.GetHeight().String()
	headerStr       string //  hex.EncodeToString(types.MustMarshalHeader(k.cdc, header))
}

// TODO: need to be careful with approach of using events, because they can still be emitted even if tx fails?
type clientUpdateEvent struct {
	clientID        string
	clientType      string
	consensusHeight exported.Height
	header          exported.Header
}

// TODO: maybe this can go before the ante handler?
func (v lcv) verifyCreateClient(
	ctx sdk.Context,
	clientState exported.ClientState,
	consensusState exported.ConsensusState,
) error {
	// TODO:
	return nil
}

// if the ibc module accepts a new consensus state, then it must have been signed by 2/3 of the voting power
// of the NEW validators, as well as 1/3 of the previously trusted validator set
// therefore, as long as know that the first every consensus state was signed by the rollapp sequencer
// then we know all future consensus states were signed by him too
// so we just need to check the sequencer signed the first ever one
// the first thing the relayer ever does is MsgCreateClient
//
//	https://github.com/dymensionxyz/go-relayer/blob/9ea09f3db32af59907c7fd598258f4ee53390e36/relayer/chains/cosmos/tx.go#L729-L731
//
// the data originates from the relayer querying the rpc node comet layer
//
//	https://github.com/cometbft/cometbft/blob/v0.34.33/light/provider/http/http.go#L69
//
// getting my thoughts down:
// need to verify the signature on the first client creation
// thereafter, can rely on induction
// at this time you can also set the canonical client
func (v lcv) verifyNewLightClientHeader(ctx sdk.Context, evt clientUpdateEvent) error {
	consState, ok := v.ibckeeper.ClientKeeper.GetClientConsensusState(ctx, evt.clientID, evt.consensusHeight)
	if !ok {
		return gerrc.ErrNotFound
	}

	tmConsState, ok := consState.(*tendermint.ConsensusState)
	if !ok {
		return errorsmod.WithType(gerrc.ErrInvalidArgument, consState)
	}

	_ = tmConsState

	return nil
}

func (v lcv) sergisType(consState exported.ConsensusState) *tendermint.ConsensusState {
	ret, ok := consState.(*tendermint.ConsensusState)
	if !ok {
		panic("oops")
	}
	return ret
}

func (v lcv) verifyNewStateUpdate(ctx sdk.Context) error {
	return nil
}

func (v lcv) getStateInfo(ctx sdk.Context) error {
	return nil
}

func (v lcv) verify(
	ctx sdk.Context,
	height uint64,
	lightClientState exported.ConsensusState,
	lightClientStateTendermint *tendermint.ConsensusState,
	rollappState *types.StateInfo, // needs to be the one for (h-1) where h is the light client one
) error {
	// We check if it matches
	raBD, err := rollappState.BlockDescriptor(height)
	if err != nil {
		return errorsmod.Wrap(err, "block descriptor")
	}
	rollappStateRoot := raBD.GetStateRoot()
	lightClientStateRoot := lightClientState.GetRoot().GetHash()
	if !bytes.Equal(rollappStateRoot, lightClientStateRoot) {
		return errors.New("bad") // TODO:
	}
	// We also need to check if it was signed by the sequencer

	return nil
}
