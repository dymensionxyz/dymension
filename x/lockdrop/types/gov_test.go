package types_test

import (
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/x/lockdrop/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestUpdateLockdropProposalMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		proposal *types.UpdateLockdropProposal
	}{
		{ // empty title
			proposal: &types.UpdateLockdropProposal{
				Title:       "",
				Description: "proposal to update pool incentives",
				Records:     []types.DistrRecord(nil),
			},
		},
		{ // empty description
			proposal: &types.UpdateLockdropProposal{
				Title:       "title",
				Description: "",
				Records:     []types.DistrRecord(nil),
			},
		},
		{ // empty records
			proposal: &types.UpdateLockdropProposal{
				Title:       "title",
				Description: "proposal to update pool incentives",
				Records:     []types.DistrRecord(nil),
			},
		},
		{ // one record
			proposal: &types.UpdateLockdropProposal{
				Title:       "title",
				Description: "proposal to update pool incentives",
				Records: []types.DistrRecord{
					{
						GaugeId: 1,
						Weight:  sdk.NewInt(1),
					},
				},
			},
		},
		{ // zero-weight record
			proposal: &types.UpdateLockdropProposal{
				Title:       "title",
				Description: "proposal to update pool incentives",
				Records: []types.DistrRecord{
					{
						GaugeId: 1,
						Weight:  sdk.NewInt(0),
					},
				},
			},
		},
		{ // two records
			proposal: &types.UpdateLockdropProposal{
				Title:       "title",
				Description: "proposal to update pool incentives",
				Records: []types.DistrRecord{
					{
						GaugeId: 1,
						Weight:  sdk.NewInt(1),
					},
					{
						GaugeId: 2,
						Weight:  sdk.NewInt(1),
					},
				},
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.proposal)
		require.NoError(t, err)
		decoded := types.UpdateLockdropProposal{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.proposal, decoded)
	}
}
