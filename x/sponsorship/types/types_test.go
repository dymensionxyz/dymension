package types_test

import (
	"sort"
	"testing"

	"cosmossdk.io/math"
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
				{GaugeId: 15, Weight: math.NewInt(101)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(10)},
			},
			errorIs:       types.ErrInvalidGaugeWeight,
			errorContains: "weight must be <= 100",
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
			name: "Sum of weighs > 100",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(60)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(20)},
			},
			errorIs:       types.ErrInvalidGaugeWeight,
			errorContains: "total weight must equal 100",
		},
		{
			name: "Sum of weighs < 100",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(60)},
				{GaugeId: 10, Weight: math.NewInt(20)},
				{GaugeId: 12, Weight: math.NewInt(10)},
			},
			errorIs:       types.ErrInvalidGaugeWeight,
			errorContains: "total weight must equal 100",
		},
		{
			name:          "Empty input",
			input:         []types.GaugeWeight{},
			errorIs:       types.ErrInvalidGaugeWeight,
			errorContains: "total weight must equal 100",
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
					{GaugeId: 12, Power: math.NewInt(100)},
				},
			},
			errorIs:       types.ErrInvalidDistribution,
			errorContains: "voting power mismatch: voting power 1000, total gauges power 900",
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
			name: "Invalid weights: weight > 100",
			input: types.Vote{
				VotingPower: math.NewInt(1000),
				Weights: []types.GaugeWeight{
					{GaugeId: 15, Weight: math.NewInt(101)},
					{GaugeId: 10, Weight: math.NewInt(30)},
					{GaugeId: 12, Weight: math.NewInt(10)},
				},
			},
			errorIs:       types.ErrInvalidVote,
			errorContains: "weight must be <= 100",
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
			name:  "Positive power, sum weights == 1",
			power: math.NewInt(1000),
			weights: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(60)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(10)},
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
			name:  "Positive power, sum weights < 1",
			power: math.NewInt(1000),
			weights: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(30)},
				{GaugeId: 10, Weight: math.NewInt(20)},
				{GaugeId: 12, Weight: math.NewInt(10)},
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
			name:  "Negative power, sum weights == 1",
			power: math.NewInt(-1000),
			weights: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(60)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(10)},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.initial.Merge(tt.update)

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

			// Additional checks:
			// a + b = c =>
			//    b = c - a
			//    a = c - b
			require.True(t, b.Equal(c.Merge(a.Negate())))
			require.True(t, a.Equal(c.Merge(b.Negate())))
		})
	}
}
