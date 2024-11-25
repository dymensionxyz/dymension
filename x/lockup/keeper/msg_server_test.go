package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"

	"github.com/dymensionxyz/dymension/v3/x/lockup/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

func (suite *KeeperTestSuite) TestMsgLockTokens() {
	type param struct {
		coinsToLock         sdk.Coins
		lockOwner           sdk.AccAddress
		duration            time.Duration
		coinsInOwnerAddress sdk.Coins
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "creation of lock via lockTokens",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: true,
		},
		{
			name: "locking more coins than are in the address",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 20)},       // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()

			suite.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)
			// fund address with lock fee
			baseDenom, _ := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
			suite.FundAcc(test.param.lockOwner, sdk.NewCoins(sdk.NewCoin(baseDenom, types.DefaultLockFee)))

			msgServer := keeper.NewMsgServerImpl(suite.App.LockupKeeper)
			c := sdk.WrapSDKContext(suite.Ctx)
			_, err := msgServer.LockTokens(c, types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))

			if test.expectPass {
				suite.Require().NoError(err)

				// Check Locks
				locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Len(locks, 1)
				suite.Require().Equal(locks[0].Coins, test.param.coinsToLock)

				// check accumulation store is correctly updated
				accum := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
					LockQueryType: types.ByDuration,
					Denom:         "stake",
					Duration:      test.param.duration,
				})
				suite.Require().Equal(accum.String(), "10")

				// add more tokens to lock via LockTokens
				suite.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)
				// fund address with lock fee
				suite.FundAcc(test.param.lockOwner, sdk.NewCoins(sdk.NewCoin(baseDenom, types.DefaultLockFee)))

				_, err = msgServer.LockTokens(sdk.WrapSDKContext(suite.Ctx), types.NewMsgLockTokens(test.param.lockOwner, locks[0].Duration, test.param.coinsToLock))
				suite.Require().NoError(err)

				// check locks after adding tokens to lock
				locks, err = suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Len(locks, 1)
				suite.Require().Equal(locks[0].Coins, test.param.coinsToLock.Add(test.param.coinsToLock...))

				// check accumulation store is correctly updated
				accum = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
					LockQueryType: types.ByDuration,
					Denom:         "stake",
					Duration:      test.param.duration,
				})
				suite.Require().Equal(accum.String(), "20")

			} else {
				// Fail simple lock token
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMsgBeginUnlocking() {
	type param struct {
		coinsToLock         sdk.Coins
		coinsToUnlock       sdk.Coins
		lockOwner           sdk.AccAddress
		duration            time.Duration
		coinsInOwnerAddress sdk.Coins
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
		isPartial  bool
	}{
		{
			name: "unlock full amount of tokens via begin unlock",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
				coinsToUnlock:       sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: true,
		},
		{
			name: "unlock partial amount of tokens via begin unlock",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
				coinsToUnlock:       sdk.Coins{sdk.NewInt64Coin("stake", 5)},        // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			isPartial:  true,
			expectPass: true,
		},
		{
			name: "unlock zero amount of tokens via begin unlock",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
				coinsToUnlock:       sdk.Coins{},                                    // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: true,
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		suite.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)
		// fund address with lock fee
		baseDenom, _ := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
		suite.FundAcc(test.param.lockOwner, sdk.NewCoins(sdk.NewCoin(baseDenom, types.DefaultLockFee)))

		msgServer := keeper.NewMsgServerImpl(suite.App.LockupKeeper)
		goCtx := sdk.WrapSDKContext(suite.Ctx)
		resp, err := msgServer.LockTokens(goCtx, types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))
		suite.Require().NoError(err)

		unlockingResponse, err := msgServer.BeginUnlocking(goCtx, types.NewMsgBeginUnlocking(test.param.lockOwner, resp.ID, test.param.coinsToUnlock))

		if test.expectPass {
			suite.Require().NoError(err)
			suite.AssertEventEmitted(suite.Ctx, types.TypeEvtBeginUnlock, 1)
			suite.Require().True(unlockingResponse.Success)
			if test.isPartial {
				suite.Require().Equal(unlockingResponse.UnlockingLockID, resp.ID+1)
			} else {
				suite.Require().Equal(unlockingResponse.UnlockingLockID, resp.ID)
			}
		} else {
			suite.Require().Error(err)
			suite.AssertEventEmitted(suite.Ctx, types.TypeEvtBeginUnlock, 0)
		}
	}
}

func (suite *KeeperTestSuite) TestMsgBeginUnlockingAll() {
	type param struct {
		coinsToLock         sdk.Coins
		lockOwner           sdk.AccAddress
		duration            time.Duration
		coinsInOwnerAddress sdk.Coins
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "unlock all lockups",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: true,
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		suite.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)
		// fund address with lock fee
		baseDenom, _ := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
		suite.FundAcc(test.param.lockOwner, sdk.NewCoins(sdk.NewCoin(baseDenom, types.DefaultLockFee)))

		msgServer := keeper.NewMsgServerImpl(suite.App.LockupKeeper)
		c := sdk.WrapSDKContext(suite.Ctx)
		_, err := msgServer.LockTokens(c, types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))
		suite.Require().NoError(err)

		_, err = msgServer.BeginUnlockingAll(c, types.NewMsgBeginUnlockingAll(test.param.lockOwner))

		if test.expectPass {
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestMsgEditLockup() {
	type param struct {
		coinsToLock       sdk.Coins
		isSyntheticLockup bool
		lockOwner         sdk.AccAddress
		duration          time.Duration
		newDuration       time.Duration
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "edit lockups by duration",
			param: param{
				coinsToLock:       sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup: false,
				lockOwner:         sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:          time.Second,
				newDuration:       time.Second * 2,
			},
			expectPass: true,
		},
		{
			name: "edit lockups by lesser duration",
			param: param{
				coinsToLock:       sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup: false,
				lockOwner:         sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:          time.Second,
				newDuration:       time.Second / 2,
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, test.param.lockOwner, test.param.coinsToLock)
		suite.Require().NoError(err)
		// fund address with lock fee
		baseDenom, _ := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
		suite.FundAcc(test.param.lockOwner, sdk.NewCoins(sdk.NewCoin(baseDenom, types.DefaultLockFee)))

		msgServer := keeper.NewMsgServerImpl(suite.App.LockupKeeper)
		c := sdk.WrapSDKContext(suite.Ctx)
		resp, err := msgServer.LockTokens(c, types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))
		suite.Require().NoError(err)

		_, err = msgServer.ExtendLockup(c, types.NewMsgExtendLockup(test.param.lockOwner, resp.ID, test.param.newDuration))

		if test.expectPass {
			suite.Require().NoError(err, test.name)
		} else {
			suite.Require().Error(err, test.name)
		}
	}
}

func (suite *KeeperTestSuite) TestMsgForceUnlock() {
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	defaultLockAmount := sdk.NewInt(1000000000)

	tests := []struct {
		name                      string
		forceUnlockAllowedAddress types.Params
		postLockSetup             func()
		forceUnlockAmount         sdk.Int
		expectPass                bool
	}{
		{
			"happy path",
			types.Params{ForceUnlockAllowedAddresses: []string{addr1.String()}},
			func() {},
			defaultLockAmount,
			true,
		},
		{
			"partial unlock",
			types.Params{ForceUnlockAllowedAddresses: []string{addr1.String()}},
			func() {},
			// try force unlocking half of locked amount
			defaultLockAmount.Quo(sdk.NewInt(2)),
			true,
		},
		{
			"force unlock more than what we have locked",
			types.Params{ForceUnlockAllowedAddresses: []string{addr1.String()}},
			func() {},
			// try force more than the locked amount
			defaultLockAmount.Add(sdk.NewInt(1)),
			false,
		},
		{
			"params with different address",
			types.Params{ForceUnlockAllowedAddresses: []string{addr2.String()}},
			func() {},
			defaultLockAmount,
			false,
		},
		{
			"param with multiple addresses ",
			types.Params{ForceUnlockAllowedAddresses: []string{addr1.String(), addr2.String()}},
			func() {},
			defaultLockAmount,
			true,
		},
	}

	for _, test := range tests {
		// set up test
		suite.SetupTest()
		suite.App.LockupKeeper.SetParams(suite.Ctx, test.forceUnlockAllowedAddress)

		// prepare pool for superfluid staking cases
		poolId := suite.PreparePoolWithCoins(sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000_000_000)), sdk.NewCoin("foo", sdk.NewInt(5000))))

		// lock tokens
		msgServer := keeper.NewMsgServerImpl(suite.App.LockupKeeper)
		c := sdk.WrapSDKContext(suite.Ctx)

		poolDenom := gammtypes.GetPoolShareDenom(poolId)
		coinsToLock := sdk.Coins{sdk.NewCoin(poolDenom, defaultLockAmount)}
		suite.FundAcc(addr1, coinsToLock)

		// fund address with lock fee
		baseDenom, _ := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
		suite.FundAcc(addr1, sdk.NewCoins(sdk.NewCoin(baseDenom, types.DefaultLockFee)))

		unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime
		resp, err := msgServer.LockTokens(c, types.NewMsgLockTokens(addr1, unbondingDuration, coinsToLock))
		suite.Require().NoError(err)

		// setup env after lock tokens
		test.postLockSetup()

		// test force unlock
		_, err = msgServer.ForceUnlock(c, types.NewMsgForceUnlock(addr1, resp.ID, sdk.Coins{sdk.NewCoin(poolDenom, test.forceUnlockAmount)}))
		if test.expectPass {
			suite.Require().NoError(err)

			// check that we have successfully force unlocked
			balanceAfterForceUnlock := suite.App.BankKeeper.GetBalance(suite.Ctx, addr1, poolDenom)
			suite.Require().Equal(test.forceUnlockAmount, balanceAfterForceUnlock.Amount)
		} else {
			suite.Require().Error(err)

			// check that we have successfully force unlocked
			balanceAfterForceUnlock := suite.App.BankKeeper.GetBalance(suite.Ctx, addr1, poolDenom)
			suite.Require().NotEqual(test.forceUnlockAmount, balanceAfterForceUnlock.Amount)
			return
		}
	}
}
