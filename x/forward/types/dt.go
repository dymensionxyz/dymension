package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (h *HookEIBCtoHL) ValidateBasic() error {
	err := h.Recovery.ValidateBasic()
	if err != nil {
		return err
	}
	return nil
}

func (h *HookHLtoIBC) ValidateBasic() error {
	err := h.Transfer.ValidateBasic()
	if err != nil {
		return err
	}
	err = h.Recovery.ValidateBasic()
	if err != nil {
		return err
	}
	// TODO: can check timeout height is zero(?)

	return nil
}

func (r *Recovery) ValidateBasic() error {
	addr := r.GetAddress()
	if addr == "" {
		return errors.New("address is empty")
	}
	return nil
}

func (r *Recovery) AccAddr() string {
	// from bech32
	addr, err := sdk.AccAddressFromBech32(r.Address)
	if err != nil {
		panic(err)
	}
	return addr.String()
}

func (r *Recovery) MustAddr() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(r.Address)
}
