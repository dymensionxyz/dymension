package keeper_test

import (
	_ "embed"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var (
	//go:embed testdata/tee/policy_values.json
	policyValues string
	//go:embed testdata/tee/query.rego
	policyQuery string
	//go:embed testdata/tee/policy.rego
	policyStructure string
)

func (s *RollappTestSuite) TestValidateAttestation() {
	s.T().Skip()
	s.SetupTest()
	s.k().SetParams(s.Ctx, types.DefaultParams().WithTeeConfig(types.TEEConfig{
		PolicyValues:    policyValues,
		PolicyQuery:     policyQuery,
		PolicyStructure: policyStructure,
	}))

	// TODO: use proper data and test
	token := "token"
	nonce := types.TEENonce{
		RollappId:          "rollapp_id",
		CurrHeight:         1,
		CurrStateRoot:      []byte("state_root"),
		FinalizedHeight:    1,
		FinalizedStateRoot: []byte("finalized_state_root"),
	}
	err := s.k().ValidateAttestation(s.Ctx, nonce.Hash(), token)
	s.Require().NoError(err)
}
