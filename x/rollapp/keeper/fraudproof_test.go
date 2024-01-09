package keeper_test

import (
	fraudtypes "github.com/dymensionxyz/dymension/app/fraudproof/types"
)

func (suite *RollappTestSuite) TestFraudProof() {
	suite.SetupTest()

	// // set rollapp
	// rollapp := types.Rollapp{
	// 	RollappId:     "rollapp1",
	// 	Creator:       alice,
	// 	Version:       3,
	// 	MaxSequencers: 1,
	// }
	// suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	fp := fraudtypes.FraudProof{}

	suite.app.RollappKeeper.VerifyFraudProof(fp)

}
