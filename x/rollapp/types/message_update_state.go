package types

import (
	"math"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgUpdateState = "update_state"

var _ sdk.Msg = &MsgUpdateState{}

func NewMsgUpdateState(creator, rollappId, dAPath string, startHeight, numBlocks uint64, bDs *BlockDescriptors) *MsgUpdateState {
	return &MsgUpdateState{
		Creator:     creator,
		RollappId:   rollappId,
		StartHeight: startHeight,
		NumBlocks:   numBlocks,
		DAPath:      dAPath,
		BDs:         *bDs,
	}
}

func (msg *MsgUpdateState) Route() string {
	return RouterKey
}

func (msg *MsgUpdateState) Type() string {
	return TypeMsgUpdateState
}

func (msg *MsgUpdateState) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateState) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateState) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// an update can't be with no BDs
	if msg.NumBlocks == uint64(0) {
		return errorsmod.Wrap(ErrInvalidNumBlocks, "number of blocks can not be zero")
	}

	if msg.NumBlocks > math.MaxUint64-msg.StartHeight {
		return errorsmod.Wrapf(ErrInvalidNumBlocks, "numBlocks(%d) + startHeight(%d) exceeds max uint64", msg.NumBlocks, msg.StartHeight)
	}

	// check to see that update contains all BDs
	if len(msg.BDs.BD) != int(msg.NumBlocks) {
		return errorsmod.Wrapf(ErrInvalidNumBlocks, "number of blocks (%d) != number of block descriptors(%d)", msg.NumBlocks, len(msg.BDs.BD))
	}

	// check to see that startHeight is not zaro
	if msg.StartHeight == 0 {
		return errorsmod.Wrapf(ErrWrongBlockHeight, "StartHeight must be greater than zero")
	}

	// TODO: add a validation for DrsVersion once empty DRS version is marked vulnerable
	// if msg.DrsVersion == "" { return gerrc.ErrInvalidArgument.Wrap("DRS version must not be empty") }

	// check that the blocks are sequential by height
	for bdIndex := uint64(0); bdIndex < msg.NumBlocks; bdIndex += 1 {
		if msg.BDs.BD[bdIndex].Height != msg.StartHeight+bdIndex {
			return ErrInvalidBlockSequence
		}
		// check to see stateRoot is a 32 byte array
		if len(msg.BDs.BD[bdIndex].StateRoot) != 32 {
			return errorsmod.Wrapf(ErrInvalidStateRoot, "StateRoot of block high (%d) must be 32 byte array. But received (%d) bytes",
				msg.BDs.BD[bdIndex].Height, len(msg.BDs.BD[bdIndex].StateRoot))
		}
	}

	return nil
}
