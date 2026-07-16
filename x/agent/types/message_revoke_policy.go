package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func NewMsgRevokePolicy(authority, fingerprint, reason string) *MsgRevokePolicy {
	return &MsgRevokePolicy{
		Authority:   authority,
		Fingerprint: fingerprint,
		Reason:      reason,
	}
}

func (m *MsgRevokePolicy) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "authority address")
	}
	return ValidateFingerprint(m.Fingerprint)
}

func NewMsgUnrevokePolicy(authority, fingerprint string) *MsgUnrevokePolicy {
	return &MsgUnrevokePolicy{
		Authority:   authority,
		Fingerprint: fingerprint,
	}
}

func (m *MsgUnrevokePolicy) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "authority address")
	}
	return ValidateFingerprint(m.Fingerprint)
}
