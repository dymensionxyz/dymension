package types_test

import (
	"testing"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func TestStateInfo_DAPathEncoding(t *testing.T) {
	testDA := types.DAPath{
		DaType: "interchainda",
	}
	marshalTestDA, err := testDA.Marshal()
	require.NoError(t, err)
	v2DaPath := string(marshalTestDA)

	tests := []struct {
		name        string
		state       types.StateInfo
		dapath      types.DAPath
		expectError bool
	}{
		{
			name: "state with empty DAPath",
			state: types.StateInfo{
				DAPath: "",
			},
			dapath: types.DAPath{},
		},
		{
			name: "state with valid old DAPath",
			state: types.StateInfo{
				DAPath: "celestia|1507341|8|7|f700184ffd37b0eaf4ab0d3603175f4009169b969f8c8557fd58a864a51cdba6|00000000000000000000000000000000000000bcfaef0d36e7428befbd|db4fc0fd77cec11e10fa8a1a0889627230be1784193b356ec9349ea05420947b",
			},
			dapath:      types.DAPath{},
			expectError: true,
		},
		{
			name: "state with v2 DAPath",
			state: types.StateInfo{
				DAPath: v2DaPath,
			},
			dapath: testDA,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError {
				daPath, err := tt.state.GetDAPathAsDAPath()
				require.Error(t, err)
				require.Equal(t, tt.dapath, daPath)

				daType := tt.state.GetDAType()
				require.Equal(t, "", daType)
			} else {
				daPath, err := tt.state.GetDAPathAsDAPath()
				require.NoError(t, err)
				require.Equal(t, tt.dapath.DaType, daPath.DaType)
				require.Equal(t, tt.dapath.Commitment, daPath.Commitment)

				daType := tt.state.GetDAType()
				require.Equal(t, tt.dapath.DaType, daType)
			}
		})
	}
}
