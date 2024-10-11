package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func TestMsgUpdateWhitelistedRelayers(t *testing.T) {
	valAddr := sdk.ValAddress(sample.Acc())
	addr := sample.AccAddress()
	relayers := []string{
		sample.AccAddress(),
		sample.AccAddress(),
	}

	tests := []struct {
		name          string
		input         types.MsgUpdateWhitelistedRelayers
		errorIs       error
		errorContains string
	}{
		{
			name: "valid",
			input: types.MsgUpdateWhitelistedRelayers{
				Creator:  valAddr.String(),
				Relayers: relayers,
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "empty is valid",
			input: types.MsgUpdateWhitelistedRelayers{
				Creator:  valAddr.String(),
				Relayers: []string{},
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "invalid relayer addr",
			input: types.MsgUpdateWhitelistedRelayers{
				Creator:  valAddr.String(),
				Relayers: []string{"invalid"},
			},
			errorIs:       gerrc.ErrInvalidArgument,
			errorContains: "validate whitelisted relayers",
		},
		{
			name: "duplicated relayers",
			input: types.MsgUpdateWhitelistedRelayers{
				Creator:  valAddr.String(),
				Relayers: []string{addr, addr},
			},
			errorIs:       gerrc.ErrInvalidArgument,
			errorContains: "validate whitelisted relayers",
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
