package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func TestMsgVote(t *testing.T) {
	addrs := accAddrsToString(apptesting.CreateRandomAccounts(1))

	tests := []struct {
		name          string
		input         types.MsgVote
		errorIs       error
		errorContains string
	}{
		{
			name: "Valid input",
			input: types.MsgVote{
				Voter: addrs[0],
				Weights: []types.GaugeWeight{
					{GaugeId: 15, Weight: math.NewInt(60)},
					{GaugeId: 10, Weight: math.NewInt(30)},
					{GaugeId: 12, Weight: math.NewInt(10)},
				},
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "Invalid signer",
			input: types.MsgVote{
				Voter: "123123",
				Weights: []types.GaugeWeight{
					{GaugeId: 15, Weight: math.NewInt(60)},
					{GaugeId: 10, Weight: math.NewInt(30)},
					{GaugeId: 12, Weight: math.NewInt(10)},
				},
			},
			errorIs:       sdkerrors.ErrInvalidAddress,
			errorContains: "voter '123123' must be a valid bech32 address",
		},
		{
			name: "Invalid distribution, Weight > 100 * 10^18",
			input: types.MsgVote{
				Voter: addrs[0],
				Weights: []types.GaugeWeight{
					{GaugeId: 15, Weight: types.DYM.MulRaw(101)},
					{GaugeId: 10, Weight: types.DYM.MulRaw(30)},
					{GaugeId: 12, Weight: types.DYM.MulRaw(10)},
				},
			},
			errorIs:       types.ErrInvalidDistribution,
			errorContains: "weight must be <= 100 * 10^18",
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

func TestMsgRevokeVote(t *testing.T) {
	addrs := accAddrsToString(apptesting.CreateRandomAccounts(1))

	tests := []struct {
		name          string
		input         types.MsgRevokeVote
		errorIs       error
		errorContains string
	}{
		{
			name: "Valid input",
			input: types.MsgRevokeVote{
				Voter: addrs[0],
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "Invalid signer",
			input: types.MsgRevokeVote{
				Voter: "123123",
			},
			errorIs:       sdkerrors.ErrInvalidAddress,
			errorContains: "voter '123123' must be a valid bech32 address",
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

func TestMsgUpdateParams(t *testing.T) {
	addrs := accAddrsToString(apptesting.CreateRandomAccounts(1))

	tests := []struct {
		name          string
		input         types.MsgUpdateParams
		errorIs       error
		errorContains string
	}{
		{
			name: "Valid input",
			input: types.MsgUpdateParams{
				Authority: addrs[0],
				NewParams: types.DefaultParams(),
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "Invalid signer",
			input: types.MsgUpdateParams{
				Authority: "123123",
				NewParams: types.DefaultParams(),
			},
			errorIs:       sdkerrors.ErrInvalidAddress,
			errorContains: "authority '123123' must be a valid bech32 address",
		},
		{
			name: "Invalid params, MinAllocationWeight < 0",
			input: types.MsgUpdateParams{
				Authority: addrs[0],
				NewParams: types.Params{
					MinAllocationWeight: math.NewInt(-20),
					MinVotingPower:      math.NewInt(20),
				},
			},
			errorIs:       types.ErrInvalidParams,
			errorContains: "MinAllocationWeight must be >= 0",
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
