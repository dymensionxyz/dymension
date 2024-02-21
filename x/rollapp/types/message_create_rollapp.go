package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCreateRollapp = "create_rollapp"

var _ sdk.Msg = &MsgCreateRollapp{}

func NewMsgCreateRollapp(creator string, rollappId string, maxSequencers uint64, permissionedAddresses []string, metadatas []TokenMetadata) *MsgCreateRollapp {
	return &MsgCreateRollapp{
		Creator:               creator,
		RollappId:             rollappId,
		MaxSequencers:         maxSequencers,
		PermissionedAddresses: permissionedAddresses,
		Metadatas:             metadatas,
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
	if msg.GetMaxSequencers() == 0 {
		return sdkerrors.Wrap(ErrInvalidMaxSequencers, "max-sequencers must be greater than 0")
	}

	// verifies that there's no duplicate address in PermissionedAddresses
	// and addresses are in Bech32 format
	permissionedAddresses := msg.GetPermissionedAddresses()
	if len(permissionedAddresses) > 0 {
		duplicateAddresses := make(map[string]bool)
		for _, item := range permissionedAddresses {
			// check if the item/element exist in the duplicateAddresses map
			_, exist := duplicateAddresses[item]
			if exist {
				return sdkerrors.Wrapf(ErrPermissionedAddressesDuplicate, "address: %s", item)
			}
			// check Bech32 format
			if _, err := sdk.AccAddressFromBech32(item); err != nil {
				return sdkerrors.Wrapf(ErrInvalidPermissionedAddress, "invalid permissioned address: %s", err)
			}
			// mark as exist
			duplicateAddresses[item] = true
		}
	}

	// verifies that token metadata, if any, must be valid
	if len(msg.GetMetadatas()) > 0 {
		for _, metadata := range msg.GetMetadatas() {
			if err := metadata.Validate(); err != nil {
				return sdkerrors.Wrapf(ErrInvalidTokenMetadata, "%s: %v", metadata.Base, err)
			}
		}
	}

	return nil
}
