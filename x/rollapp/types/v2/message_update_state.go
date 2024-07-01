package v2

import (
	"math"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

const TypeMsgUpdateState = "update_state"

var _ sdk.Msg = &MsgUpdateState{}

func NewMsgUpdateState(creator string, rollappId string, startHeight uint64, numBlocks uint64, dAPath *types.DAPath, version uint64, bDs *types.BlockDescriptors) *MsgUpdateState {
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
	return types.RouterKey
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
		return errorsmod.Wrapf(types.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// an update can't be with no BDs
	if msg.NumBlocks == uint64(0) {
		return errorsmod.Wrap(types.ErrInvalidNumBlocks, "number of blocks can not be zero")
	}

	if msg.NumBlocks > math.MaxUint64-msg.StartHeight {
		return errorsmod.Wrapf(types.ErrInvalidNumBlocks, "numBlocks(%d) + startHeight(%d) exceeds max uint64", msg.NumBlocks, msg.StartHeight)
	}

	// check to see that update contains all BDs
	if len(msg.BDs.BD) != int(msg.NumBlocks) {
		return errorsmod.Wrapf(types.ErrInvalidNumBlocks, "number of blocks (%d) != number of block descriptors(%d)", msg.NumBlocks, len(msg.BDs.BD))
	}

	// check to see that startHeight is not zaro
	if msg.StartHeight == 0 {
		return errorsmod.Wrapf(types.ErrWrongBlockHeight, "StartHeight must be greater than zero")
	}

	if msg.DAPath.DaType == "" {
		return errorsmod.Wrap(types.ErrInvalidDAClientType, "DAPath.DaType can't be empty")
	}

	// check that the blocks are sequential by height
	for bdIndex := uint64(0); bdIndex < msg.NumBlocks; bdIndex += 1 {
		if msg.BDs.BD[bdIndex].Height != msg.StartHeight+bdIndex {
			return types.ErrInvalidBlockSequence
		}
		// check to see stateRoot is a 32 byte array
		if len(msg.BDs.BD[bdIndex].StateRoot) != 32 {
			return errorsmod.Wrapf(types.ErrInvalidStateRoot, "StateRoot of block high (%d) must be 32 byte array. But received (%d) bytes",
				msg.BDs.BD[bdIndex].Height, len(msg.BDs.BD[bdIndex].StateRoot))
		}
	}

	return nil
}
