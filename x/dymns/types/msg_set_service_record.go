package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgSetServiceRecord{}

// ValidateBasic performs basic validation for the MsgSetServiceRecord.
func (m *MsgSetServiceRecord) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "name is not a valid dym name")
	}

	_, config := m.GetDymNameConfig()
	if err := config.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "config is invalid: %v", err.Error())
	}

	if !dymnsutils.IsValidBech32AccountAddress(m.Controller, true) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "controller is not a valid bech32 account address")
	}

	return nil
}

// GetDymNameConfig casts MsgSetServiceRecord into DymNameConfig.
func (m *MsgSetServiceRecord) GetDymNameConfig() (name string, config DymNameConfig) {
	return m.Name, DymNameConfig{
		Type:    DymNameConfigType_DCT_SERVICE,
		ChainId: "",
		Path:    m.ServiceKey,
		Value:   m.Value,
	}
}
