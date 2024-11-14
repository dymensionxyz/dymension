package types

import (
	"testing"

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
