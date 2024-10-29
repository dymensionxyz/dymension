package ante

import (
	"bytes"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (i IBCMessagesDecorator) HandleMsgUpdateClient(ctx sdk.Context, msg *ibcclienttypes.MsgUpdateClient) error {
	chainID, ok := i.getChainID(ctx, msg)
	if !ok {
		// not relevant
		return nil
	}
	header, err := i.getHeader(ctx, msg)
	canonical := i.isCanonical(ctx, msg, chainID)
	if !canonical && errorsmod.IsOf(err, errIsMisbehaviour) {
		// We don't want to block misbehavior submission for non rollapps
		return nil
	}
	if err != nil {
		return err
	}
	seq, err := i.getSequencer(ctx, header)
	if errorsmod.IsOf(err, errProposerMismatch) {
		// this should not occur on any chain, regardless of being a rollapp or not
		return err
	}
	if errorsmod.IsOf(err, gerrc.ErrNotFound) {
		if !canonical {
			// not from sequencer, and not canonical - it's not interesting
			return nil
		}
		return gerrc.ErrInvalidArgument.Wrap("update canonical client with non sequencer header")
	}
	if err != nil {
		return err
	}
}

var (
	errIsMisbehaviour   = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "misbehavior evidence is disabled for canonical clients")
	errProposerMismatch = errorsmod.Wrap(gerrc.ErrInvalidArgument, "validator set proposer not equal header proposer field")
)

func (i IBCMessagesDecorator) getSequencer(ctx sdk.Context, header *ibctm.Header) (sequencertypes.Sequencer, error) {
	proposerBySignature := header.ValidatorSet.Proposer.GetAddress() // TODO: does ibc already guarantee this equal to header.ProposerAddr?
	proposerByData := header.Header.ProposerAddress
	if !bytes.Equal(proposerBySignature, proposerByData) {
		return sequencertypes.Sequencer{}, errProposerMismatch
	}
	addr := sdk.AccAddress(proposerByData).String()
	return i.rollappKeeper.GetRealSequencer(ctx, addr)
}

func (i IBCMessagesDecorator) isCanonical(ctx sdk.Context, msg *ibcclienttypes.MsgUpdateClient, chainID string) bool {
	canonicalID, _ := i.lightClientKeeper.GetCanonicalClient(ctx, chainID)
	return msg.ClientId == canonicalID
}

func (i IBCMessagesDecorator) getHeader(ctx sdk.Context, msg *ibcclienttypes.MsgUpdateClient) (*ibctm.Header, error) {
	clientMessage, err := ibcclienttypes.UnpackClientMessage(msg.ClientMessage)
	if err != nil {
		return nil, err
	}
	_, ok := clientMessage.(*ibctm.Misbehaviour)
	if ok {
		return nil, errIsMisbehaviour
	}
	header, ok := clientMessage.(*ibctm.Header)
	if !ok {
		return nil, nil
	}
	return header, nil
}

func (i IBCMessagesDecorator) getChainID(ctx sdk.Context, msg *ibcclienttypes.MsgUpdateClient) (string, bool) {
	clientState, ok := i.ibcClientKeeper.GetClientState(ctx, msg.ClientId)
	if !ok {
		return "", false
	}
	tmClientState, ok := clientState.(*ibctm.ClientState)
	if !ok {
		return "", false
	}
	return tmClientState.ChainId, true
}

func (i IBCMessagesDecorator) HandleMsgUpdateClientLegacy(ctx sdk.Context, msg *ibcclienttypes.MsgUpdateClient) error {
	clientState, found := i.ibcClientKeeper.GetClientState(ctx, msg.ClientId)
	if !found {
		return nil
	}
	// Cast client state to tendermint client state - we need this to get the chain id(rollapp id)
	tmClientState, ok := clientState.(*ibctm.ClientState)
	if !ok {
		return nil
	}
	// Check if the client is the canonical client for the rollapp
	rollappID := tmClientState.ChainId
	canonicalClient, _ := i.lightClientKeeper.GetCanonicalClient(ctx, rollappID)
	if canonicalClient != msg.ClientId {
		return nil // The client is not a rollapp's canonical client. Continue with default behaviour.
	}

	clientMessage, err := ibcclienttypes.UnpackClientMessage(msg.ClientMessage)
	if err != nil {
		return nil
	}
	_, ok = clientMessage.(*ibctm.Misbehaviour)
	if ok {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "misbehavior evidence is disabled for canonical clients")
	}
	header, ok := clientMessage.(*ibctm.Header)
	if !ok {
		return nil
	}

	h := header.GetHeight().GetRevisionHeight()
	s0, s1, err := i.getStateInfos(ctx, rollappID, h)
	if err != nil {
		return errorsmod.Wrap(err, "get state infos")
	}
	bd, _ := s0.GetBlockDescriptor(h)
	sequencerPubKey, err := i.lightClientKeeper.GetSequencerPubKey(ctx, s1.Sequencer)
	if err != nil {
		return err
	}
	rollappState := types.RollappState{
		BlockDescriptor:    bd,
		NextBlockSequencer: sequencerPubKey,
	}
	// Ensure that the ibc header is compatible with the existing rollapp state
	// If it's not, we error and prevent the MsgUpdateClient from being processed
	err = types.CheckCompatibility(*header.ConsensusState(), rollappState)
	if err != nil {
		return err
	}

	return nil
}

// getStateInfos gets state infos for h and h+1
func (i IBCMessagesDecorator) getStateInfos(ctx sdk.Context, rollapp string, h uint64) (*rollapptypes.StateInfo, *rollapptypes.StateInfo, error) {
	// Check if there are existing block descriptors for the given height of client state
	s0, err := i.rollappKeeper.FindStateInfoByHeight(ctx, rollapp, h)
	if errorsmod.IsOf(err, gerrc.ErrNotFound) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	s1 := s0
	if !s1.ContainsHeight(h) {
		s1, err = i.rollappKeeper.FindStateInfoByHeight(ctx, rollapp, h+1)
		if errorsmod.IsOf(err, gerrc.ErrNotFound) {
			return nil, nil, nil
		}
		if err != nil {
			return nil, nil, err
		}
	}
	return s0, s1, nil
}

func (i IBCMessagesDecorator) acceptUpdateOptimistically(ctx sdk.Context, clientID string, header *ibctm.Header) {
	i.lightClientKeeper.SetConsensusStateValHash(ctx, clientID, uint64(header.Header.Height), header.Header.ValidatorsHash)
}
