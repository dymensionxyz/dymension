package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgPunishSequencer{}

// ValidateBasic runs basic stateless validity checks
func (m *MsgPunishSequencer) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Rewardee)
	if err != nil {
		return fmt.Errorf("failed to parse rewardee address: %w", err)
	}

	_, err = sdk.AccAddressFromBech32(m.PunishSequencerAddress)
	if err != nil {
		return fmt.Errorf("failed to parse punish sequencer address: %w", err)
	}

	return nil
}
