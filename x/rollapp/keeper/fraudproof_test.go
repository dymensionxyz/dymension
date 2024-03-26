package keeper_test

import (
	_ "embed"
	"encoding/json"

	fraudtypes "github.com/cosmos/cosmos-sdk/baseapp"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

//go:embed testdata/fraud_proof_0.json
var embeddedFp []byte

func (suite *RollappTestSuite) TestRunFraudProof() {
	suite.SetupTest()

	// TODO: test validation too

	fp := abcitypes.FraudProof{}
	err := json.Unmarshal(embeddedFp, &fp)
	suite.NoError(err)

	fraud := fraudtypes.FraudProof{}
	err = fraud.FromABCI(fp)
	suite.NoError(err)
	err = suite.app.RollappKeeper.RunFraudProof(fraud)
	suite.NoError(err)
}
