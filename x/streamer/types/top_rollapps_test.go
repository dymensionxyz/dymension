package types_test

import (
	"testing"

	"cosmossdk.io/math"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func TestMergeTopRollApps(t *testing.T) {
	tests := []struct {
		name string
		tops [][]types.PumpPressure
		want []types.PumpPressure
	}{
		{
			name: "no slices",
			tops: [][]types.PumpPressure{},
			want: []types.PumpPressure{},
		},
		{
			name: "single slice",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
				{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
			},
		},
		{
			name: "single slice with zero pressure",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(0)},
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
			},
		},
		{
			name: "both empty",
			tops: [][]types.PumpPressure{{}, {}},
			want: []types.PumpPressure{},
		},
		{
			name: "left empty",
			tops: [][]types.PumpPressure{
				{},
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
				{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
			},
		},
		{
			name: "right empty",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				},
				{},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
				{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
			},
		},
		{
			name: "no overlap - interleaved",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
					{RollappId: "rollapp_3", Pressure: math.NewInt(300)},
				},
				{
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
					{RollappId: "rollapp_4", Pressure: math.NewInt(400)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
				{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				{RollappId: "rollapp_3", Pressure: math.NewInt(300)},
				{RollappId: "rollapp_4", Pressure: math.NewInt(400)},
			},
		},
		{
			name: "no overlap - all lhs before rhs",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				},
				{
					{RollappId: "rollapp_3", Pressure: math.NewInt(300)},
					{RollappId: "rollapp_4", Pressure: math.NewInt(400)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
				{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				{RollappId: "rollapp_3", Pressure: math.NewInt(300)},
				{RollappId: "rollapp_4", Pressure: math.NewInt(400)},
			},
		},
		{
			name: "complete overlap - same rollapp IDs",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				},
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(50)},
					{RollappId: "rollapp_2", Pressure: math.NewInt(150)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_1", Pressure: math.NewInt(150)},
				{RollappId: "rollapp_2", Pressure: math.NewInt(350)},
			},
		},
		{
			name: "partial overlap",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
					{RollappId: "rollapp_3", Pressure: math.NewInt(300)},
				},
				{
					{RollappId: "rollapp_2", Pressure: math.NewInt(50)},
					{RollappId: "rollapp_4", Pressure: math.NewInt(400)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
				{RollappId: "rollapp_2", Pressure: math.NewInt(250)},
				{RollappId: "rollapp_3", Pressure: math.NewInt(300)},
				{RollappId: "rollapp_4", Pressure: math.NewInt(400)},
			},
		},
		{
			name: "zero pressure filtered out - from merge",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				},
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(-100)},
					{RollappId: "rollapp_2", Pressure: math.NewInt(50)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_2", Pressure: math.NewInt(250)},
			},
		},
		{
			name: "zero pressure in input - not included",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(0)},
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				},
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
				{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
			},
		},
		{
			name: "negative pressures",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(-100)},
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				},
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(-50)},
					{RollappId: "rollapp_3", Pressure: math.NewInt(300)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_1", Pressure: math.NewInt(-150)},
				{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				{RollappId: "rollapp_3", Pressure: math.NewInt(300)},
			},
		},
		{
			name: "complex scenario with multiple overlaps",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
					{RollappId: "rollapp_3", Pressure: math.NewInt(300)},
					{RollappId: "rollapp_5", Pressure: math.NewInt(500)},
					{RollappId: "rollapp_7", Pressure: math.NewInt(700)},
				},
				{
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
					{RollappId: "rollapp_3", Pressure: math.NewInt(30)},
					{RollappId: "rollapp_4", Pressure: math.NewInt(400)},
					{RollappId: "rollapp_5", Pressure: math.NewInt(-500)},
					{RollappId: "rollapp_6", Pressure: math.NewInt(600)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
				{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				{RollappId: "rollapp_3", Pressure: math.NewInt(330)},
				{RollappId: "rollapp_4", Pressure: math.NewInt(400)},
				{RollappId: "rollapp_6", Pressure: math.NewInt(600)},
				{RollappId: "rollapp_7", Pressure: math.NewInt(700)},
			},
		},
		{
			name: "three slices - no overlap",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
				},
				{
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				},
				{
					{RollappId: "rollapp_3", Pressure: math.NewInt(300)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
				{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				{RollappId: "rollapp_3", Pressure: math.NewInt(300)},
			},
		},
		{
			name: "three slices - with overlap",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
				},
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(50)},
					{RollappId: "rollapp_3", Pressure: math.NewInt(300)},
				},
				{
					{RollappId: "rollapp_2", Pressure: math.NewInt(25)},
					{RollappId: "rollapp_3", Pressure: math.NewInt(75)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_1", Pressure: math.NewInt(150)},
				{RollappId: "rollapp_2", Pressure: math.NewInt(225)},
				{RollappId: "rollapp_3", Pressure: math.NewInt(375)},
			},
		},
		{
			name: "four slices - complex",
			tops: [][]types.PumpPressure{
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(100)},
					{RollappId: "rollapp_3", Pressure: math.NewInt(300)},
				},
				{
					{RollappId: "rollapp_2", Pressure: math.NewInt(200)},
					{RollappId: "rollapp_3", Pressure: math.NewInt(30)},
				},
				{
					{RollappId: "rollapp_1", Pressure: math.NewInt(10)},
					{RollappId: "rollapp_4", Pressure: math.NewInt(400)},
				},
				{
					{RollappId: "rollapp_2", Pressure: math.NewInt(20)},
					{RollappId: "rollapp_3", Pressure: math.NewInt(-330)},
				},
			},
			want: []types.PumpPressure{
				{RollappId: "rollapp_1", Pressure: math.NewInt(110)},
				{RollappId: "rollapp_2", Pressure: math.NewInt(220)},
				{RollappId: "rollapp_4", Pressure: math.NewInt(400)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.MergeTopRollApps(tt.tops...)

			if len(got) != len(tt.want) {
				t.Errorf("MergeTopRollApps() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].RollappId != tt.want[i].RollappId {
					t.Errorf("MergeTopRollApps()[%d].RollappId = %v, want %v", i, got[i].RollappId, tt.want[i].RollappId)
				}
				if !got[i].Pressure.Equal(tt.want[i].Pressure) {
					t.Errorf("MergeTopRollApps()[%d].Pressure = %v, want %v", i, got[i].Pressure, tt.want[i].Pressure)
				}
			}
		})
	}
}
