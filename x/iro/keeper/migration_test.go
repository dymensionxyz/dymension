package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeper "github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func (s *KeeperTestSuite) setV1IRO(rollappId string) types.Plan {
	var (
		fooCoin           = sdk.NewCoin("foo", math.NewInt(100_000).MulRaw(1e18))
		curve             = types.DefaultBondingCurve()
		defaultIncentives = types.DefaultIncentivePlanParams()
	)

	// simulate old IRO plan
	curve.LiquidityDenomDecimals = 0
	curve.RollappDenomDecimals = 0

	plan := types.NewPlan(1, rollappId, "", fooCoin, curve, 0, defaultIncentives, math.LegacyOneDec(), 0, 0)
	plan.MaxAmountToSell = math.ZeroInt()
	plan.LiquidityPart = math.LegacyDec{}
	plan.VestingPlan = types.IROVestingPlan{}

	s.App.IROKeeper.SetPlan(s.Ctx, plan)
	return plan
}

func (s *KeeperTestSuite) TestMigrate1to2() {
	// set simple IRO plan
	_ = s.setV1IRO("rollapp1")

	// set settled IRO plan
	rollappDenom := "rollappSettledDenom"
	plan := s.setV1IRO("rollapp2")
	plan.SettledDenom = rollappDenom
	s.App.IROKeeper.SetPlan(s.Ctx, plan)

	// create balancer pool to simulate settled IRO plan
	poolId := s.PreparePoolWithCoins(sdk.NewCoins(
		sdk.NewCoin(rollappDenom, math.NewInt(1_000_000).MulRaw(1e18)),
		sdk.NewCoin("adym", math.NewInt(1_000_000).MulRaw(1e18)),
	))
	// assert fee token created
	_, err := s.App.TxFeesKeeper.GetFeeToken(s.Ctx, rollappDenom)
	s.Require().NoError(err)

	m := keeper.NewMigrator(*s.App.IROKeeper)
	err = m.Migrate1to2(s.Ctx)
	s.Require().NoError(err)

	// check plan1
	plan, ok := s.App.IROKeeper.GetPlanByRollapp(s.Ctx, "rollapp1")
	s.Require().True(ok, "plan should be found")

	s.Require().False(plan.MaxAmountToSell.IsZero(), "max amount to sell not set correctly")
	s.Require().True(plan.LiquidityPart.Equal(math.LegacyOneDec()), "liquidity part not set correctly")
	s.Require().NotEmpty(plan.LiquidityDenom, "liquidity denom not set correctly")
	s.Require().NotZero(plan.BondingCurve.LiquidityDenomDecimals, "bonding curve liquidity denom decimals not set correctly")
	s.Require().NotZero(plan.BondingCurve.RollappDenomDecimals, "bonding curve rollapp denom decimals not set correctly")
	s.Require().True(plan.TradingEnabled, "trading enabled not set correctly")

	// check poolID set for plan2
	plan, ok = s.App.IROKeeper.GetPlanByRollapp(s.Ctx, "rollapp2")
	s.Require().True(ok, "plan should be found")
	s.Require().Equal(poolId, plan.GraduatedPoolId, "poolID not set correctly")
}
