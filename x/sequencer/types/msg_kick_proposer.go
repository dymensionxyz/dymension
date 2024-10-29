package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgKickProposer{}

func NewMsgKickProposer(creator string) *MsgKickProposer {
	return &MsgKickProposer{
		Creator: creator,
	}
}

func (m *MsgKickProposer) ValidateBasic() error {
	// TODO implement me
	return nil //
}

func (m *MsgKickProposer) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}
