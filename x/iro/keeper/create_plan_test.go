package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func (s *KeeperTestSuite) TestValidateRollappPreconditions_MissingGenesisInfo() {
	rollappId := s.CreateDefaultRollapp()
	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)

	// test missing genesis checksum
	rollapp.GenesisInfo.GenesisChecksum = ""
	s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)
	_, err := k.CreatePlan(s.Ctx, sdk.NewInt(100), time.Now(), time.Now().Add(time.Hour), rollapp, curve, incentives)
	s.Require().Error(err)

	// test already launched
	rollapp.GenesisInfo.GenesisChecksum = "aaaaaa"
	rollapp.Launched = true
	s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)
	_, err = k.CreatePlan(s.Ctx, sdk.NewInt(100), time.Now(), time.Now().Add(time.Hour), rollapp, curve, incentives)
	s.Require().Error(err)
	rollapp.Launched = false

	// add check for happy path
	s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)
	_, err = k.CreatePlan(s.Ctx, sdk.NewInt(100), time.Now(), time.Now().Add(time.Hour), rollapp, curve, incentives)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestCreatePlan() {
	rollappId := s.CreateDefaultRollapp()
	rollappId2 := s.CreateDefaultRollapp()

	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, sdk.NewInt(100), time.Now(), time.Now().Add(time.Hour), rollapp, curve, incentives)
	s.Require().NoError(err)

	// creating a a plan for same rollapp should fail
	_, err = k.CreatePlan(s.Ctx, sdk.NewInt(100), time.Now(), time.Now().Add(time.Hour), rollapp, curve, incentives)
	s.Require().Error(err)

	// create plan for different rollappID. test last planId increases
	rollapp2, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId2)
	planId2, err := k.CreatePlan(s.Ctx, sdk.NewInt(100), time.Now(), time.Now().Add(time.Hour), rollapp2, curve, incentives)
	s.Require().NoError(err)
	s.Require().Greater(planId2, planId)

	// test plan exists
	plan, found := k.GetPlanByRollapp(s.Ctx, rollappId)
	s.Require().True(found)
	s.Require().Equal(planId, fmt.Sprintf("%d", plan.Id))

	plan, found = k.GetPlanByRollapp(s.Ctx, rollappId2)
	s.Require().True(found)
	s.Require().Equal(planId2, fmt.Sprintf("%d", plan.Id))

	// test get all plans
	plans := k.GetAllPlans(s.Ctx)
	s.Require().Len(plans, 2)

	ok := s.App.AccountKeeper.HasAccount(s.Ctx, plan.GetAddress())
	s.Require().True(ok)

	// test module account has the expected creation fee
	expectedCreationFee := s.App.IROKeeper.GetParams(s.Ctx).CreationFee
	balances := s.App.BankKeeper.GetAllBalances(s.Ctx, plan.GetAddress())
	s.Require().Len(balances, 1)
	s.Require().Equal(expectedCreationFee, balances[0].Amount)

	// assert that genesis info is sealed
	rollapp, _ = s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	s.Require().True(rollapp.GenesisInfo.Sealed)
}

func (s *KeeperTestSuite) TestMintAllocation() {
	rollappId := s.CreateDefaultRollapp()

	k := s.App.IROKeeper

	allocatedAmount := sdk.NewInt(10).MulRaw(1e18)
	expectedBaseDenom := fmt.Sprintf("%s_%s", types.IROTokenPrefix, rollappId)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	minted, err := k.MintAllocation(s.Ctx, allocatedAmount, rollapp.RollappId, rollapp.GenesisInfo.NativeDenom.Base, uint64(rollapp.GenesisInfo.NativeDenom.Exponent))
	s.Require().NoError(err)

	// assert denom metadata registered
	_, found := s.App.BankKeeper.GetDenomMetaData(s.Ctx, expectedBaseDenom)
	s.Require().True(found)

	// assert coins minted
	s.Assert().True(allocatedAmount.Equal(minted.Amount))
	coins := s.App.BankKeeper.GetSupply(s.Ctx, expectedBaseDenom)
	s.Require().Equal(allocatedAmount, coins.Amount)
}
