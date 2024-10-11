package keeper_test

import (
	"slices"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestUpdateWhitelistedRelayers() {
	rollappId, pk := suite.CreateDefaultRollapp()
	defaultSequencerAddress := suite.CreateSequencer(suite.Ctx, rollappId, pk)
	relayers := []string{sample.AccAddress(), sample.AccAddress()}

	testCase := []struct {
		name        string
		msg         types.MsgUpdateWhitelistedRelayers
		expectedErr error
	}{
		{
			name: "valid",
			msg: types.MsgUpdateWhitelistedRelayers{
				Creator:  defaultSequencerAddress,
				Relayers: relayers,
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCase {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.UpdateWhitelistedRelayers(suite.Ctx, &tc.msg)
			if tc.expectedErr != nil {
				suite.Require().ErrorIs(err, tc.expectedErr)
			} else {
				suite.Require().NoError(err)
				seq, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, tc.msg.Creator)
				slices.Sort(tc.msg.Relayers)
				suite.Require().Equal(tc.msg.Relayers, seq.WhitelistedRelayers)
			}
		})
	}
}
