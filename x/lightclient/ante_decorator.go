package lightclient

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type RollappKeeper interface {
	FindBlockDescriptorByHeight(ctx sdk.Context, rollappId string, height uint64) (rollapptypes.BlockDescriptor, error) {
}

type UpdateBlockerDecorator struct{}

func NewUpdateBlockerDecorator() *UpdateBlockerDecorator {
	return &UpdateBlockerDecorator{}
}

func (h UpdateBlockerDecorator) allowCreateClient(ctx sdk.Context, msg *clienttypes.MsgCreateClient) (bool, error) {
	consensusState, err := clienttypes.UnpackConsensusState(msg.ConsensusState)
	if err != nil {
		return nil, err
	}
}

// allowUpdateClient only returns true if the rollapp has not yet submitted a height to the rollapp keeper, or
// they have and the root matches
func (h UpdateBlockerDecorator) allowUpdateClient(ctx sdk.Context, msg *clienttypes.MsgUpdateClient) (bool, error) {
	header, err := clienttypes.UnpackHeader(msg.Header)
	if err != nil {
		// the ibc stack will take care of this
		return true, nil
	}
	tmHeader, ok := header.(*ibctmtypes.Header)
	if !ok {
		// we only care about tendermint
		return true, nil
	}
	height := tmHeader.GetHeight()
	root := commitmenttypes.NewMerkleRoot(tmHeader.GetHeader().GetAppHash())
}

// AnteHandle will return an error if the tx contains an ibc transfer message to a rollapp that has not finished the transfer genesis protocol.
func (h UpdateBlockerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		typeURL := sdk.MsgTypeURL(msg)
		if typeURL == sdk.MsgTypeURL(&clienttypes.MsgCreateClient{}) {
			m, ok := msg.(*clienttypes.MsgCreateClient)
			if !ok {
				return ctx, errorsmod.WithType(gerrc.ErrUnknown, msg)
			}
			ok, err := h.allowCreateClient(ctx, m)
			if err != nil {
				return ctx, errorsmod.Wrap(err, "light client: allow create client")
			}
			if !ok {
				return ctx, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "create client disabled")
			}
		}
		if typeURL == sdk.MsgTypeURL(&clienttypes.MsgUpdateClient{}) {
			m, ok := msg.(*clienttypes.MsgUpdateClient)
			if !ok {
				return ctx, errorsmod.WithType(gerrc.ErrUnknown, msg)
			}
			ok, err := h.allowUpdateClient(ctx, m)
			if err != nil {
				return ctx, errorsmod.Wrap(err, "light client: allow update client")
			}
			if !ok {
				return ctx, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "update client disabled")
			}
		}
	}

	return next(ctx, tx, simulate)
}
