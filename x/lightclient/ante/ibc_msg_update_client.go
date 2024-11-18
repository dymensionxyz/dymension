package ante

import (
	"bytes"
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (i IBCMessagesDecorator) HandleMsgUpdateClient(ctx sdk.Context, msg *ibcclienttypes.MsgUpdateClient) error {
	if !i.k.Enabled() {
		return nil
	}
	_, canonical := i.k.GetRollappForClientID(ctx, msg.ClientId)
	header, err := getHeader(msg)
	if !canonical && errorsmod.IsOf(err, errIsMisbehaviour) {
		// We don't want to block misbehavior submission for non rollapps
		return nil
	}
	if errorsmod.IsOf(err, errNoHeader) {
		// it doesn't concern us
		return nil
	}
	if err != nil {
		return errorsmod.Wrap(err, "get header")
	}
	seq, err := i.getSequencer(ctx, header)
	err = errorsmod.Wrap(err, "get sequencer")
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

	// ~~~~~
	// now we know that the msg is a header, and it was produced by a sequencer
	// ~~~~~

	if !seq.Bonded() {
		// we assume here that sequencers will not propose blocks on other chains connected to the hub except for their rollapp
		return gerrc.ErrInvalidArgument.Wrap("header is from unbonded sequencer")
	}

	rollapp, ok := i.raK.GetRollapp(ctx, seq.RollappId)
	if !ok {
		return gerrc.ErrInternal.Wrapf("get rollapp from sequencer: rollapp: %s", seq.RollappId)
	}

	// cannot update the LC unless fork is resolved (after receiving state post fork state update)
	if i.k.IsHardForkingInProgress(ctx, rollapp.RollappId) {
		return types.ErrorHardForkInProgress
	}

	// this disallows LC updates from previous revisions but should be fine since new state roots can be used to prove
	// state older than the one in the current state root.
	if header.Header.Version.App != rollapp.LatestRevision().Number {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "client update revision mismatch")
	}

	h := header.GetHeight().GetRevisionHeight()
	stateInfos, err := i.getStateInfos(ctx, rollapp.RollappId, h)
	if err != nil {
		return errorsmod.Wrap(err, "get state infos")
	}

	if stateInfos.containingHPlus1 != nil {
		// the header is pessimistic: the state update has already been received, so we check the header doesn't mismatch
		return errorsmod.Wrap(i.validateUpdatePessimistically(ctx, stateInfos, header.ConsensusState(), h), "validate pessimistic")
	}

	// the header is optimistic: the state update has not yet been received, so we save optimistically
	return errorsmod.Wrap(i.k.SaveSigner(ctx, seq.Address, msg.ClientId, h), "save updater")
}

var (
	errIsMisbehaviour   = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "misbehavior evidence is disabled for canonical clients")
	errNoHeader         = errors.New("message does not contain header")
	errProposerMismatch = errorsmod.Wrap(gerrc.ErrInvalidArgument, "validator set proposer not equal header proposer field")
)

func (i IBCMessagesDecorator) getSequencer(ctx sdk.Context, header *ibctm.Header) (sequencertypes.Sequencer, error) {
	proposerBySignature := header.ValidatorSet.Proposer.GetAddress()
	proposerByData := header.Header.ProposerAddress
	// Does ibc already guarantee this equal to header.ProposerAddr? I don't think so
	if !bytes.Equal(proposerBySignature, proposerByData) {
		return sequencertypes.Sequencer{}, errProposerMismatch
	}
	return i.k.SeqK.SequencerByDymintAddr(ctx, proposerByData)
}

func getHeader(msg *ibcclienttypes.MsgUpdateClient) (*ibctm.Header, error) {
	clientMessage, err := ibcclienttypes.UnpackClientMessage(msg.ClientMessage)
	if err != nil {
		return nil, errorsmod.Wrap(err, "unpack client message")
	}
	_, ok := clientMessage.(*ibctm.Misbehaviour)
	if ok {
		return nil, errIsMisbehaviour
	}
	header, ok := clientMessage.(*ibctm.Header)
	if !ok {
		return nil, errNoHeader
	}
	return header, nil
}

// if containingHPlus1 is not nil then containingH also guaranteed to not be nil
type stateInfos struct {
	containingH      *rollapptypes.StateInfo
	containingHPlus1 *rollapptypes.StateInfo
}

// getStateInfos gets state infos for h and h+1
func (i IBCMessagesDecorator) getStateInfos(ctx sdk.Context, rollapp string, h uint64) (stateInfos, error) {
	// Check if there are existing block descriptors for the given height of client state
	s0, err := i.raK.FindStateInfoByHeight(ctx, rollapp, h)
	if errorsmod.IsOf(err, gerrc.ErrNotFound) {
		return stateInfos{}, nil
	}
	if err != nil {
		return stateInfos{}, err
	}
	s1 := s0
	if !s1.ContainsHeight(h + 1) {
		s1, err = i.raK.FindStateInfoByHeight(ctx, rollapp, h+1)
		if errorsmod.IsOf(err, gerrc.ErrNotFound) {
			return stateInfos{s0, nil}, nil
		}
		if err != nil {
			return stateInfos{}, err
		}
	}
	return stateInfos{s0, s1}, nil
}

func (i IBCMessagesDecorator) validateUpdatePessimistically(ctx sdk.Context, infos stateInfos, consState *ibctm.ConsensusState, h uint64) error {
	bd, _ := infos.containingH.GetBlockDescriptor(h)
	seq, err := i.k.SeqK.RealSequencer(ctx, infos.containingHPlus1.Sequencer)
	if err != nil {
		return errorsmod.Wrap(errors.Join(err, gerrc.ErrInternal), "get sequencer of state info")
	}
	rollappState := types.RollappState{
		BlockDescriptor:    bd,
		NextBlockSequencer: seq,
	}
	return errorsmod.Wrap(types.CheckCompatibility(*consState, rollappState), "check compatibility")
}
