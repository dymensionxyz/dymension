package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func TestMsgUpdateRewardAddress(t *testing.T) {
	valAddr := sdk.ValAddress(sample.Acc())
	rewardAddr := sample.AccAddress()

	tests := []struct {
		name          string
		input         types.MsgUpdateRewardAddress
		errorIs       error
		errorContains string
	}{
		{
			name: "valid",
			input: types.MsgUpdateRewardAddress{
				Creator:    valAddr.String(),
				RewardAddr: rewardAddr,
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "invalid creator",
			input: types.MsgUpdateRewardAddress{
				Creator:    "invalid_creator",
				RewardAddr: rewardAddr,
			},
			errorIs:       gerrc.ErrInvalidArgument,
			errorContains: "get creator addr from bech32",
		},
		{
			name: "invalid reward addr",
			input: types.MsgUpdateRewardAddress{
				Creator:    valAddr.String(),
				RewardAddr: "invalid_reward_addr",
			},
			errorIs:       gerrc.ErrInvalidArgument,
			errorContains: "get reward addr from bech32",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.ValidateBasic()

			expectError := tt.errorIs != nil
			switch expectError {
			case true:
				require.Error(t, err)
				require.ErrorIs(t, err, tt.errorIs)
				require.Contains(t, err.Error(), tt.errorContains)
			case false:
				require.NoError(t, err)
			}
		})
	}
}
