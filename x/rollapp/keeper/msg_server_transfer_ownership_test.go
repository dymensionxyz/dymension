package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (suite *RollappTestSuite) TestTransferOwnership() {
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
			name: "Transfer rollapp ownership: failed, frozen rollapp",
			request: types.NewMsgTransferOwnership(
				alice, bob, rollappId,
			),
			malleate: func(rollapp types.Rollapp) types.Rollapp {
				rollapp.Frozen = true
				return rollapp
			},
			expError: types.ErrRollappFrozen,
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
		suite.Run(tc.name, func() {
			rollapp := types.Rollapp{
				RollappId:   rollappId,
				Owner:       alice,
				GenesisInfo: *mockGenesisInfo,
			}

			if tc.malleate != nil {
				rollapp = tc.malleate(rollapp)
			}

			suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

			goCtx := sdk.WrapSDKContext(suite.Ctx)
			_, err := suite.msgServer.TransferOwnership(goCtx, tc.request)
			if tc.expError == nil {
				suite.Require().NoError(err)
				resp, err := suite.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{RollappId: tc.request.RollappId})
				suite.Require().NoError(err)
				suite.Equal(tc.expRollapp, resp.Rollapp)
			} else {
				suite.ErrorIs(err, tc.expError)
			}
		})
	}
}
