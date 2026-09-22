package types

import (
	"fmt"

	dymerrors "github.com/dymensionxyz/dymension/v3/internal/errors"

	"github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgCreateBridgingFeeHook{}
	_ sdk.Msg = &MsgSetBridgingFeeHook{}
	_ sdk.Msg = &MsgCreateAggregationHook{}
	_ sdk.Msg = &MsgSetAggregationHook{}
)

func (m MsgCreateBridgingFeeHook) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return dymerrors.Joinf(sdkerrors.ErrInvalidAddress, err, "owner '%s' must be a valid bech32 address", m.Owner)
	}

	// Validate each fee configuration (empty fees are allowed to disable all fees)
	feeSet := make(map[util.HexAddress]struct{}, len(m.Fees))
	for i, fee := range m.Fees {
		if err := fee.Validate(); err != nil {
			return dymerrors.Joinf(ErrInvalidFee, err, "invalid fee at index %d", i)
		}

		if _, ok := feeSet[fee.TokenId]; ok {
			return fmt.Errorf("duplicate fee entry: %s", fee.TokenId)
		}
		feeSet[fee.TokenId] = struct{}{}
	}

	return nil
}

func (m MsgSetBridgingFeeHook) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return dymerrors.Joinf(sdkerrors.ErrInvalidAddress, err, "owner '%s' must be a valid bech32 address", m.Owner)
	}

	// Validate new owner if ownership transfer is requested
	if !m.RenounceOwnership && m.NewOwner != "" {
		_, err := sdk.AccAddressFromBech32(m.NewOwner)
		if err != nil {
			return dymerrors.Joinf(sdkerrors.ErrInvalidAddress, err, "new_owner '%s' must be a valid bech32 address", m.NewOwner)
		}
	}

	// Cannot both renounce ownership and set new owner
	if m.RenounceOwnership && m.NewOwner != "" {
		return ErrInvalidOwner.Wrap("cannot both renounce ownership and set new owner")
	}

	// Validate each fee configuration (empty fees are allowed to disable all fees)
	for i, fee := range m.Fees {
		if err := fee.Validate(); err != nil {
			return dymerrors.Joinf(ErrInvalidFee, err, "invalid fee at index %d", i)
		}
	}

	return nil
}

func (m MsgCreateAggregationHook) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return dymerrors.Joinf(sdkerrors.ErrInvalidAddress, err, "owner '%s' must be a valid bech32 address", m.Owner)
	}
	return nil
}

func (m MsgSetAggregationHook) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return dymerrors.Joinf(sdkerrors.ErrInvalidAddress, err, "owner '%s' must be a valid bech32 address", m.Owner)
	}

	// Validate new owner if ownership transfer is requested
	if !m.RenounceOwnership && m.NewOwner != "" {
		_, err := sdk.AccAddressFromBech32(m.NewOwner)
		if err != nil {
			return dymerrors.Joinf(sdkerrors.ErrInvalidAddress, err, "new_owner '%s' must be a valid bech32 address", m.NewOwner)
		}
	}

	// Cannot both renounce ownership and set new owner
	if m.RenounceOwnership && m.NewOwner != "" {
		return ErrInvalidOwner.Wrap("cannot both renounce ownership and set new owner")
	}

	return nil
}
