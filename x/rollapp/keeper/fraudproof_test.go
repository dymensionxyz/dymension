package keeper_test

import (
	"encoding/json"
	"log"
	"os"

	fraudtypes "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/stretchr/testify/assert"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

func (suite *RollappTestSuite) TestFraudProof() {
	suite.SetupTest()

	file, err := os.Open("/Users/mtsitrin/Applications/dymension/rollapp-evm/fraudProof_rollapp_with_tx.json")
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	// Decode the JSON-encoded data into your struct
	jsonDecoder := json.NewDecoder(file)
	fp := abcitypes.FraudProof{}
	err = jsonDecoder.Decode(&fp)
	if err != nil {
		log.Fatalf("failed decoding JSON: %s", err)
	}

	fraud := fraudtypes.FraudProof{}
	err = fraud.FromABCI(fp)
	assert.NoError(suite.T(), err)
	err = suite.app.RollappKeeper.VerifyFraudProof(fraud)
	assert.NoError(suite.T(), err)
}
