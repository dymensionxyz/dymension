package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChainID(t *testing.T) {
	tests := []struct {
		name      string
		rollappId string
		revision  uint64
		expErr    error
	}{
		{
			name:      "default is valid",
			rollappId: "rollapp_1234-1",
			revision:  1,
			expErr:    nil,
		},
		{
			name:      "too long id",
			rollappId: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz_1234-1",
			expErr:    ErrInvalidRollappID,
		},
		{
			name:      "wrong EIP155",
			rollappId: "rollapp_ea2413-1",
			expErr:    ErrInvalidRollappID,
		},
		{
			name:      "no EIP155 with revision",
			rollappId: "rollapp-1",
			expErr:    ErrInvalidRollappID,
		},
		{
			name:      "starts with dash",
			rollappId: "-1234",
			expErr:    ErrInvalidRollappID,
		},
		{
			name:      "revision set without eip155",
			rollappId: "rollapp-3",
			expErr:    ErrInvalidRollappID,
		},
		{
			name:      "revision not set",
			rollappId: "rollapp",
			expErr:    ErrInvalidRollappID,
		},
		{
			name:      "invalid revision",
			rollappId: "rollapp_1234-x",
			expErr:    ErrInvalidRollappID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, err := NewChainID(test.rollappId)
			require.ErrorIs(t, err, test.expErr)
			if err == nil {
				require.Equal(t, test.revision, id.GetRevisionNumber())
			}
		})
	}
}

func TestNewChainID_Blumbus(t *testing.T) {
	testCases := []struct {
		chainID       string
		expectedError bool
	}{
		{"aib_777777-1", false},
		{"artpav_100000-1", false},
		{"artpav_100003-1", false},
		{"bah_29993410-1", false},
		{"bordel_909-1", false},
		{"crynux_10000-1", false},
		{"dummy_404-1", false},
		{"dyml_3200-1", false},
		{"dyml_3201-1", false},
		{"dyml_3202-1", false},
		{"gogol_090909090-1", false}, // Special case
		{"gotem_43438-1", false},
		{"jkjk_77778787-1", false},
		{"lucyna_504-1", false},
		{"mande_1807-1", false},
		{"nebulafi_1336830-1", false},
		{"nebulafi_1336831-1", false},
		{"nebulafi_1336836-1", false},
		{"nim_9999-1", false},
		{"nimtestnet_1122-1", false},
		{"playplayplay_4344343-1", false},
		{"rivalz_1231-1", false},
		{"rollappx_696969-1", false},
		{"rollappx_700000-1", false},
		{"rollappx_700001-1", false},
		{"rollappy_700002-1", false},
		{"rolv_1111111-1", false},
		{"rolx_100004-1", false},
		{"shaolinshaolin_60000021-1", false},
		{"tramp_999-1", false},
		{"trump_6900069-1", false},
		{"trumpnumba1_101010-1", false},
		{"wasmdymgas_20000003-1", false},
		{"yanks_898989-1", false},
	}

	for _, tc := range testCases {
		t.Run(tc.chainID, func(t *testing.T) {
			_, err := NewChainID(tc.chainID)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
