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
- avoid requiring a consensus state for every height

A note on headers:
	In a header with height H:
		- The AppHash is for height H-1
		- The ValidatorsHash is for height H
		- The NextValidatorsHash is for height H+1
	Crucially, that means when the APP returns the 'next validator set' S to ABCI at height H:
		- Header H+1 contains the Root of H
		- Header H+1 contains the NextValidatorsHash (hash of S)
		- Header H+2 contains ValidatorsHash (hash of S)
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
		(Actually: just need someone to blame if the light client diverges from the state root.
		 That means we need to be sure, that, for a given light client consensus state, if it's wrong, then
         we want to be sure that it was actually the seqeuencer who produced it)

How to compare state root with light client?
	On update client:
		Data:
			height H
			nextValidatorsHash X
			root R
		Steps:
			Find the block descriptor (app hash) of height H-1 using binary search
			If it exists:
				compare root with app hash
	On update rollapp:
		For each block descriptor (app hash and height):
			find cons state for height h+1. It contains the sequencer bech32 addr.
			If it exists:
				compare app hash with root

How to make sure the light client is updated by the sequencer only?
	First, note that for consensus states
			#1: (Height h1, Root h1-1, NextValidatorsHash A)
			#2: (Height h2, Root h2-1, NextValidatorsHash B)
		then we know at least one of the validators in A signed #2, but we don't yet know that anyone in B is guilty of anything at all
		This only holds if #1 is not the original unvalidated consensus state from MsgCreateClient

How do we know who to blame, if we get a state root mismatch?
	Suppose we get  a block descriptor app hash H that doesn't equal (!=) light client app hash H+1
	Then, we know that the sequencer in the predecessor consensus state to H+1 signed block H

Sequencer rotation trust
	If the sequencer changes, we assume he's trustworthy for T time
	If the sequencer ch

Attack idea
	h :     cons state #1 sequencer A (bad guy)
	h + k1: cons state #2 sequencer B (good guy)
    h + k2: sequencer A creates a wrong header with himself as the trusted val set. It's accepted because of trust period from h.

Pseudocode for root matching:
	EndBlock:
		highestChecked
		for blockDescriptor in uncheckedBlockDescriptors ascending by height:
			h = block_descriptor.height - 1
			consState = ibc.GetNextConsensusState(h)


Design:
	There are two cases when a light client update arrives
	1. The state update already exists
	2. The state update does not exist
	Case 1 is trivial: only allow the client update if the root agrees with the state update.
	Case 2:
		We accept the light client updates optimistically.
		When the state update arrives, we compare the roots.
		If there is a mismatch, we cannot immediately blame the sequencer of the height, due to Attack (see below).
		Solution:
			- Ensure all sequencers are still slashable while they are within the light client trusting period. (We should already have this).
			- For each light client update, we save the 'trustedHeight' argument. When there is a root dispute, we can know which sequencer created
              the light client update by looking at the sequencer at the trustedHeight.

Attack:
	To submit a light client update, you just need to provide a trusted height and 'trusted' validator set.
	The ibc module on chain will check that the nextValidatorsHash at the trustedHeight hashes to the trusted validator set. This validates the trusted validator set.
	Then it will check that +1/3 of the trusted validator actually signed the header, and that the trusted validator set is still within the trusting period.
	In this way, it is guaranteed that only the current and recent sequencers can create light client updates.
	That means, in a rotating sequencer system, we cannot (without more work) blame the seqeuencer at height H for a wrong light client root at height H, because
	it may have been created by a different (but recent) sequencer.


	If the state root already exists, we only allow the UpdateClient if it agrees.
	So suppose that it does not yet exist. Then we accept optimistically. At some point the state root will arrive.
	We will compare the roots in ascending order.
	Take the first mismatch.
	How do we know it came from the sequencer?
	What if we stored the trusted height from every header?
	Then, know it was signed by the sequencer from that height, so we can slash them.




	If we can check the sequencer is the same before we accept the LC update
		then we can rollack to the previous state info and slash
	Then we CAN rollback to the previous state info




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
