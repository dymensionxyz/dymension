package lightclient

import (
	"bytes"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	seqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type RollappKeeper interface {
	FindStateInfoByHeight(ctx sdk.Context, rollappId string, height uint64) (*types.StateInfo, error)
	FindBlockDescriptorByHeight(ctx sdk.Context, rollappId string, height uint64) (rollapptypes.BlockDescriptor, error)
	GetRollapp(ctx sdk.Context, rollappId string) (types.Rollapp, bool)
}

type SequencerKeeper interface {
	GetSequencer(ctx sdk.Context, sequencerAddress string) (val seqtypes.Sequencer, found bool)
}

type IBCKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool)
}

// UpdateBlockerDecorator intercepts incoming ibc CreateClient and UpdateClient messages and only allow them to proceed
// under conditions of the canonical light client ADR https://www.notion.so/dymension/ADR-x-Canonical-Light-Client-ccecd0907a8c40289f0c3339c8655dbd
type UpdateBlockerDecorator struct {
	raK  RollappKeeper
	seqK SequencerKeeper
	ibcK IBCKeeper
}

func NewUpdateBlockerDecorator(
	raK RollappKeeper,
	seqK SequencerKeeper,
	ibcK IBCKeeper,
) *UpdateBlockerDecorator {
	return &UpdateBlockerDecorator{raK: raK, seqK: seqK, ibcK: ibcK}
}

// verifyCreateClient checks that args <h, nextValidatorsHash, root> matches the corresponding existing state update root for h and validator for h+1
// it also verifies that the params such as trustingPeriod are acceptable by the rollapp.
// TODO: need to check timestamp to? probably not because it would just mean the sequencer is wrong
func (d UpdateBlockerDecorator) verifyCreateClient(ctx sdk.Context, msg *clienttypes.MsgCreateClient) error {
	clientStateI, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		// the ibc stack will take care of this
		return nil
	}

	consStateI, err := clienttypes.UnpackConsensusState(msg.ConsensusState)
	if err != nil {
		// the ibc stack will take care of this
		return nil
	}

	clientState, ok := clientStateI.(*ibctmtypes.ClientState)
	if !ok {
		// the ibc stack will take care of this, and we only care about tendermint
		return nil
	}

	consState, ok := consStateI.(*ibctmtypes.ConsensusState)
	if !ok {
		// the ibc stack will take care of this, and we only care about tendermint
		return nil
	}

	chainID := clientState.GetChainID()
	root := consState.GetRoot()
	rootHeight := clientState.GetLatestHeight().GetRevisionHeight() // TODO: check revision number?
	nextValidatorsHash := consState.NextValidatorsHash

	ra, ok := d.raK.GetRollapp(ctx, chainID)
	if !ok {
		// not relevant
		return nil
	}

	bd, err := d.raK.FindBlockDescriptorByHeight(ctx, ra.RollappId, rootHeight)
	if err != nil {
		return errorsmod.Wrapf(err, "find block descriptor by height: %d", rootHeight)
	}
	if !bytes.Equal(bd.GetStateRoot(), root.GetHash()) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "supplied root hash does not match block descriptor root hash")
	}
	nextState, err := d.raK.FindStateInfoByHeight(ctx, ra.RollappId, rootHeight+1)
	if err != nil {
		return errorsmod.Wrapf(err, "find state info by height: %d", rootHeight+1)
	}
	sequencer, ok := d.seqK.GetSequencer(ctx, nextState.GetSequencer())
	if !ok {
		return errorsmod.Wrapf(err, "get sequencer: %s", nextState.GetSequencer())
	}
	sequencerHash, err := sequencer.GetDymintPubKeyHash()
	if err != nil {
		return errorsmod.Wrap(err, "get sequencer public key hash")
	}

	if !bytes.Equal(sequencerHash, nextValidatorsHash) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "supplied next validators hash does not match actual sequencer for next height")
	}

	if err := d.validateClientStateParams(ctx, ra.RollappId, clientState); err != nil {
		return errorsmod.Wrap(err, "validate client state params")
	}

	return nil
}

// validateClientStateParams validates the 'params' on the client state.
func (d UpdateBlockerDecorator) validateClientStateParams(ctx sdk.Context, rollappID string, cs *ibctmtypes.ClientState) error {
	// TODO: check trust level etc
	return nil
}

// verifyUpdateClient only returns true if the rollapp has not yet submitted a height to the rollapp keeper, or
// they have and the root matches
func (d UpdateBlockerDecorator) verifyUpdateClient(ctx sdk.Context, msg *clienttypes.MsgUpdateClient) error {
	clientStateI, ok := d.ibcK.GetClientState(ctx, msg.ClientId)
	if !ok {
		// the ibc stack will take care of this, and we only care about tendermint
		return nil
	}

	clientState, ok := clientStateI.(*ibctmtypes.ClientState)
	if !ok {
		// the ibc stack will take care of this, and we only care about tendermint
		return nil
	}

	headerI, err := clienttypes.UnpackHeader(msg.Header)
	if err != nil {
		// the ibc stack will take care of this
		return nil
	}
	header, ok := headerI.(*ibctmtypes.Header)
	if !ok {
		// the ibc stack will take care of this, and we only care about tendermint
		return nil
	}

	ra, ok := d.raK.GetRollapp(ctx, clientState.GetChainID())
	if !ok {
		// not relevant
		return nil
	}

	rootHeight := header.GetHeight().GetRevisionHeight() // TODO: check revision number?
	root := commitmenttypes.NewMerkleRoot(header.GetHeader().GetAppHash())

	bd, err := d.raK.FindBlockDescriptorByHeight(ctx, ra.RollappId, rootHeight)
	if err != nil {
		return errorsmod.Wrapf(err, "find block descriptor by height: %d", rootHeight)
	}
	if !bytes.Equal(bd.GetStateRoot(), root.GetHash()) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "supplied root hash does not match block descriptor root hash")
	}

	return nil
}

// AnteHandle will return an error if the tx contains an ibc transfer message to a rollapp that has not finished the transfer genesis protocol.
func (d UpdateBlockerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		typeURL := sdk.MsgTypeURL(msg)
		if typeURL == sdk.MsgTypeURL(&clienttypes.MsgCreateClient{}) {
			m, ok := msg.(*clienttypes.MsgCreateClient)
			if !ok {
				return ctx, errorsmod.WithType(gerrc.ErrUnknown, msg)
			}
			err := d.verifyCreateClient(ctx, m)
			if err != nil {
				return ctx, errorsmod.Wrap(err, "light client: allow create client")
			}
		}
		if typeURL == sdk.MsgTypeURL(&clienttypes.MsgUpdateClient{}) {
			m, ok := msg.(*clienttypes.MsgUpdateClient)
			if !ok {
				return ctx, errorsmod.WithType(gerrc.ErrUnknown, msg)
			}
			err := d.verifyUpdateClient(ctx, m)
			if err != nil {
				return ctx, errorsmod.Wrap(err, "light client: allow update client")
			}
		}
	}

	return next(ctx, tx, simulate)
}
