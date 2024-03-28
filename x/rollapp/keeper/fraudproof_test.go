package keeper_test

import (
	_ "embed"
	"encoding/json"

	fraudtypes "github.com/cosmos/cosmos-sdk/baseapp"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

//go:embed testdata/fraud_proof_0.json
var fraudProof0 []byte

func (suite *RollappTestSuite) TestRunFraudProof() {
	suite.SetupTest()

	// TODO: document where the file came from, how it was generated
	// TODO: test validation too

	fp := abcitypes.FraudProof{}
	err := json.Unmarshal(fraudProof0, &fp)
	suite.NoError(err)

	fraud := fraudtypes.FraudProof{}
	err = fraud.FromABCI(fp)
	suite.NoError(err)
	err = suite.app.RollappKeeper.RunFraudProof(fraud)
	suite.NoError(err)
}
