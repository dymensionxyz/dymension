package keeper

import (
	"bytes"
	"context"
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// wraps the normal ibc client keeper update client message but routes it through our ante
// Now we have two ways to update: direct through normal pathway or here, which is messy.
// We can improve in SDK v0.52+ with pre/post message hooks.
func (m msgServer) UpdateClient(goCtx context.Context, msg *clienttypes.MsgUpdateClient) (*clienttypes.MsgUpdateClientResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	d := PreUpdateClientHandler{
		IbcClientKeeper:  m.ibcClientKeeper,
		IbcChannelKeeper: m.ibcChannelK,
		RaK:              m.rollappKeeper,
		K:                *m.Keeper,
	}

	err := d.HandleMsgUpdateClient(ctx, msg)

	if err != nil {
		return nil, err
	}

	return m.ibcKeeper.UpdateClient(ctx, msg)
}

type PreUpdateClientHandler struct {
	IbcClientKeeper  types.IBCClientKeeperExpected
	IbcChannelKeeper types.IBCChannelKeeperExpected
	RaK              types.RollappKeeperExpected
	K                Keeper
}

var (
	errIsMisbehaviour   = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "misbehavior evidence is disabled for canonical clients")
	errNoHeader         = errors.New("message does not contain header")
	errProposerMismatch = errorsmod.Wrap(gerrc.ErrInvalidArgument, "validator set proposer not equal header proposer field")
)

func (i PreUpdateClientHandler) HandleMsgUpdateClient(ctx sdk.Context, msg *clienttypes.MsgUpdateClient) error {
	if !i.K.Enabled() {
		return nil
	}
	_, canonical := i.K.GetRollappForClientID(ctx, msg.ClientId)
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

	rollapp, ok := i.RaK.GetRollapp(ctx, seq.RollappId)
	if !ok {
		return gerrc.ErrInternal.Wrapf("get rollapp from sequencer: rollapp: %s", seq.RollappId)
	}

	// this disallows LC updates from previous revisions but should be fine since new state roots can be used to prove
	// state older than the one in the current state root.
	if header.Header.Version.App != rollapp.LatestRevision().Number {
		return errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "client update revision mismatch (expected: %d , actual: %d)", rollapp.LatestRevision().Number, header.Header.Version.App)
	}

	h := header.GetHeight().GetRevisionHeight()
	sInfo, err := i.RaK.FindStateInfoByHeight(ctx, rollapp.RollappId, h)
	if errorsmod.IsOf(err, gerrc.ErrNotFound) {
		// the header is optimistic: the state update has not yet been received, so we save optimistically
		err := i.K.SaveSigner(ctx, seq.Address, msg.ClientId, h)
		if err != nil {
			return errorsmod.Wrap(err, "save signer")
		}
		return nil
	}
	if err != nil {
		return errorsmod.Wrap(err, "find state info by height")
	}

	err = i.K.ValidateHeaderAgainstStateInfo(ctx, sInfo, header.ConsensusState(), h)
	if err != nil {
		return errorsmod.Wrap(err, "validate pessimistic")
	}

	return nil
}

func (i PreUpdateClientHandler) getSequencer(ctx sdk.Context, header *ibctm.Header) (sequencertypes.Sequencer, error) {
	proposerBySignature := header.ValidatorSet.Proposer.GetAddress()
	proposerByData := header.Header.ProposerAddress
	// Does ibc already guarantee this equal to header.ProposerAddr? I don't think so
	if !bytes.Equal(proposerBySignature, proposerByData) {
		return sequencertypes.Sequencer{}, errProposerMismatch
	}
	return i.K.SeqK.SequencerByDymintAddr(ctx, proposerByData)
}

func getHeader(msg *clienttypes.MsgUpdateClient) (*ibctm.Header, error) {
	clientMessage, err := clienttypes.UnpackClientMessage(msg.ClientMessage)
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
