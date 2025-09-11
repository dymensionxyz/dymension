package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	testutil "github.com/dymensionxyz/dymension/v3/testutil/math"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var (
	dymDenomMetadata = banktypes.Metadata{
		Description: "Denom of the Hub",
		Base:        "adym",
		Display:     "DYM",
		Name:        "DYM",
		Symbol:      "adym",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "adym",
				Exponent: 0,
			}, {
				Denom:    "DYM",
				Exponent: 18,
			},
		},
	}

	usdcDenomMetadata = banktypes.Metadata{
		Description: "Denom of the USDC",
		Base:        "usdc",
		Display:     "USDC",
		Name:        "USDC",
		Symbol:      "usdc",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "usdc",
				Exponent: 0,
			}, {
				Denom:    "USDC",
				Exponent: 6,
			},
		},
	}
)

// TestValidateRollappPreconditions tests the validation of rollapp preconditions in the CreatePlan function.
// It covers the following cases:
// - Rollapp is missing genesis checksum
// - Rollapp is already launched
// - Happy path with valid rollapp
func (s *KeeperTestSuite) TestValidateRollappPreconditions() {
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()
	allocation := math.NewInt(100).MulRaw(1e18)
	liquidityPart := types.DefaultParams().MinLiquidityPart

	s.Run("MissingGenesisChecksum", func() {
		s.SetupTest()
		rollappId := s.CreateDefaultRollapp()
		k := s.App.IROKeeper

		rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
		rollapp.GenesisInfo.GenesisChecksum = ""
		s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)

		_, err := k.CreatePlan(s.Ctx, "adym", allocation, time.Hour, time.Now(), true, false, rollapp, curve, incentives, liquidityPart, time.Hour, 0)
		s.Require().Error(err)
	})

	s.Run("AlreadyLaunched", func() {
		s.SetupTest()
		rollappId := s.CreateDefaultRollapp()
		k := s.App.IROKeeper

		rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
		rollapp.GenesisInfo.GenesisChecksum = "aaaaaa"
		rollapp.Launched = true
		s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)

		_, err := k.CreatePlan(s.Ctx, "adym", allocation, time.Hour, time.Now(), true, false, rollapp, curve, incentives, liquidityPart, time.Hour, 0)
		s.Require().Error(err)
	})

	s.Run("HappyPath", func() {
		s.SetupTest()
		rollappId := s.CreateDefaultRollapp()
		k := s.App.IROKeeper

		rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
		rollapp.GenesisInfo.GenesisChecksum = "aaaaaa"
		rollapp.Launched = false
		s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)

		_, err := k.CreatePlan(s.Ctx, "adym", allocation, time.Hour, time.Now(), true, false, rollapp, curve, incentives, liquidityPart, time.Hour, 0)
		s.Require().NoError(err)
	})
}

// TestCreatePlan tests the CreatePlan method of the keeper.
// It creates a plan for a given rollapp, tests that creating a plan for the same rollapp fails,
// creates a plan for a different rollapp and tests that the plan IDs increase.
//
// It also tests that the plan exists, that the plan can be retrieved by ID, and that the
// module account has the expected creation fee.
//
// Finally, it tests that the genesis info is sealed after creating a plan.
func (s *KeeperTestSuite) TestCreatePlan() {
	rollappId := s.CreateDefaultRollapp()
	rollappId2 := s.CreateDefaultRollapp()

	k := s.App.IROKeeper
	curve := types.DefaultBondingCurve()
	incentives := types.DefaultIncentivePlanParams()
	allocation := math.NewInt(100).MulRaw(1e18)
	liquidityPart := types.DefaultParams().MinLiquidityPart

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	planId, err := k.CreatePlan(s.Ctx, "adym", allocation, time.Hour, time.Now(), true, false, rollapp, curve, incentives, liquidityPart, time.Hour, 0)
	s.Require().NoError(err)

	// creating a plan for same rollapp should fail
	_, err = k.CreatePlan(s.Ctx, "adym", allocation, time.Hour, time.Now(), true, false, rollapp, curve, incentives, liquidityPart, time.Hour, 0)
	s.Require().Error(err)

	// create plan for different rollappID. test last planId increases
	rollapp2, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId2)
	planId2, err := k.CreatePlan(s.Ctx, "adym", allocation, time.Hour, time.Now(), true, false, rollapp2, curve, incentives, liquidityPart, time.Hour, 0)
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
	expectedCreationFee := plan.BondingCurve.Cost(math.ZeroInt(), s.App.IROKeeper.GetParams(s.Ctx).CreationFee)
	balances := s.App.BankKeeper.GetAllBalances(s.Ctx, plan.GetAddress())
	s.Require().Len(balances, 1)
	s.Require().Equal(expectedCreationFee, balances[0].Amount)

	// assert that genesis info is sealed
	rollapp, _ = s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	s.Require().True(rollapp.GenesisInfo.Sealed)
}

// TestMintAllocation tests that MintAllocation works correctly.
//
// It creates a rollapp and then uses MintAllocation to mint a certain amount of
// tokens. It then asserts that the denom metadata is registered, the virtual
// frontier bank contract is created and the coins have been minted.
func (s *KeeperTestSuite) TestMintAllocation() {
	rollappId := s.CreateDefaultRollapp()

	k := s.App.IROKeeper

	allocatedAmount := math.NewInt(10).MulRaw(1e18)
	expectedBaseDenom := types.IRODenom(rollappId)

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	minted, err := k.MintAllocation(s.Ctx, allocatedAmount, rollapp.RollappId, rollapp.GenesisInfo.NativeDenom.Base, uint64(rollapp.GenesisInfo.NativeDenom.Exponent))
	s.Require().NoError(err)

	// assert denom metadata registered
	_, found := s.App.BankKeeper.GetDenomMetaData(s.Ctx, expectedBaseDenom)
	s.Require().True(found)

	// assert virtual frontier bank contract created
	found = s.App.EvmKeeper.HasVirtualFrontierBankContractByDenom(s.Ctx, expectedBaseDenom)
	s.Require().True(found)

	// assert coins minted
	s.Assert().True(allocatedAmount.Equal(minted.Amount))
	coins := s.App.BankKeeper.GetSupply(s.Ctx, expectedBaseDenom)
	s.Require().Equal(allocatedAmount, coins.Amount)
}

// TestStandardLaunchIROPreconditions tests the preconditions for fair launch IRO creation
// should be 100% IRO allocation
// should be registered and whitelisted liquidity denom
func (s *KeeperTestSuite) TestStandardLaunchIROPreconditions() {
	s.Run("happy flow", func() {
		s.SetupTest()
		s.App.BankKeeper.SetDenomMetaData(s.Ctx, dymDenomMetadata)
		gammParams := s.App.GAMMKeeper.GetParams(s.Ctx)
		gammParams.AllowedPoolCreationDenoms = append(gammParams.AllowedPoolCreationDenoms, "adym")
		s.App.GAMMKeeper.SetParams(s.Ctx, gammParams)

		rollappId := s.CreateStandardLaunchRollapp()
		rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
		owner := rollapp.Owner

		// Fund owner with liquidity denom for creation fee
		s.FundAcc(sdk.MustAccAddressFromBech32(owner), sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100_000).MulRaw(1e18))))

		_, err := s.msgServer.CreateStandardLaunchPlan(s.Ctx, &types.MsgCreateStandardLaunchPlan{
			RollappId:      rollappId,
			Owner:          owner,
			TradingEnabled: true,
			LiquidityDenom: "adym",
		})
		s.Require().NoError(err)
	})

	s.Run("not 100% IRO allocation", func() {
		s.SetupTest()
		s.App.BankKeeper.SetDenomMetaData(s.Ctx, dymDenomMetadata)
		gammParams := s.App.GAMMKeeper.GetParams(s.Ctx)
		gammParams.AllowedPoolCreationDenoms = append(gammParams.AllowedPoolCreationDenoms, "adym")
		s.App.GAMMKeeper.SetParams(s.Ctx, gammParams)

		rollappId := s.CreateDefaultRollapp()
		k := s.App.IROKeeper
		params := k.GetParams(s.Ctx)

		rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
		owner := rollapp.Owner

		partialAmount := params.StandardLaunch.AllocationAmount.QuoRaw(2) // Half allocation
		rollapp.GenesisInfo.InitialSupply = params.StandardLaunch.AllocationAmount
		rollapp.GenesisInfo.GenesisAccounts = &rollapptypes.GenesisAccounts{
			Accounts: []rollapptypes.GenesisAccount{
				{
					Address: k.GetModuleAccountAddress(),
					Amount:  partialAmount, // Only partial amount for IRO
				},
				{
					Address: "dym1otheraddress", // Some other account with remaining
					Amount:  partialAmount,
				},
			},
		}
		s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)

		// Fund owner with liquidity denom for creation fee
		s.FundAcc(sdk.MustAccAddressFromBech32(owner), sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100_000).MulRaw(1e18))))
		_, err := s.msgServer.CreateStandardLaunchPlan(s.Ctx, &types.MsgCreateStandardLaunchPlan{
			RollappId:      rollappId,
			Owner:          owner,
			TradingEnabled: true,
			LiquidityDenom: "adym",
		})
		s.Require().Error(err)
	})
}

func (s *KeeperTestSuite) TestStandardLaunch_TargetRaise() {
	s.SetupTest()
	s.App.BankKeeper.SetDenomMetaData(s.Ctx, dymDenomMetadata)
	gammParams := s.App.GAMMKeeper.GetParams(s.Ctx)
	gammParams.AllowedPoolCreationDenoms = append(gammParams.AllowedPoolCreationDenoms, "adym")
	s.App.GAMMKeeper.SetParams(s.Ctx, gammParams)

	rollappId := s.CreateStandardLaunchRollapp()
	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	owner := rollapp.Owner

	// Fund owner with liquidity denom for creation fee
	s.FundAcc(sdk.MustAccAddressFromBech32(owner), sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100_000).MulRaw(1e18))))
	res, err := s.msgServer.CreateStandardLaunchPlan(s.Ctx, &types.MsgCreateStandardLaunchPlan{
		RollappId:      rollappId,
		Owner:          owner,
		TradingEnabled: true,
		LiquidityDenom: "adym",
	})
	s.Require().NoError(err)

	k := s.App.IROKeeper
	params := k.GetParams(s.Ctx)
	plan := k.MustGetPlan(s.Ctx, res.PlanId)
	actualRaised := plan.BondingCurve.Cost(math.ZeroInt(), plan.MaxAmountToSell)
	err = testutil.ApproxEqualRatio(params.StandardLaunch.TargetRaise.Amount, actualRaised, 0.01) // 1% tolerance
	s.Require().NoError(err)
}

// TestStandardLaunchTargetRaiseConversion tests the conversion from params' target raise
// to liquidity denom and validates that the bonding curve is correctly configured
func (s *KeeperTestSuite) TestStandardLaunch_TargetRaiseConversion() {
	s.SetupTest()
	s.App.BankKeeper.SetDenomMetaData(s.Ctx, usdcDenomMetadata)
	gammParams := s.App.GAMMKeeper.GetParams(s.Ctx)
	gammParams.AllowedPoolCreationDenoms = append(gammParams.AllowedPoolCreationDenoms, "usdc")
	s.App.GAMMKeeper.SetParams(s.Ctx, gammParams)

	// create pool with usdc and adym with price 1 usdc = 5 adym
	priceRatio := int64(5)
	s.PreparePoolWithCoins(sdk.NewCoins(
		sdk.NewCoin("usdc", math.NewInt(1_000_000).MulRaw(1e6)),
		sdk.NewCoin("adym", math.NewInt(1_000_000).MulRaw(priceRatio).MulRaw(1e18)),
	))

	rollappId := s.CreateStandardLaunchRollapp()
	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	owner := rollapp.Owner

	// Fund owner with liquidity denom for creation fee
	s.FundAcc(sdk.MustAccAddressFromBech32(owner), sdk.NewCoins(sdk.NewCoin("usdc", math.NewInt(100_000).MulRaw(1e6))))
	res, err := s.msgServer.CreateStandardLaunchPlan(s.Ctx, &types.MsgCreateStandardLaunchPlan{
		RollappId:      rollappId,
		Owner:          owner,
		TradingEnabled: true,
		LiquidityDenom: "usdc",
	})
	s.Require().NoError(err)

	k := s.App.IROKeeper
	params := k.GetParams(s.Ctx)
	plan := k.MustGetPlan(s.Ctx, res.PlanId)

	usdcTargetRaise := plan.BondingCurve.Cost(math.ZeroInt(), plan.MaxAmountToSell)
	scaledUsdcTargetRaise := types.ScaleFromBase(usdcTargetRaise, 6)
	scaledTargetRaise := types.ScaleFromBase(params.StandardLaunch.TargetRaise.Amount, 18)
	err = testutil.ApproxEqualRatio(scaledUsdcTargetRaise.MulInt64(priceRatio), scaledTargetRaise, 0.01) // 1% tolerance
	s.Require().NoError(err)
}

// TestStandardLaunch_TargetRaise_InUSDC tests the case where the target raise is in USDC.
// we create fair launch rollapp with DYM as the liquidity denom
func (s *KeeperTestSuite) TestStandardLaunch_TargetRaise_InUSDC() {
	s.App.BankKeeper.SetDenomMetaData(s.Ctx, dymDenomMetadata)
	gammParams := s.App.GAMMKeeper.GetParams(s.Ctx)
	gammParams.AllowedPoolCreationDenoms = append(gammParams.AllowedPoolCreationDenoms, "adym")
	s.App.GAMMKeeper.SetParams(s.Ctx, gammParams)

	// create pool with usdc and adym with price 1 usdc = 5 adym
	priceRatio := int64(5)
	s.PreparePoolWithCoins(sdk.NewCoins(
		sdk.NewCoin("usdc", math.NewInt(1_000_000).MulRaw(1e6)),
		sdk.NewCoin("adym", math.NewInt(1_000_000).MulRaw(priceRatio).MulRaw(1e18)),
	))

	iroParams := s.App.IROKeeper.GetParams(s.Ctx)
	iroParams.StandardLaunch.TargetRaise = sdk.NewCoin("usdc", math.NewInt(5_000).MulRaw(1e6)) // 5K USDC
	s.App.IROKeeper.SetParams(s.Ctx, iroParams)

	rollappId := s.CreateStandardLaunchRollapp()
	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
	owner := rollapp.Owner

	// Fund owner with liquidity denom for creation fee
	s.FundAcc(sdk.MustAccAddressFromBech32(owner), sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100_000).MulRaw(1e18))))
	res, err := s.msgServer.CreateStandardLaunchPlan(s.Ctx, &types.MsgCreateStandardLaunchPlan{
		RollappId:      rollappId,
		Owner:          owner,
		TradingEnabled: true,
		LiquidityDenom: "adym",
	})
	s.Require().NoError(err)

	k := s.App.IROKeeper
	params := k.GetParams(s.Ctx)
	plan := k.MustGetPlan(s.Ctx, res.PlanId)

	dymRaised := plan.BondingCurve.Cost(math.ZeroInt(), plan.MaxAmountToSell)
	scaledDymRaised := types.ScaleFromBase(dymRaised, 18)
	scaledTargetRaise := types.ScaleFromBase(params.StandardLaunch.TargetRaise.Amount, 6)
	err = testutil.ApproxEqualRatio(scaledTargetRaise, scaledDymRaised.QuoInt64(priceRatio), 0.01) // 1% tolerance
	s.Require().NoError(err)
}
