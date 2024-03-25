package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PacketMetadata struct {
	EIBC *EIBCMetadata `json:"eibc"`
}

type EIBCMetadata struct {
	Fee string `json:"fee"`
}

func (p PacketMetadata) ValidateBasic() error {
	if p.EIBC == nil {
		return ErrMissingEIBCMetadata
	}
	return p.EIBC.ValidateBasic()
}

func (e EIBCMetadata) ValidateBasic() error {
	_, ok := sdk.NewIntFromString(e.Fee)
	if !ok {
		return ErrInvalidEIBCFee
	}
	return nil
}
