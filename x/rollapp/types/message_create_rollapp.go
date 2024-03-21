package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgCreateRollapp = "create_rollapp"

var _ sdk.Msg = &MsgCreateRollapp{}

func NewMsgCreateRollapp(creator string, rollappId string, maxSequencers uint64, permissionedAddresses []string,
	metadatas []TokenMetadata, genesisAccounts []GenesisAccount,
) *MsgCreateRollapp {
	return &MsgCreateRollapp{
		Creator:               creator,
		RollappId:             rollappId,
		MaxSequencers:         maxSequencers,
		PermissionedAddresses: permissionedAddresses,
		Metadatas:             metadatas,
		GenesisAccounts:       genesisAccounts,
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
	// validate the basics fields
	baseRollapp := Rollapp{
		Creator:               msg.Creator,
		RollappId:             msg.RollappId,
		MaxSequencers:         msg.MaxSequencers,
		PermissionedAddresses: msg.PermissionedAddresses,
	}
	if err := baseRollapp.ValidateBasic(); err != nil {
		return err
	}

	// verifies that token metadata, if any, must be valid
	if len(msg.GetMetadatas()) > 0 {
		for _, metadata := range msg.GetMetadatas() {
			if err := metadata.Validate(); err != nil {
				return errorsmod.Wrapf(ErrInvalidTokenMetadata, "%s: %v", metadata.Base, err)
			}
		}
	}

	// genesisAccounts address validation
	for _, acc := range msg.GenesisAccounts {
		_, err := sdk.AccAddressFromBech32(acc.Address)
		if err != nil {
			return errorsmod.Wrap(err, ErrInvalidGenesisAccount.Error())
		}
	}

	return nil
}
