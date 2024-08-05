package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func TestValidateGenesis(t *testing.T) {
	addrs := accAddrsToString(apptesting.CreateRandomAccounts(3))
	valAddrs := accAddrsToString(apptesting.CreateRandomAccounts(3))

	tests := []struct {
		name          string
		input         *types.GenesisState
		errorIs       error
		errorContains string
	}{
		{
			name: "Valid",
			input: &types.GenesisState{
				Params: types.Params{
					MinAllocationWeight: math.NewInt(20),
					MinVotingPower:      math.NewInt(20),
				},
				VoterInfos: []types.VoterInfo{
					{
						Voter: addrs[0],
						Vote: types.Vote{
							VotingPower: math.NewInt(600),
							Weights: []types.GaugeWeight{
								{GaugeId: 1, Weight: math.NewInt(100)},
							},
						},
						Validators: []types.ValidatorVotingPower{
							{Validator: valAddrs[0], Power: math.NewInt(600)},
						},
					},
					{
						Voter: addrs[1],
						Vote: types.Vote{
							VotingPower: math.NewInt(400),
							Weights: []types.GaugeWeight{
								{GaugeId: 2, Weight: math.NewInt(100)},
							},
						},
						Validators: []types.ValidatorVotingPower{
							{Validator: valAddrs[0], Power: math.NewInt(400)},
						},
					},
				},
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name:          "Default is valid",
			input:         types.DefaultGenesis(),
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "Invalid params: MinAllocationWeight < 0",
			input: &types.GenesisState{
				Params: types.Params{
					MinAllocationWeight: math.NewInt(-20),
					MinVotingPower:      math.NewInt(20),
				},
				VoterInfos: []types.VoterInfo{
					{
						Voter: addrs[0],
						Vote: types.Vote{
							VotingPower: math.NewInt(600),
							Weights: []types.GaugeWeight{
								{GaugeId: 1, Weight: math.NewInt(100)},
							},
						},
						Validators: []types.ValidatorVotingPower{
							{Validator: valAddrs[0], Power: math.NewInt(600)},
						},
					},
					{
						Voter: addrs[1],
						Vote: types.Vote{
							VotingPower: math.NewInt(400),
							Weights: []types.GaugeWeight{
								{GaugeId: 2, Weight: math.NewInt(100)},
							},
						},
						Validators: []types.ValidatorVotingPower{
							{Validator: valAddrs[0], Power: math.NewInt(400)},
						},
					},
				},
			},
			errorIs:       types.ErrInvalidGenesis,
			errorContains: "MinAllocationWeight must be >= 0, got -20",
		},
		{
			name: "Invalid voter info: voting power mismatch",
			input: &types.GenesisState{
				Params: types.Params{
					MinAllocationWeight: math.NewInt(20),
					MinVotingPower:      math.NewInt(20),
				},
				VoterInfos: []types.VoterInfo{
					{
						Voter: addrs[0],
						Vote: types.Vote{
							VotingPower: math.NewInt(600),
							Weights: []types.GaugeWeight{
								{GaugeId: 1, Weight: math.NewInt(100)},
							},
						},
						Validators: []types.ValidatorVotingPower{
							{Validator: valAddrs[0], Power: math.NewInt(600)},
						},
					},
					{
						Voter: addrs[1],
						Vote: types.Vote{
							VotingPower: math.NewInt(400),
							Weights: []types.GaugeWeight{
								{GaugeId: 2, Weight: math.NewInt(100)},
							},
						},
						Validators: []types.ValidatorVotingPower{
							{Validator: valAddrs[0], Power: math.NewInt(500)}, // <-- mismatch
						},
					},
				},
			},
			errorIs:       types.ErrInvalidGenesis,
			errorContains: "voting power mismatch: vote voting power 400 is less than total validator power 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()

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

func TestValidateVoterInfo(t *testing.T) {
	addrs := accAddrsToString(apptesting.CreateRandomAccounts(3))
	valAddrs := accAddrsToString(apptesting.CreateRandomAccounts(3))

	tests := []struct {
		name          string
		input         types.VoterInfo
		errorIs       error
		errorContains string
	}{
		{
			name: "Valid",
			input: types.VoterInfo{
				Voter: addrs[0],
				Vote: types.Vote{
					VotingPower: math.NewInt(600),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(100)},
					},
				},
				Validators: []types.ValidatorVotingPower{
					{Validator: valAddrs[0], Power: math.NewInt(400)},
					{Validator: valAddrs[1], Power: math.NewInt(200)},
				},
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "Invalid voter address",
			input: types.VoterInfo{
				Voter: "asdasd", // <-- invalid
				Vote: types.Vote{
					VotingPower: math.NewInt(600),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(100)},
					},
				},
				Validators: []types.ValidatorVotingPower{
					{Validator: valAddrs[0], Power: math.NewInt(400)},
					{Validator: valAddrs[1], Power: math.NewInt(200)},
				},
			},
			errorIs:       types.ErrInvalidVoterInfo,
			errorContains: "voter 'asdasd' must be a valid bech32 address",
		},
		{
			name: "Invalid vote: weight > 100",
			input: types.VoterInfo{
				Voter: addrs[0],
				Vote: types.Vote{
					VotingPower: math.NewInt(600),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(101)}, // <-- 101%
					},
				},
				Validators: []types.ValidatorVotingPower{
					{Validator: valAddrs[0], Power: math.NewInt(400)},
					{Validator: valAddrs[1], Power: math.NewInt(200)},
				},
			},
			errorIs:       types.ErrInvalidVoterInfo,
			errorContains: "weight must be <= 100, got 101",
		},
		{
			name: "Invalid validators: duplicated addresses",
			input: types.VoterInfo{
				Voter: addrs[0],
				Vote: types.Vote{
					VotingPower: math.NewInt(600),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(100)},
					},
				},
				Validators: []types.ValidatorVotingPower{
					{Validator: valAddrs[0], Power: math.NewInt(400)},
					{Validator: valAddrs[0], Power: math.NewInt(200)}, // <-- duplicated
				},
			},
			errorIs:       types.ErrInvalidVoterInfo,
			errorContains: "duplicated validators",
		},
		{
			name: "Invalid validators: voting power for validators is greater than the vote voting power",
			input: types.VoterInfo{
				Voter: addrs[0],
				Vote: types.Vote{
					VotingPower: math.NewInt(600),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(100)},
					},
				},
				Validators: []types.ValidatorVotingPower{
					// 300 + 400 > 600
					{Validator: valAddrs[0], Power: math.NewInt(400)},
					{Validator: valAddrs[1], Power: math.NewInt(300)},
				},
			},
			errorIs:       types.ErrInvalidVoterInfo,
			errorContains: "voting power mismatch: vote voting power 600 is less than total validator power 700",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()

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
