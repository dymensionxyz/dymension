package keeper_test

import (
	_ "embed"
	"encoding/json"
	"strconv"
	"time"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var (
	//go:embed testdata/tee/confidential_space_root.pem
	gcpRootCertificate string

	//go:embed testdata/tee/insecure_policy_values.json
	policyValues string
	//go:embed testdata/tee/insecure_query.rego
	policyQuery string
	//go:embed testdata/tee/insecure_policy.rego
	policyStructure string

	// TODO: embed response and reactivate test
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
	s.T().Skip("Requires a real response from GCP: need to setup again because pr https://github.com/dymensionxyz/dymension/pull/2059 changed nonce")
	s.SetupTest()
	s.k().SetParams(s.Ctx, types.DefaultParams().WithTeeConfig(types.TEEConfig{
		PolicyValues:    policyValues,
		PolicyQuery:     policyQuery,
		PolicyStructure: policyStructure,
		GcpRootCertPem:  gcpRootCertificate,
	}))

	var res ExampleResponse
	err := json.Unmarshal([]byte(exampleResponse), &res)
	s.Require().NoError(err)

	token := res.Result.Token

	rollappId := res.Result.Nonce.RollappId
	currHeight, err := strconv.ParseUint(res.Result.Nonce.CurrHeight, 10, 64)
	s.Require().NoError(err)

	nonce := types.TEENonce{
		RollappId:       rollappId,
		CurrHeight:      currHeight,
		HubChainId:      s.Ctx.ChainID(),
		FinalizedHeight: 0,
	}

	s.Ctx = s.Ctx.WithBlockTime(time.Date(2025, 9, 18, 9, 47, 0, 0, time.UTC))
	err = s.k().ValidateAttestation(s.Ctx, nonce.Hash(), token)
	s.Require().NoError(err)
}
