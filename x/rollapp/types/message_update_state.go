package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateState = "update_state"

var _ sdk.Msg = &MsgUpdateState{}

func NewMsgUpdateState(creator string, rollappId string, startHeight uint64, numBlocks uint64, dAPath string, version uint64, bDs *BlockDescriptors) *MsgUpdateState {
	return &MsgUpdateState{
		Creator:     creator,
		RollappId:   rollappId,
		StartHeight: startHeight,
		NumBlocks:   numBlocks,
		DAPath:      dAPath,
		Version:     version,
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

	// check to see that startHeight is not zaro
	if msg.StartHeight == 0 {
		return sdkerrors.Wrapf(ErrWrongBlockHeight, "StartHeight must be greater than zero")
	}

	// check that the blocks are sequential by height
	for bdIndex := uint64(0); bdIndex < msg.NumBlocks; bdIndex += 1 {
		block := msg.BDs.BD[bdIndex]
		if block.Height != msg.StartHeight+bdIndex {
			return ErrInvalidBlockSequence
		}
		// check to see stateRoot is a 32 byte array
		if len(block.StateRoot) != 32 {
			return sdkerrors.Wrapf(ErrInvalidStateRoot, "StateRoot of block high (%d) must be 32 byte array. But received (%d) bytes",
				block.Height, len(block.StateRoot))
		}

		// check to see IntermediateStatesRoot is a 32 byte array
		if block.IntermediateStatesRoots != nil {
			for _, intermediateStatesRoot := range block.IntermediateStatesRoots {
				if len(intermediateStatesRoot) != 32 {
					return sdkerrors.Wrapf(ErrInvalidIntermediateStatesRoot, "IntermediateStatesRoot of block high (%d) must be 32 byte array. But received (%d) bytes",
						block.Height, len(intermediateStatesRoot))
				}
			}
		}
	}

	return nil
}
