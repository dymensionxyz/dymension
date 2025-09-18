package keeper_test

import (
	_ "embed"
	"encoding/json"
	"strconv"
	"time"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var (
	//go:embed testdata/tee/policy_values.json
	policyValues string
	//go:embed testdata/tee/query.rego
	policyQuery string
	//go:embed testdata/tee/policy.rego
	policyStructure string
	//go:embed testdata/tee/example_response.json
	exampleResponse string
)

type ExampleResponse struct {
	Result struct {
		Token string `json:"token"`
		Nonce struct {
			RollappId  string `json:"rollapp_id"`
			CurrHeight string `json:"curr_height"`
		} `json:"nonce"`
	} `json:"result"`
}

func (s *RollappTestSuite) TestValidateAttestation() {
	// s.T().Skip()
	s.SetupTest()
	s.k().SetParams(s.Ctx, types.DefaultParams().WithTeeConfig(types.TEEConfig{
		PolicyValues:    policyValues,
		PolicyQuery:     policyQuery,
		PolicyStructure: policyStructure,
	}))

	var res ExampleResponse
	err := json.Unmarshal([]byte(exampleResponse), &res)
	s.Require().NoError(err)

	token := res.Result.Token

	rollappId := res.Result.Nonce.RollappId
	currHeight, err := strconv.ParseUint(res.Result.Nonce.CurrHeight, 10, 64)
	s.Require().NoError(err)

	// TODO: use proper data and test
	nonce := types.TEENonce{
		RollappId:       rollappId,
		CurrHeight:      currHeight,
		FinalizedHeight: 0,
	}

	s.Ctx = s.Ctx.WithBlockTime(time.Date(2025, 9, 18, 9, 46, 0, 0, time.UTC))
	// s.Ctx = s.Ctx.WithBlockTime(time.Now())
	err = s.k().ValidateAttestation(s.Ctx, nonce.Hash(), token)
	s.Require().NoError(err)
}
