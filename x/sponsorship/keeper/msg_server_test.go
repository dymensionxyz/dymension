package keeper_test

import (
	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (s *KeeperTestSuite) TestUpdateParams() {
	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	testCases := []struct {
		name  string
		msg   types.MsgUpdateParams
		error error
	}{
		{
			name: "valid",
			msg: types.MsgUpdateParams{
				Authority: authority,
				NewParams: types.Params{
					MinAllocationWeight: types.DefaultMinAllocationWeight,
					MinVotingPower:      types.DefaultMinVotingPower,
				},
			},
			error: nil,
		},
		{
			name: "invalid authority",
			msg: types.MsgUpdateParams{
				Authority: apptesting.CreateRandomAccounts(1)[0].String(), // random address
				NewParams: types.Params{
					MinAllocationWeight: types.DefaultMinAllocationWeight,
					MinVotingPower:      types.DefaultMinVotingPower,
				},
			},
			error: sdkerrors.ErrorInvalidSigner,
		},
		{
			name: "invalid params",
			msg: types.MsgUpdateParams{
				Authority: authority,
				NewParams: types.Params{
					MinAllocationWeight: math.NewInt(101), // > 100%
					MinVotingPower:      types.DefaultMinVotingPower,
				},
			},
			error: types.ErrInvalidParams,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			oldParams := s.GetParams()

			// Call UpdateParams
			resp, err := s.msgServer.UpdateParams(s.Ctx, &tc.msg)

			// Check the results
			switch {
			case tc.error != nil:
				s.Require().ErrorIs(err, tc.error)
				s.Require().Nil(resp)

				newParams := s.GetParams()
				s.Require().Equal(oldParams, newParams)

				s.AssertEventEmitted(s.Ctx, proto.MessageName(new(types.EventUpdateParams)), 0)
			case tc.error == nil:
				s.Require().NoError(err)
				s.Require().Equal(new(types.MsgUpdateParamsResponse), resp)

				newParams := s.GetParams()
				s.Require().Equal(tc.msg.NewParams, newParams)

				s.AssertEventEmitted(s.Ctx, proto.MessageName(new(types.EventUpdateParams)), 1)
			}
		})
	}
}
