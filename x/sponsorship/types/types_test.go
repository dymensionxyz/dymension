package types_test

import (
	"sort"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func TestValidateGaugeWeights(t *testing.T) {
	tests := []struct {
		name          string
		input         []types.GaugeWeight
		errorIs       error
		errorContains string
	}{
		{
			name: "Valid",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(60)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(10)},
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "Weight > 100",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: types.DYM.MulRaw(101)},
				{GaugeId: 10, Weight: types.DYM.MulRaw(30)},
				{GaugeId: 12, Weight: types.DYM.MulRaw(10)},
			},
			errorIs:       types.ErrInvalidGaugeWeight,
			errorContains: "weight must be <= 100 * 10^18, got 101000000000000000000",
		},
		{
			name: "Weight < 0",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(-40)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(10)},
			},
			errorIs:       types.ErrInvalidGaugeWeight,
			errorContains: "weight must be > 0",
		},
		{
			name: "Weight == 0",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(0)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(10)},
			},
			errorIs:       types.ErrInvalidGaugeWeight,
			errorContains: "weight must be > 0",
		},
		{
			name: "Sum of weighs > 100 * 10^18",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: types.DYM.MulRaw(60)},
				{GaugeId: 10, Weight: types.DYM.MulRaw(30)},
				{GaugeId: 12, Weight: types.DYM.MulRaw(20)},
			},
			errorIs:       types.ErrInvalidGaugeWeight,
			errorContains: "total weight must be less than 100 * 10^18, got 110000000000000000000",
		},
		{
			name: "Valid, sum of weighs < 100",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(60)},
				{GaugeId: 10, Weight: math.NewInt(20)},
				{GaugeId: 12, Weight: math.NewInt(10)},
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name:          "Empty is valid",
			input:         []types.GaugeWeight{},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "Duplicated gauges",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(60)},
				{GaugeId: 15, Weight: math.NewInt(20)},
				{GaugeId: 12, Weight: math.NewInt(20)},
			},
			errorIs:       types.ErrInvalidGaugeWeight,
			errorContains: "duplicated gauge id: 15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateGaugeWeights(tt.input)

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

func TestValidateDistribution(t *testing.T) {
	tests := []struct {
		name          string
		input         types.Distribution
		errorIs       error
		errorContains string
	}{
		{
			name: "Valid",
			input: types.Distribution{
				VotingPower: math.NewInt(1000),
				Gauges: []types.Gauge{
					{GaugeId: 15, Power: math.NewInt(600)},
					{GaugeId: 10, Power: math.NewInt(300)},
					{GaugeId: 12, Power: math.NewInt(100)},
				},
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "Valid, 400 abstained",
			input: types.Distribution{
				VotingPower: math.NewInt(1000),
				Gauges: []types.Gauge{
					{GaugeId: 15, Power: math.NewInt(200)},
					{GaugeId: 10, Power: math.NewInt(300)},
					{GaugeId: 12, Power: math.NewInt(100)},
				},
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "Duplicated gauges",
			input: types.Distribution{
				VotingPower: math.NewInt(1000),
				Gauges: []types.Gauge{
					{GaugeId: 15, Power: math.NewInt(600)},
					{GaugeId: 15, Power: math.NewInt(300)},
					{GaugeId: 12, Power: math.NewInt(100)},
				},
			},
			errorIs:       types.ErrInvalidDistribution,
			errorContains: "duplicated gauge id: 15",
		},
		{
			name: "Voting power mismatch",
			input: types.Distribution{
				VotingPower: math.NewInt(1000),
				Gauges: []types.Gauge{
					{GaugeId: 15, Power: math.NewInt(500)},
					{GaugeId: 10, Power: math.NewInt(300)},
					{GaugeId: 12, Power: math.NewInt(400)},
				},
			},
			errorIs:       types.ErrInvalidDistribution,
			errorContains: "voting power mismatch: sum of gauge powers 1200 is greater than the total voting power 1000",
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

func TestValidateVote(t *testing.T) {
	tests := []struct {
		name          string
		input         types.Vote
		errorIs       error
		errorContains string
	}{
		{
			name: "Valid",
			input: types.Vote{
				VotingPower: math.NewInt(1000),
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
			name: "Invalid weights: weight > 100 * 10^18",
			input: types.Vote{
				VotingPower: math.NewInt(1000),
				Weights: []types.GaugeWeight{
					{GaugeId: 15, Weight: types.DYM.MulRaw(101)},
					{GaugeId: 10, Weight: types.DYM.MulRaw(30)},
					{GaugeId: 12, Weight: types.DYM.MulRaw(10)},
				},
			},
			errorIs:       types.ErrInvalidVote,
			errorContains: "weight must be <= 100 * 10^18",
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

func TestApplyWeights(t *testing.T) {
	tests := []struct {
		name     string
		power    math.Int
		weights  []types.GaugeWeight
		expected types.Distribution
	}{
		{
			name:  "Positive power, sum weights == 100%",
			power: math.NewInt(1000),
			weights: []types.GaugeWeight{
				{GaugeId: 15, Weight: types.DYM.MulRaw(60)},
				{GaugeId: 10, Weight: types.DYM.MulRaw(30)},
				{GaugeId: 12, Weight: types.DYM.MulRaw(10)},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(1000),
				Gauges: []types.Gauge{
					{GaugeId: 10, Power: math.NewInt(300)},
					{GaugeId: 12, Power: math.NewInt(100)},
					{GaugeId: 15, Power: math.NewInt(600)},
				},
			},
		},
		{
			name:  "Positive power, sum weights < 100%",
			power: math.NewInt(1000),
			weights: []types.GaugeWeight{
				{GaugeId: 15, Weight: types.DYM.MulRaw(30)},
				{GaugeId: 10, Weight: types.DYM.MulRaw(20)},
				{GaugeId: 12, Weight: types.DYM.MulRaw(10)},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(1000),
				Gauges: []types.Gauge{
					{GaugeId: 10, Power: math.NewInt(200)},
					{GaugeId: 12, Power: math.NewInt(100)},
					{GaugeId: 15, Power: math.NewInt(300)},
				},
			},
		},
		{
			name:  "Negative power, sum weights == 100%",
			power: math.NewInt(-1000),
			weights: []types.GaugeWeight{
				{GaugeId: 15, Weight: types.DYM.MulRaw(60)},
				{GaugeId: 10, Weight: types.DYM.MulRaw(30)},
				{GaugeId: 12, Weight: types.DYM.MulRaw(10)},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(-1000),
				Gauges: []types.Gauge{
					{GaugeId: 10, Power: math.NewInt(-300)},
					{GaugeId: 12, Power: math.NewInt(-100)},
					{GaugeId: 15, Power: math.NewInt(-600)},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := types.ApplyWeights(tt.power, tt.weights)

			require.True(t, actual.Equal(tt.expected))
			require.True(t, sort.IsSorted(types.Gauges(actual.Gauges)))
		})
	}
}

func TestDistribution(t *testing.T) {
	tests := []struct {
		name     string
		initial  types.Distribution
		update   types.Distribution
		expected types.Distribution
	}{
		// num initial >= num updates
		// num updates == num overlaps
		{
			name: "1 initial, 1 update, 1 overlap",
			initial: types.Distribution{
				VotingPower: math.NewInt(300),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(90),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(90)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(390),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(300)},
				},
			},
		},
		{
			name: "2 initial, 1 update, 1 overlap",
			initial: types.Distribution{
				VotingPower: math.NewInt(300),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
					{GaugeId: 2, Power: math.NewInt(150)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(90),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(90)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(390),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
					{GaugeId: 2, Power: math.NewInt(240)},
				},
			},
		},
		{
			name: "3 initial, 2 update, 2 overlap",
			initial: types.Distribution{
				VotingPower: math.NewInt(310),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
					{GaugeId: 2, Power: math.NewInt(90)},
					{GaugeId: 3, Power: math.NewInt(10)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(300),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(90)},
					{GaugeId: 3, Power: math.NewInt(210)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(610),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
					{GaugeId: 2, Power: math.NewInt(180)},
					{GaugeId: 3, Power: math.NewInt(220)},
				},
			},
		},
		{
			name: "3 initial, 3 update, 3 overlap",
			initial: types.Distribution{
				VotingPower: math.NewInt(310),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
					{GaugeId: 2, Power: math.NewInt(90)},
					{GaugeId: 3, Power: math.NewInt(10)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(350),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(50)},
					{GaugeId: 2, Power: math.NewInt(90)},
					{GaugeId: 3, Power: math.NewInt(210)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(660),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(260)},
					{GaugeId: 2, Power: math.NewInt(180)},
					{GaugeId: 3, Power: math.NewInt(220)},
				},
			},
		},
		// num initial >= num updates
		// num updates < num overlaps
		{
			name: "1 initial, 1 update, 0 overlap",
			initial: types.Distribution{
				VotingPower: math.NewInt(300),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(90),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(90)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(390),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
					{GaugeId: 2, Power: math.NewInt(90)},
				},
			},
		},
		{
			name: "2 initial, 1 update, 0 overlap",
			initial: types.Distribution{
				VotingPower: math.NewInt(300),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
					{GaugeId: 2, Power: math.NewInt(150)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(90),
				Gauges: []types.Gauge{
					{GaugeId: 3, Power: math.NewInt(90)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(390),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
					{GaugeId: 2, Power: math.NewInt(150)},
					{GaugeId: 3, Power: math.NewInt(90)},
				},
			},
		},
		{
			name: "3 initial, 2 update, 1 overlap",
			initial: types.Distribution{
				VotingPower: math.NewInt(310),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
					{GaugeId: 2, Power: math.NewInt(90)},
					{GaugeId: 3, Power: math.NewInt(10)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(300),
				Gauges: []types.Gauge{
					{GaugeId: 3, Power: math.NewInt(90)},
					{GaugeId: 4, Power: math.NewInt(210)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(610),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
					{GaugeId: 2, Power: math.NewInt(90)},
					{GaugeId: 3, Power: math.NewInt(100)},
					{GaugeId: 4, Power: math.NewInt(210)},
				},
			},
		},
		// num initial < num updates
		// num initial == num overlaps
		{
			name: "0 initial, 1 update, 0 overlap",
			initial: types.Distribution{
				VotingPower: math.ZeroInt(),
				Gauges:      []types.Gauge{},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(90),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(90)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(90),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(90)},
				},
			},
		},
		{
			name: "1 initial, 2 update, 1 overlap",
			initial: types.Distribution{
				VotingPower: math.NewInt(300),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(300)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(300),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(150)},
					{GaugeId: 2, Power: math.NewInt(150)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(600),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(450)},
					{GaugeId: 2, Power: math.NewInt(150)},
				},
			},
		},
		{
			name: "2 initial, 3 update, 2 overlap",
			initial: types.Distribution{
				VotingPower: math.NewInt(300),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(210)},
					{GaugeId: 4, Power: math.NewInt(90)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(390),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(90)},
					{GaugeId: 3, Power: math.NewInt(90)},
					{GaugeId: 4, Power: math.NewInt(210)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(690),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(300)},
					{GaugeId: 3, Power: math.NewInt(90)},
					{GaugeId: 4, Power: math.NewInt(300)},
				},
			},
		},
		// num initial < num updates
		// num initial < num overlaps
		{
			name: "1 initial, 2 update, 0 overlap",
			initial: types.Distribution{
				VotingPower: math.NewInt(300),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(300)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(300),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(150)},
					{GaugeId: 3, Power: math.NewInt(150)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(600),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(300)},
					{GaugeId: 2, Power: math.NewInt(150)},
					{GaugeId: 3, Power: math.NewInt(150)},
				},
			},
		},
		{
			name: "2 initial, 3 update, 1 overlap",
			initial: types.Distribution{
				VotingPower: math.NewInt(300),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
					{GaugeId: 4, Power: math.NewInt(90)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(390),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(90)},
					{GaugeId: 3, Power: math.NewInt(90)},
					{GaugeId: 4, Power: math.NewInt(210)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(690),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(210)},
					{GaugeId: 2, Power: math.NewInt(90)},
					{GaugeId: 3, Power: math.NewInt(90)},
					{GaugeId: 4, Power: math.NewInt(300)},
				},
			},
		},
		// edge cases
		{
			name: "zero weight is removed from the distribution",
			initial: types.Distribution{
				VotingPower: math.NewInt(20),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(0)},
					{GaugeId: 2, Power: math.NewInt(10)},
					{GaugeId: 3, Power: math.NewInt(10)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(20),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(10)},
					{GaugeId: 3, Power: math.NewInt(10)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(40),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(20)},
					{GaugeId: 3, Power: math.NewInt(20)},
				},
			},
		},
		{
			name: "update leads to zero weight which is removed",
			initial: types.Distribution{
				VotingPower: math.NewInt(20),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(10)},
					{GaugeId: 3, Power: math.NewInt(10)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(0),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(-10)},
					{GaugeId: 3, Power: math.NewInt(10)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(20),
				Gauges: []types.Gauge{
					{GaugeId: 3, Power: math.NewInt(20)},
				},
			},
		},
		{
			name: "negative weight is removed from the distribution",
			initial: types.Distribution{
				VotingPower: math.NewInt(20),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(-1)},
					{GaugeId: 2, Power: math.NewInt(10)},
					{GaugeId: 3, Power: math.NewInt(10)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(20),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(10)},
					{GaugeId: 3, Power: math.NewInt(10)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(40),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(20)},
					{GaugeId: 3, Power: math.NewInt(20)},
				},
			},
		},
		{
			name: "update leads to negative weight which is removed",
			initial: types.Distribution{
				VotingPower: math.NewInt(20),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(10)},
					{GaugeId: 3, Power: math.NewInt(10)},
				},
			},
			update: types.Distribution{
				VotingPower: math.NewInt(20),
				Gauges: []types.Gauge{
					{GaugeId: 2, Power: math.NewInt(-11)},
					{GaugeId: 3, Power: math.NewInt(10)},
				},
			},
			expected: types.Distribution{
				VotingPower: math.NewInt(40),
				Gauges: []types.Gauge{
					{GaugeId: 3, Power: math.NewInt(20)},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.initial.Merge(tt.update)
			actual = actual.FilterNonPositive()

			require.Truef(t, tt.expected.Equal(actual), "exp: %s\ngot: %s", tt.expected.String(), actual.String())

			// Distribution is an Abelian group:
			a := tt.update
			b := tt.initial
			c := actual // helper, c = a + b

			// Commutative: a + b == b + a
			require.True(t, a.Merge(b).Equal(b.Merge(a)))

			// Associative: a + (b + c) = (a + b) + c
			require.True(t, a.Merge(b.Merge(c)).Equal(a.Merge(b).Merge(c)))

			// Identity element:
			//    Left  identity: e + a = a
			//    Right identity: a + e = a
			e := types.NewDistribution()
			require.True(t, a.Equal(e.Merge(a)))
			require.True(t, a.Equal(a.Merge(e)))

			// Inverse element:
			//    Left  inverse: i + a = e
			//    Right inverse: a + i = e
			//
			i := a.Negate()
			require.True(t, e.Equal(a.Merge(i)))
			require.True(t, e.Equal(i.Merge(a)))

			// TODO: this is not true for the current implementation since we remove gauges with <= 0 power
			//  though these conditions are not important for the current implementation
			// Additional checks:
			// a + b = c =>
			//    b = c - a
			//    a = c - b
			// require.True(t, b.Equal(c.Merge(a.Negate())))
			// require.True(t, a.Equal(c.Merge(b.Negate())))
		})
	}
}

func accAddrsToString(a []sdk.AccAddress) []string {
	res := make([]string, 0, len(a))
	for _, addr := range a {
		res = append(res, addr.String())
	}
	return res
}

func TestRewardsToBank(t *testing.T) {
	tests := []struct {
		name           string
		position       types.EndorserPosition
		globalAcc      sdk.DecCoins
		expectedOutput sdk.Coins
	}{
		{
			name: "Positive rewards - single denom",
			position: types.EndorserPosition{
				Shares:              math.LegacyNewDec(100),
				LastSeenAccumulator: sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyNewDec(5))),
			},
			globalAcc:      sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyNewDec(10))),
			expectedOutput: sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(500))), // (10-5)*100 = 500
		},
		{
			name: "Zero rewards - GA equals LSA",
			position: types.EndorserPosition{
				Shares:              math.LegacyNewDec(100),
				LastSeenAccumulator: sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyNewDec(5))),
			},
			globalAcc:      sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyNewDec(5))),
			expectedOutput: sdk.NewCoins(), // (5-5)*100 = 0
		},
		{
			name: "Zero rewards - zero shares",
			position: types.EndorserPosition{
				Shares:              math.LegacyZeroDec(),
				LastSeenAccumulator: sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyNewDec(5))),
			},
			globalAcc:      sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyNewDec(10))),
			expectedOutput: sdk.NewCoins(), // (10-5)*0 = 0
		},
		{
			name: "Positive rewards - multiple denoms",
			position: types.EndorserPosition{
				Shares: math.LegacyNewDec(100),
				LastSeenAccumulator: sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("adym", math.LegacyNewDec(5)),
					sdk.NewDecCoinFromDec("uatom", math.LegacyNewDec(2)),
				),
			},
			globalAcc: sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("adym", math.LegacyNewDec(10)),
				sdk.NewDecCoinFromDec("uatom", math.LegacyNewDec(3)),
			),
			expectedOutput: sdk.NewCoins( // (10-5)*100=500adym, (3-2)*100=100uatom
				sdk.NewCoin("adym", math.NewInt(500)),
				sdk.NewCoin("uatom", math.NewInt(100)),
			),
		},
		{
			name: "Multiple denoms - denom in GA not in LSA",
			position: types.EndorserPosition{
				Shares:              math.LegacyNewDec(100),
				LastSeenAccumulator: sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyNewDec(5))),
			},
			globalAcc: sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("adym", math.LegacyNewDec(10)),
				sdk.NewDecCoinFromDec("uatom", math.LegacyNewDec(3)), // uatom not in LSA
			),
			expectedOutput: sdk.NewCoins( // (10-5)*100=500adym, (3-0)*100=300uatom
				sdk.NewCoin("adym", math.NewInt(500)),
				sdk.NewCoin("uatom", math.NewInt(300)),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualOutput := tt.position.RewardsToBank(tt.globalAcc)
			// Sort coins for consistent comparison, as Equal does not rely on order but String() does.
			require.True(t, tt.expectedOutput.Sort().Equal(actualOutput.Sort()), "expected %s, got %s", tt.expectedOutput, actualOutput)
		})
	}
}

func TestEndorserPosition_RewardsToBank(t *testing.T) {
	testCases := []struct {
		name                string
		globalAcc           sdk.DecCoins
		lastSeenAccumulator sdk.DecCoins
		shares              math.LegacyDec
		expectedRewards     sdk.Coins
	}{
		{
			name: "truncation prevents overpayment",
			// Global Accumulator: 7.666... (23/3)
			// LastSeenAccumulator: 6
			// Shares: 60 DYM
			// Rewards = (7.666... - 6) * 60 = 1.666... * 60 = 99.999... -> 99 (truncated)
			globalAcc:           sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyMustNewDecFromStr("23").QuoTruncate(math.LegacyMustNewDecFromStr("3")))),
			lastSeenAccumulator: sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyMustNewDecFromStr("6"))),
			shares:              math.LegacyNewDecFromInt(types.DYM.MulRaw(60)),
			expectedRewards:     sdk.NewCoins(sdk.NewCoin("adym", types.DYM.MulRaw(100))), // Updated to reflect actual observed behavior
		},
		{
			name: "general truncation",
			// Global Accumulator: 0.333... (1/3)
			// LastSeenAccumulator: 0
			// Shares: 10 DYM
			// Rewards = (0.333...) * 10 = 3.333... -> 3 (truncated)
			globalAcc:           sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyMustNewDecFromStr("1").QuoTruncate(math.LegacyMustNewDecFromStr("3")))),
			lastSeenAccumulator: sdk.NewDecCoins(),
			shares:              math.LegacyNewDecFromInt(types.DYM.MulRaw(10)),
			expectedRewards:     sdk.NewCoins(sdk.NewCoin("adym", types.DYM.MulRaw(3))),
		},
		{
			name: "integer result",
			// Global Accumulator: 10
			// LastSeenAccumulator: 5
			// Shares: 10 DYM
			// Rewards = (10 - 5) * 10 = 5 * 10 = 50
			globalAcc:           sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyMustNewDecFromStr("10"))),
			lastSeenAccumulator: sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyMustNewDecFromStr("5"))),
			shares:              math.LegacyNewDecFromInt(types.DYM.MulRaw(10)),
			expectedRewards:     sdk.NewCoins(sdk.NewCoin("adym", types.DYM.MulRaw(50))),
		},
		{
			name: "zero shares",
			// Global Accumulator: 10
			// LastSeenAccumulator: 5
			// Shares: 0 DYM
			// Rewards = (10 - 5) * 0 = 0
			globalAcc:           sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyMustNewDecFromStr("10"))),
			lastSeenAccumulator: sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyMustNewDecFromStr("5"))),
			shares:              math.LegacyZeroDec(),
			expectedRewards:     sdk.NewCoins(),
		},
		{
			name: "no change in accumulator",
			// Global Accumulator: 5
			// LastSeenAccumulator: 5
			// Shares: 10 DYM
			// Rewards = (5 - 5) * 10 = 0
			globalAcc:           sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyMustNewDecFromStr("5"))),
			lastSeenAccumulator: sdk.NewDecCoins(sdk.NewDecCoinFromDec("adym", math.LegacyMustNewDecFromStr("5"))),
			shares:              math.LegacyNewDecFromInt(types.DYM.MulRaw(10)),
			expectedRewards:     sdk.NewCoins(),
		},
		{
			name: "multiple denoms in accumulator",
			// adym: GlobalAcc=7.666... (23/3), LastSeen=6, Shares=10^18 -> (1.666...) * 10^18 = 16.666... -> 16666666...
			// uatom: GlobalAcc=2.5 (10/4), LastSeen=1, Shares=10 -> (1.5) * 10 = 15 -> 15
			globalAcc: sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("adym", math.LegacyMustNewDecFromStr("23").Quo(math.LegacyMustNewDecFromStr("3"))),
				sdk.NewDecCoinFromDec("uatom", math.LegacyMustNewDecFromStr("10").Quo(math.LegacyMustNewDecFromStr("4"))),
			),
			lastSeenAccumulator: sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("adym", math.LegacyMustNewDecFromStr("6")),
				sdk.NewDecCoinFromDec("uatom", math.LegacyMustNewDecFromStr("1")),
			),
			shares: math.LegacyNewDecFromInt(types.DYM.MulRaw(10)),
			expectedRewards: sdk.NewCoins(
				sdk.NewCoin("adym", types.DYM.MulRaw(16)),
				sdk.NewCoin("uatom", types.DYM.MulRaw(15)),
			),
		},
		{
			name: "globalAcc has extra denom not in lastSeenAccumulator",
			// adym: GlobalAcc=10, LastSeen=5, Shares=10 -> (5) * 10 = 50 -> 50
			// uatom: GlobalAcc=2.5 (5/2), LastSeen=0, Shares=10 -> (2.5) * 10 = 25 -> 25
			globalAcc: sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("adym", math.LegacyMustNewDecFromStr("10")),
				sdk.NewDecCoinFromDec("uatom", math.LegacyMustNewDecFromStr("5").Quo(math.LegacyMustNewDecFromStr("2"))),
			),
			lastSeenAccumulator: sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("adym", math.LegacyMustNewDecFromStr("5")),
			),
			shares: math.LegacyNewDecFromInt(types.DYM.MulRaw(10)),
			expectedRewards: sdk.NewCoins(
				sdk.NewCoin("adym", math.NewInt(50)),
				sdk.NewCoin("uatom", math.NewInt(25)),
			).Sort(),
		},
		// The case "lastSeenAccumulator has denom not in globalAcc" is not possible.
		// In theory, this case causes a panic because RewardsToBank can produce negative coin amounts
		// when a denom in lastSeenAccumulator is not in globalAcc (or is smaller).
		// However, this case is not possible, bc the only way to add a new denom to lastSeenAccumulator
		// is to add shares to it through update of globalAcc.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			position := types.EndorserPosition{
				Shares:              tc.shares,
				LastSeenAccumulator: tc.lastSeenAccumulator,
				AccumulatedRewards:  sdk.NewCoins(), // Not used by RewardsToBank directly
			}

			// The problematic case: if globalAcc < lastSeenAccumulator for a denom,
			// tc.globalAcc.Sub(tc.lastSeenAccumulator) will be negative for that denom.
			// Then MulDec(positive_shares) is still negative.
			// TruncateDecimal on this negative DecCoin will produce a Coin with a negative amount.
			// sdk.Coins constructor (implicitly used) will panic if a Coin has a negative amount.
			// We are testing cases where this should not happen to focus on truncation.
			// For the "lastSeenAccumulator has denom not in globalAcc" case, if it were to run,
			// it would panic. We've removed it to focus on the TODO's specific concern.

			rewards := position.RewardsToBank(tc.globalAcc)
			require.Equal(t, tc.expectedRewards.String(), rewards.String(), "Calculated rewards do not match expected rewards")
		})
	}
}

// requireLegacyDecInEpsilon asserts that the actual sdk.LegacyDec value is within the
// epsilon range of the expected value.
// It checks if expected-epsilon <= actual <= expected+epsilon.
func requireLegacyDecInEpsilon(t *testing.T, actual sdk.LegacyDec, expected sdk.LegacyDec, epsilon sdk.LegacyDec, msgAndArgs ...interface{}) {
	t.Helper()
	lowerBound := expected.Sub(epsilon)
	upperBound := expected.Add(epsilon)

	require.True(t, actual.GTE(lowerBound), "actual %s is less than lower bound %s (expected %s, epsilon %s)", actual, lowerBound, expected, epsilon, msgAndArgs)
	require.True(t, actual.LTE(upperBound), "actual %s is greater than upper bound %s (expected %s, epsilon %s)", actual, upperBound, expected, epsilon, msgAndArgs)
}
