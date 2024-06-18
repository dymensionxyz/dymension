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
