package lightclient

import (
	"bytes"

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
	FindStateInfoByHeight(ctx sdk.Context, rollappId string, height uint64) (*types.StateInfo, error)
	FindBlockDescriptorByHeight(ctx sdk.Context, rollappId string, height uint64) (rollapptypes.BlockDescriptor, error)
	GetRollapp(ctx sdk.Context, rollappId string) (types.Rollapp, bool)
}

// UpdateBlockerDecorator intercepts incoming ibc CreateClient and UpdateClient messages and only allow them to proceed
// under conditions of the canonical light client ADR https://www.notion.so/dymension/ADR-x-Canonical-Light-Client-ccecd0907a8c40289f0c3339c8655dbd
type UpdateBlockerDecorator struct {
	rk RollappKeeper
}

func NewUpdateBlockerDecorator() *UpdateBlockerDecorator {
	return &UpdateBlockerDecorator{}
}

// TODO: need to check timestamp to? probably not because it would just mean the sequencer is wrong
func (d UpdateBlockerDecorator) allowCreateClient(ctx sdk.Context, msg *clienttypes.MsgCreateClient) (bool, error) {
	clientStateI, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		// the ibc stack will take care of this
		return true, nil
	}

	consStateI, err := clienttypes.UnpackConsensusState(msg.ConsensusState)
	if err != nil {
		// the ibc stack will take care of this
		return true, nil
	}

	clientState, ok := clientStateI.(*ibctmtypes.ClientState)
	if !ok {
		// the ibc stack will take care of this, and we only care about tendermint
		return true, nil
	}

	consState, ok := consStateI.(*ibctmtypes.ConsensusState)
	if !ok {
		// the ibc stack will take care of this, and we only care about tendermint
		return true, nil
	}

	chainID := clientState.GetChainID()
	root := consState.GetRoot()
	rootHeight := clientState.GetLatestHeight().GetRevisionHeight() // TODO: check revision number?
	nextValidatorsHash := consState.NextValidatorsHash

	ra, ok := d.rk.GetRollapp(ctx, chainID)
	if !ok {
		// not relevant
		return true, nil
	}

	bd, err := d.rk.FindBlockDescriptorByHeight(ctx, ra.RollappId, rootHeight)
	if err != nil  {
		return false, errorsmod.Wrap(err, "find block descriptor by height")
	}
	if !bytes.Equal(bd.GetStateRoot(), root.GetHash()){
		return false, nil
	}
	nextBD, err := d.rk.FindStateInfoByHeight(ctx, ra.RollappId, rootHeight+1)
	if err != nil {
		return false, errorsmod.Wrap(err, "find state info by height")
	}

	if bd.GetStateRoot().eq

}

func (d UpdateBlockerDecorator) allowClientStateParams(ctx sdk.Context, rollappID string, cs *ibctmtypes.ClientState) (bool, error) {
	// TODO: check trust level etc
	return true, nil
}

// allowUpdateClient only returns true if the rollapp has not yet submitted a height to the rollapp keeper, or
// they have and the root matches
func (d UpdateBlockerDecorator) allowUpdateClient(ctx sdk.Context, msg *clienttypes.MsgUpdateClient) (bool, error) {
	headerI, err := clienttypes.UnpackHeader(msg.Header)
	if err != nil {
		// the ibc stack will take care of this
		return true, nil
	}
	header, ok := headerI.(*ibctmtypes.Header)
	if !ok {
		// the ibc stack will take care of this, and we only care about tendermint
		return true, nil
	}
	height := header.GetHeight()
	root := commitmenttypes.NewMerkleRoot(header.GetHeader().GetAppHash())
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
			ok, err := d.allowCreateClient(ctx, m)
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
			ok, err := d.allowUpdateClient(ctx, m)
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
