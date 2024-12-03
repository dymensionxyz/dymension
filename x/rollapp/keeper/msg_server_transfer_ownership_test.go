package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (s *RollappTestSuite) TestTransferOwnership() {
	const rollappId = "rollapp_1234-1"

	tests := []struct {
		name       string
		request    *types.MsgTransferOwnership
		malleate   func(rollapp types.Rollapp) types.Rollapp
		expError   error
		expRollapp types.Rollapp
	}{
		{
			name: "Transfer rollapp ownership: success",
			request: types.NewMsgTransferOwnership(
				alice, bob, rollappId,
			),
			expError: nil,
			expRollapp: types.Rollapp{
				Owner:       bob,
				RollappId:   rollappId,
				GenesisInfo: *mockGenesisInfo,
			},
		}, {
			name: "Transfer rollapp ownership: failed, rollapp not found",
			request: types.NewMsgTransferOwnership(
				alice, bob, "rollapp_1235-2",
			),
			expError: types.ErrUnknownRollappID,
		}, {
			name: "Transfer rollapp ownership: failed, same owner",
			request: types.NewMsgTransferOwnership(
				alice, alice, rollappId,
			),
			expError: types.ErrSameOwner,
		}, {
			name: "Transfer rollapp ownership: failed, unauthorized signer",
			request: types.NewMsgTransferOwnership(
				bob, alice, rollappId,
			),
			expError: types.ErrUnauthorizedSigner,
		}, {
			name: "Transfer rollapp ownership: failed, invalid current owner address",
			request: types.NewMsgTransferOwnership(
				"invalid_address", bob, rollappId,
			),
			expError: types.ErrInvalidRequest,
		}, {
			name: "Transfer rollapp ownership: failed, invalid new owner address",
			request: types.NewMsgTransferOwnership(
				alice, "invalid_address", rollappId,
			),
			expError: types.ErrInvalidRequest,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			rollapp := types.Rollapp{
				RollappId:   rollappId,
				Owner:       alice,
				GenesisInfo: *mockGenesisInfo,
			}

			if tc.malleate != nil {
				rollapp = tc.malleate(rollapp)
			}

			s.k().SetRollapp(s.Ctx, rollapp)

			goCtx := sdk.WrapSDKContext(s.Ctx)
			_, err := s.msgServer.TransferOwnership(goCtx, tc.request)
			if tc.expError == nil {
				s.Require().NoError(err)
				resp, err := s.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{RollappId: tc.request.RollappId})
				s.Require().NoError(err)
				s.Equal(tc.expRollapp, resp.Rollapp)
			} else {
				s.ErrorIs(err, tc.expError)
			}
		})
	}
}
