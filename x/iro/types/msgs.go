package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgBuy{}
)

/*
   // ValidateBasic does a simple validation check that
   // doesn't require access to any other information.
   ValidateBasic() error

   // GetSigners returns the addrs of signers that must sign.
   // CONTRACT: All signatures must be present to be valid.
   // CONTRACT: Returns addrs in some deterministic order.
   GetSigners() []AccAddress
*/

func (m *MsgBuy) ValidateBasic() error {
	return nil
}

func (m *MsgBuy) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Buyer}
}
