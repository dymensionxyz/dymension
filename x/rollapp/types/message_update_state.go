package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateState = "update_state"

var _ sdk.Msg = &MsgUpdateState{}

func NewMsgUpdateState(creator string, rollappId string, startHeight uint64, numBlocks uint64, dAPath string, version uint64, lastBD *BlockDescriptor, bDs *BlockDescriptors) *MsgUpdateState {
	return &MsgUpdateState{
		Creator:     creator,
		RollappId:   rollappId,
		StartHeight: startHeight,
		NumBlocks:   numBlocks,
		DAPath:      dAPath,
		Version:     version,
		LastBD:      *lastBD,
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// an update cann't be with no BDs
	if msg.NumBlocks == uint64(0) {
		return sdkerrors.Wrap(ErrInvalidNumBlocks, "number of blocks can not be zero")
	}

	// check to see that update contains all BDs
	if len(msg.BDs.BD) != int(msg.NumBlocks) {
		return sdkerrors.Wrapf(ErrInvalidNumBlocks, "number of blocks (%d) != number of block descriptors(%d)", msg.NumBlocks, len(msg.BDs.BD))
	}

	// check if it's not the first state update for this rollapp
	if msg.LastBD.Height == 0 {
		// check to see if this update starts from the first block height
		if msg.BDs.BD[0].Height != 0 {
			return sdkerrors.Wrapf(ErrInvalidBlockSequence, "new updated height (%d), but the last reported height indicates it's the first update", msg.BDs.BD[0].Height)
		}
	} else {
		// check to see if this update provides the latest known state
		// and that it updates start from that height
		if msg.BDs.BD[0].Height != msg.LastBD.Height+1 {
			return sdkerrors.Wrapf(ErrInvalidBlockSequence, "new updated height (%d) is not following the last reported height (%d)", msg.BDs.BD[0].Height, msg.LastBD.Height+1)
		}
	}

	// check that the blocks are sequential by height
	for bdIndex := uint64(0); bdIndex < msg.NumBlocks; bdIndex += 1 {
		if msg.BDs.BD[bdIndex].Height != msg.StartHeight+bdIndex {
			return ErrInvalidBlockSequence
		}
	}

	return nil
}
