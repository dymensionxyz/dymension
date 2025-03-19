package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeper "github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func (s *KeeperTestSuite) setV1IRO() {
	var (
		fooCoin           = sdk.NewCoin("foo", math.NewInt(100_000).MulRaw(1e18))
		curve             = types.DefaultBondingCurve()
		defaultIncentives = types.DefaultIncentivePlanParams()
	)

	// simulate old IRO plan
	curve.LiquidityDenomDecimals = 0
	curve.RollappDenomDecimals = 0

	plan := types.NewPlan(1, "rollapp1", "", fooCoin, curve, 0, defaultIncentives, math.LegacyOneDec(), 0, 0)
	plan.MaxAmountToSell = math.ZeroInt()
	plan.LiquidityPart = math.LegacyDec{}
	plan.VestingPlan = types.IROVestingPlan{}

	s.App.IROKeeper.SetPlan(s.Ctx, plan)
}

func (s *KeeperTestSuite) TestMigrate1to2() {
	s.setV1IRO()

	m := keeper.NewMigrator(*s.App.IROKeeper)
	err := m.Migrate1to2(s.Ctx)
	s.Require().NoError(err)

	// check plan
	plan, ok := s.App.IROKeeper.GetPlanByRollapp(s.Ctx, "rollapp1")
	s.Require().True(ok, "plan not found")

	s.Require().False(plan.MaxAmountToSell.IsZero(), "max amount to sell not set correctly")

	s.Require().True(plan.LiquidityPart.Equal(math.LegacyOneDec()), "liquidity part not set correctly")

	s.Require().NotEmpty(plan.LiquidityDenom, "liquidity denom not set correctly")

	s.Require().NotZero(plan.BondingCurve.LiquidityDenomDecimals, "bonding curve liquidity denom decimals not set correctly")
	s.Require().NotZero(plan.BondingCurve.RollappDenomDecimals, "bonding curve rollapp denom decimals not set correctly")

	s.Require().True(plan.TradingEnabled, "trading enabled not set correctly")
}
