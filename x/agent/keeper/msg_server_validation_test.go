package keeper_test

import (
	"bytes"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func setupValidation(t *testing.T) (sdk.Context, *keeper.Keeper, types.MsgServer, string, string) {
	t.Helper()
	ctx, k, _ := setup(t)
	p := types.DefaultParams()
	p.ValidationRequestFee = sdk.NewCoin(p.ValidationRequestFee.Denom, math.ZeroInt())
	require.NoError(t, k.SetParams(ctx, p))
	validatorOwner, requester := owner(t), owner(t)
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: "validator", Owner: validatorOwner, Active: true}))
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: "subject", Owner: owner(t), Active: true, ActionSeq: 2}))
	return ctx, k, keeper.NewMsgServerImpl(*k), validatorOwner, requester
}

func TestValidationRequestResponseProgressiveAndQueries(t *testing.T) {
	ctx, k, ms, validatorOwner, requester := setupValidation(t)
	hash := bytes.Repeat([]byte{1}, 32)
	_, err := ms.RequestValidation(ctx, types.NewMsgRequestValidation(requester, "validator", "subject", 1, hash, "ipfs://request"))
	require.NoError(t, err)
	for i, score := range []uint32{40, 100} {
		res, err := ms.RespondValidation(ctx, types.NewMsgRespondValidation(validatorOwner, hash, score, "ipfs://response", nil, "reexecution"))
		require.NoError(t, err)
		require.Equal(t, uint64(i), res.Seq)
	}
	req, found := k.GetValidationRequest(ctx, hash)
	require.True(t, found)
	require.Equal(t, uint64(2), req.ResponseCount)
	resp, found := k.GetValidationResponse(ctx, hash, 0)
	require.True(t, found)
	require.Equal(t, uint32(40), resp.Response)
	got, err := k.ValidationRequest(ctx, &types.QueryValidationRequestRequest{RequestHash: hash})
	require.NoError(t, err)
	require.Equal(t, req, got.ValidationRequest)
	responses, err := k.ValidationResponses(ctx, &types.QueryValidationResponsesRequest{RequestHash: hash, Pagination: &query.PageRequest{Limit: 1, CountTotal: true}})
	require.NoError(t, err)
	require.Len(t, responses.ValidationResponses, 1)
	require.Equal(t, uint64(2), responses.Pagination.Total)
	byAgent, err := k.ValidationRequestsByAgent(ctx, &types.QueryValidationRequestsByAgentRequest{AgentId: "subject", Pagination: &query.PageRequest{Limit: 1}})
	require.NoError(t, err)
	require.Len(t, byAgent.ValidationRequests, 1)
}

func TestValidationRejectsInvalidRequestsAndResponses(t *testing.T) {
	ctx, k, ms, validatorOwner, requester := setupValidation(t)
	hash := bytes.Repeat([]byte{2}, 32)
	tests := []struct {
		name string
		msg  *types.MsgRequestValidation
		want error
	}{
		{"short hash", types.NewMsgRequestValidation(requester, "validator", "subject", 0, hash[:31], ""), types.ErrInvalidValidationHash},
		{"missing validator", types.NewMsgRequestValidation(requester, "missing", "subject", 0, hash, ""), types.ErrAgentNotFound},
		{"missing subject", types.NewMsgRequestValidation(requester, "validator", "missing", 0, hash, ""), types.ErrAgentNotFound},
		{"self", types.NewMsgRequestValidation(requester, "subject", "subject", 0, hash, ""), types.ErrSelfValidation},
		{"bad evidence", types.NewMsgRequestValidation(requester, "validator", "subject", 2, hash, ""), types.ErrInvalidEvidence},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) { _, err := ms.RequestValidation(ctx, tc.msg); require.ErrorIs(t, err, tc.want) })
	}
	_, err := ms.RequestValidation(ctx, types.NewMsgRequestValidation(requester, "validator", "subject", 0, hash, ""))
	require.NoError(t, err)
	_, err = ms.RequestValidation(ctx, types.NewMsgRequestValidation(requester, "validator", "subject", 0, hash, ""))
	require.ErrorIs(t, err, types.ErrValidationRequestExists)
	_, err = ms.RespondValidation(ctx, types.NewMsgRespondValidation(requester, hash, 50, "", nil, ""))
	require.ErrorIs(t, err, types.ErrUnauthorized)
	_, err = ms.RespondValidation(ctx, types.NewMsgRespondValidation(validatorOwner, hash, 101, "", nil, ""))
	require.ErrorIs(t, err, types.ErrInvalidValidationResponse)
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: "validator", Owner: validatorOwner, Active: false}))
	_, err = ms.RespondValidation(ctx, types.NewMsgRespondValidation(validatorOwner, hash, 50, "", nil, ""))
	require.ErrorIs(t, err, types.ErrValidatorInactive)
}

func TestValidationGenesisRoundTripAndRejectsOrphan(t *testing.T) {
	ctx, k, ms, validatorOwner, requester := setupValidation(t)
	hash := bytes.Repeat([]byte{3}, 32)
	_, err := ms.RequestValidation(ctx, types.NewMsgRequestValidation(requester, "validator", "subject", 0, hash, ""))
	require.NoError(t, err)
	_, err = ms.RespondValidation(ctx, types.NewMsgRespondValidation(validatorOwner, hash, 100, "", nil, ""))
	require.NoError(t, err)
	exported := keeper.ExportGenesis(ctx, k)
	require.NoError(t, exported.Validate())
	require.Len(t, exported.ValidationRequests, 1)
	require.Len(t, exported.ValidationResponses, 1)
	orphan := *types.DefaultGenesis()
	orphan.ValidationResponses = []types.ValidationResponse{{RequestHash: hash}}
	require.ErrorIs(t, orphan.Validate(), types.ErrValidationRequestNotFound)
	ctx2, k2, _ := setup(t)
	keeper.InitGenesis(ctx2, k2, *exported)
	indexed, err := k2.ValidationRequestsByAgent(ctx2, &types.QueryValidationRequestsByAgentRequest{AgentId: "subject"})
	require.NoError(t, err)
	require.Len(t, indexed.ValidationRequests, 1)
}

func TestValidationResponseIntegrityInvariant(t *testing.T) {
	ctx, k, ms, validatorOwner, requester := setupValidation(t)
	hash := bytes.Repeat([]byte{4}, 32)
	_, err := ms.RequestValidation(ctx, types.NewMsgRequestValidation(requester, "validator", "subject", 0, hash, ""))
	require.NoError(t, err)
	_, err = ms.RespondValidation(ctx, types.NewMsgRespondValidation(validatorOwner, hash, 100, "", nil, ""))
	require.NoError(t, err)
	require.NoError(t, keeper.InvariantValidationResponseIntegrity(*k)(ctx))

	ctx2, k2, _ := setup(t)
	orphanHash := bytes.Repeat([]byte{5}, 32)
	keeper.InitGenesis(ctx2, k2, types.GenesisState{Params: types.DefaultParams(), ValidationResponses: []types.ValidationResponse{{RequestHash: orphanHash}}})
	require.Error(t, keeper.InvariantValidationResponseIntegrity(*k2)(ctx2))
}

type ValidationFeeTestSuite struct {
	apptesting.KeeperTestHelper
	msgServer types.MsgServer
}

func TestValidationFeeTestSuite(t *testing.T) { suite.Run(t, new(ValidationFeeTestSuite)) }
func (s *ValidationFeeTestSuite) SetupTest() {
	s.App = apptesting.Setup(s.T())
	s.Ctx = s.App.NewContext(false)
	s.msgServer = keeper.NewMsgServerImpl(*s.App.AgentKeeper)
}

func (s *ValidationFeeTestSuite) seed() (sdk.Coin, sdk.AccAddress, string) {
	p, err := s.App.AgentKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	_, _, validator := testdata.KeyTestPubAddr()
	_, _, subject := testdata.KeyTestPubAddr()
	_, _, requester := testdata.KeyTestPubAddr()
	s.Require().NoError(s.App.AgentKeeper.SetAgent(s.Ctx, types.Agent{Id: "validator", Owner: validator.String(), Active: true}))
	s.Require().NoError(s.App.AgentKeeper.SetAgent(s.Ctx, types.Agent{Id: "subject", Owner: subject.String(), Active: true, ActionSeq: 1}))
	return p.ValidationRequestFee, requester, validator.String()
}

func (s *ValidationFeeTestSuite) TestRequestFeeBurned() {
	fee, requester, _ := s.seed()
	s.FundAcc(requester, sdk.NewCoins(sdk.NewCoin(fee.Denom, fee.Amount.MulRaw(2))))
	beforeBal := s.App.BankKeeper.GetBalance(s.Ctx, requester, fee.Denom)
	beforeSupply := s.App.BankKeeper.GetSupply(s.Ctx, fee.Denom)
	_, err := s.msgServer.RequestValidation(s.Ctx, types.NewMsgRequestValidation(requester.String(), "validator", "subject", 0, bytes.Repeat([]byte{6}, 32), ""))
	s.Require().NoError(err)
	s.Require().Equal(beforeBal.Amount.Sub(fee.Amount), s.App.BankKeeper.GetBalance(s.Ctx, requester, fee.Denom).Amount)
	s.Require().Equal(beforeSupply.Amount.Sub(fee.Amount), s.App.BankKeeper.GetSupply(s.Ctx, fee.Denom).Amount)
}

func (s *ValidationFeeTestSuite) TestTransactionRollbackRestoresFee() {
	fee, requester, _ := s.seed()
	s.FundAcc(requester, sdk.NewCoins(sdk.NewCoin(fee.Denom, fee.Amount.MulRaw(2))))
	before := s.App.BankKeeper.GetBalance(s.Ctx, requester, fee.Denom)
	cacheCtx, _ := s.Ctx.CacheContext()
	cacheServer := keeper.NewMsgServerImpl(*s.App.AgentKeeper)
	_, err := cacheServer.RequestValidation(cacheCtx, types.NewMsgRequestValidation(requester.String(), "validator", "subject", 0, bytes.Repeat([]byte{7}, 32), ""))
	s.Require().NoError(err)
	// A later message failure causes BaseApp to discard this cache context.
	s.Require().Equal(before, s.App.BankKeeper.GetBalance(s.Ctx, requester, fee.Denom))
	_, found := s.App.AgentKeeper.GetValidationRequest(s.Ctx, bytes.Repeat([]byte{7}, 32))
	s.Require().False(found)
}
