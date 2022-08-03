package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dymensionxyz/dymension/x/sequencer/types"
)

const TypeMsgCreateRollapp = "create_rollapp"

var _ sdk.Msg = &MsgCreateRollapp{}

func NewMsgCreateRollapp(creator string, rollappId string, codeStamp string, genesisPath string, maxWithholdingBlocks uint64, maxSequencers uint64, permissionedAddresses *types.Sequencers) *MsgCreateRollapp {
	return &MsgCreateRollapp{
		Creator:               creator,
		RollappId:             rollappId,
		CodeStamp:             codeStamp,
		GenesisPath:           genesisPath,
		MaxWithholdingBlocks:  maxWithholdingBlocks,
		MaxSequencers:         maxSequencers,
		PermissionedAddresses: *permissionedAddresses,
	}
}

func (msg *MsgCreateRollapp) Route() string {
	return RouterKey
}

func (msg *MsgCreateRollapp) Type() string {
	return TypeMsgCreateRollapp
}

func (msg *MsgCreateRollapp) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateRollapp) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateRollapp) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
