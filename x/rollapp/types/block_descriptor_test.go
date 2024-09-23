package types_test

import (
	"testing"
	time "time"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func TestBlockDescriptorsValidate(t *testing.T) {
	testCases := []struct {
		name    string
		bds     types.BlockDescriptors
		expPass bool
	}{
		{
			name: "valid block descriptors",
			bds: types.BlockDescriptors{
				BD: []types.BlockDescriptor{
					{
						Height:    1,
						Timestamp: time.Now(),
					},
					{
						Height:    2,
						Timestamp: time.Now(),
					},
				},
			},
			expPass: true,
		},
		{
			name: "invalid block descriptor",
			bds: types.BlockDescriptors{
				BD: []types.BlockDescriptor{
					{
						Height:    1,
						Timestamp: time.Now(),
					},
					{
						Height: 2,
					},
				},
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.bds.Validate()
			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBlockDescriptorValidate(t *testing.T) {
	testCases := []struct {
		name    string
		bd      types.BlockDescriptor
		expPass bool
	}{
		{
			name: "valid block descriptor",
			bd: types.BlockDescriptor{
				Height:    1,
				Timestamp: time.Now(),
			},
			expPass: true,
		},
		{
			name: "invalid block descriptor",
			bd: types.BlockDescriptor{
				Height: 1,
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.bd.Validate()
			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
