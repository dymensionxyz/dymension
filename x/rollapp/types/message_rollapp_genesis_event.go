package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (msg *MsgRollappGenesisEvent) GetSigners() []sdk.AccAddress {
	return nil
}

func (msg *MsgRollappGenesisEvent) ValidateBasic() error {
	return sdkerrors.ErrNotSupported
}
