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

Whats the attack we are trying to stop?
A sequencer from saying he did X via the light client, when in reality he did Y

So what will be the basic idea here?

When a light client update comes in, it has a client id, and a header

The height on the header is the height of the rollapp


Links
update entry point https://github.com/cosmos/ibc-go/blob/edb6efaddf45053efb9cbd2ad1ceab64b22fb085/modules/core/02-client/keeper/client.go#L60
ibc go header https://github.com/cosmos/ibc-go/blob/v6.2.2/modules/light-clients/07-tendermint/types/update.go
tendermind header https://github.com/cosmos/ibc-go/blob/a89bb2f556f519a3bbc3c97e107e29f722d59155/modules/light-clients/07-tendermint/types/tendermint.pb.go#L185-L202
what the validators actually sign https://github.com/cometbft/cometbft/blob/v0.34.33/types/canonical.go#L54-L65
verify commit https://github.com/cometbft/cometbft/blob/v0.34.33/types/validator_set.go#L716-L743


Note, the root in the tendermint consensus state, is the app hash from the state after TXs from the PREVIOUS block
	https://github.com/cosmos/ibc-go/blob/edb6efaddf45053efb9cbd2ad1ceab64b22fb085/modules/light-clients/07-tendermint/types/header.go#L21
	https://github.com/dymensionxyz/cometbft/blob/a059b062dcfc719406354e0a80f5f6d3cf7401e1/proto/tendermint/types/types.proto#L76
which should be equal to the corresponding state info block descriptor 'StateRoot'
	https://github.com/dymensionxyz/dymint/blob/d9dec3e96bd058732186d80bcc3d01f489f71634/settlement/dymension/dymension.go#L429

How does the light client update work?
The update contains a trusted height and a trusted validators
If that height is available and trusted on the LC, check that the nextValidatorsHash on the trusted data matches (the hash) in the header
then do light.Verify
light.Verify does 2 cases. For h+1, or more than h+1:
	basically it makes sure that
		- the trusting period not passed
		- header timestamp is inside trusting period
		- header timestamp is progressing forward from the last trusted one
		- make sure trust level proportion of trusted validators signed new commit (?)
	if h+1, requires the untrusted validators hash to be exactly equal to the trusted nextValidatorsHash
	otherwise;
		Trust level is default 1/3
		If (+)1/3 of trusted guys signed the new header AND 2/3 of new validators signed the new header


How would the system work?
	On receiving a new light client update, check if there is a state on the hub
	On receiving a new state update, check if there is a corresponding light client state
	If there is, verify

TODO: ned to make sure pruning handled in case that the state update arrives much later after the LC update
https://github.com/cosmos/ibc-go/blob/edb6efaddf45053efb9cbd2ad1ceab64b22fb085/modules/light-clients/07-tendermint/types/update.go#L112-L119
*/

/*
Friday afternoon notes:
I was thinking, that, the initial light client creation includes a next validators hash, but it does not check that it was signed by it
This is immediately followed up by a light client update, where a 'signed header' is created, which uses the trusted cons state

Let's walk through what happens
You get a 'header' which is completely untrusted
You read the previous trusted state from the 'trusted height' on the untrusted header

Then you have 8 arguments

these map to the following in verifyNonAdjacent
func VerifyNonAdjacent(
    trustedHeader *types.SignedHeader, = 'signed header =  the last trusted header (which was not actually signed if from MsgCreateClient)
	trustedVals *types.ValidatorSet, = tm trusted validators = the set of validators that the new header says hashes to the trusted nextValidatorsHash
	untrustedHeader *types.SignedHeader, = tm signed header = the new header
	untrustedVals *types.ValidatorSet, = tm validator set = the new validator set
	trustingPeriod time.Duration,
	now time.Time,
	maxClockDrift time.Duration,
	trustLevel cmtmath.Fraction
)

first we check that the trustedVals is actually the trustedVals
then we check that the next header is h+1, and the trusting period didnt expire
then we check that the SIGNED untrusted header .ValidatorsHash matches the validator set on the header
then we make sure that at least (1) of the trusted validators indeed 'committed' to the untrusted header
then we make sure that at least (1) of the new untrusted validators indeed 'committed' to the untrusted header

so, questions
a) where do we check that the hash of the trustedVals actually hashes to trustedHeader.nextvalidatorshash?
	yes, this is checked
b) what do they actually sign?
	they seem to sign the block id, which contains the hash of the block
c) how can we know that LC header H was signed by some sequencer signature?
	facts:
		1) createClient only contains the nextValidatorsHash
		2) a light client header H has an AppHash which is the stateinfo.root of H-1
	logic:
		a) if the light client accepted header h, then the supplied trusted validators indeed matched the stored nextvalidatorsHash from h-1 AND one of those trusted validators signed the new header
		b) suppose someone creates a light client with nextValidatorsHash = the sequencer, then they cannot switch to another one without the sequencers signature
		c) someone may create a light client with nextValidatorsHash = the seqeuencer without the sequencers signature, but then they will not be able to update it because they will need the signature

   so we don't need to hook anything to explicity look at signatures! it is enough for us to just check the light client nextValidatorsHash

basically for a state root h, we check the light client header root of h+1. The corresponding consensus state will have a 'next validators hash' and this is the

so basically each incoming header has three things
1. trusted validators = must hash to the trusted stored NextValidatorsHash. Note, it's a bad name, because it's really the current validators
2. validators, they must have signed the header, must be equal to
3. nextValidatorsHash, this is signed

the stored cons state has
1. app hash for height - 1
2. nextValidatorsHash

Let me walk through
at first all you have is a stored nextValidatorsHash
then the first update comes in
	it must have a trusted val set which hashes to a stored nextValidatorsHash
	this trusted val set must overlap with the signer of the update

attack (charlie = bad):
	create light client with nextValidatorsHash from Charlies private key
	update light client with nextValidatorsHash = sequencer, but sign this with Charlies pk
	thus: the stored nextValidatorsHash = sequencer, which is wrong

	but, charlie will not be able to do it again
	so you can mitigate by never checking the latest consensus state, but always the one before
		if you have a consensus state for height H, then the app hash inside it was signed by

	if you have a consensus state for H+1, the you have the


Q1) How do we know the header came from the sequencer?
	Given a header at height H, if the last header nextValidatorsHash is the sequencer
Q2) Say we suddenly have a root mismatch. At header height H and state height H-1. How do we find out who to blame?
	If we get the previous consensus state (via GetPreviousConsensusState), and we assume validator set is always size 1
	then we know he signed it


How does it look?
	New header arrives (H):
		We check that the state root = state root of stateInfo for H-1
		If match: no problem, someone might be busting the light client, but it doesnt matter
		If no match:
			Punish the
		Now we check that the PREVIOUS header nextValsetHash = sequencer

Friday afternoon.

TODO: need to think through protecting against larger valsets
TODO: need to tthink about sequencer change
TODO: need to think about non adjacency
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
