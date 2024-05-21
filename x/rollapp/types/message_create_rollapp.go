package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgCreateRollapp = "create_rollapp"

var _ sdk.Msg = &MsgCreateRollapp{}

const MaxAllowedSequencers = 100

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

func (msg *MsgCreateRollapp) GetRollapp() Rollapp {
	// Build the genesis state from the genesis accounts
	rollappGenesisState := RollappGenesisState{}
	rollappGenesisState.GenesisAccounts = make([]*GenesisAccount, len(msg.GenesisAccounts))
	for i := range msg.GenesisAccounts {
		rollappGenesisState.GenesisAccounts[i] = &msg.GenesisAccounts[i]
	}
	metadata := make([]*TokenMetadata, len(msg.Metadatas))
	for i := range msg.Metadatas {
		metadata[i] = &msg.Metadatas[i]
	}

	return NewRollapp(msg.Creator, msg.RollappId, msg.MaxSequencers, msg.PermissionedAddresses, metadata, rollappGenesisState)
}

func (msg *MsgCreateRollapp) ValidateBasic() error {
	rollapp := msg.GetRollapp()
	if err := rollapp.ValidateBasic(); err != nil {
		return err
	}

	return nil
}
