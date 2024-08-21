package apptesting

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// 10^18 multiplier
var EXP = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

// Default sender for amm messages
var sender = sdk.MustAccAddressFromBech32(alice)

var DefaultAcctFunds sdk.Coins = sdk.NewCoins(
	sdk.NewCoin("adym", EXP.Mul(sdk.NewInt(1_000_000))),
	sdk.NewCoin("foo", EXP.Mul(sdk.NewInt(1_000_000))),
	sdk.NewCoin("bar", EXP.Mul(sdk.NewInt(1_000_000))),
	sdk.NewCoin("baz", EXP.Mul(sdk.NewInt(1_000_000))),
)

var DefaultPoolParams = balancer.PoolParams{
	SwapFee: sdk.NewDec(0),
	ExitFee: sdk.NewDec(0),
}

var DefaultPoolAssets = []balancer.PoolAsset{
	{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", EXP.Mul(sdk.NewInt(500))),
	},
	{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("adym", EXP.Mul(sdk.NewInt(500))),
	},
}

// PrepareCustomPool sets up a Balancer pool with an array of assets and given parameters
// This is the generic method called by other PreparePool wrappers
// It funds the sender account with DefaultAcctFunds
func (s *KeeperTestHelper) PrepareCustomPool(assets []balancer.PoolAsset, params balancer.PoolParams) uint64 {
	s.FundAcc(sender, DefaultAcctFunds)

	msg := balancer.NewMsgCreateBalancerPool(sender, params, assets, "")
	poolId, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, msg)
	s.NoError(err)
	return poolId
}

// PrepareDefaultPool sets up a pool with default pool assets and parameters.
func (s *KeeperTestHelper) PrepareDefaultPool() uint64 {
	poolId := s.PrepareCustomPool(DefaultPoolAssets, DefaultPoolParams)

	spotPrice, err := s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, "foo", "adym")
	s.NoError(err)
	s.Equal(sdk.NewDec(1).String(), spotPrice.String())

	return poolId
}

// PreparePoolWithCoins returns a pool consisted of given coins with equal weight and default pool parameters.
func (s *KeeperTestHelper) PreparePoolWithCoins(coins sdk.Coins) uint64 {
	poolAssets := coinsToAssets(coins)
	return s.PrepareCustomPool(poolAssets, DefaultPoolParams)
}

// PreparePoolWithPoolParams sets up a pool with given poolParams and default pool assets.
func (s *KeeperTestHelper) PreparePoolWithPoolParams(poolParams balancer.PoolParams) uint64 {
	return s.PrepareCustomPool(DefaultPoolAssets, poolParams)
}

// PrepareCustomPoolFromCoins sets up a Balancer pool with an array of coins and given parameters
// The coins are converted to pool assets where each asset has a weight of 1.
func (s *KeeperTestHelper) PrepareCustomPoolFromCoins(coins sdk.Coins, params balancer.PoolParams) uint64 {
	poolAssets := coinsToAssets(coins)
	return s.PrepareCustomPool(poolAssets, params)
}

func coinsToAssets(coins sdk.Coins) []balancer.PoolAsset {
	var poolAssets []balancer.PoolAsset
	for _, coin := range coins {
		poolAsset := balancer.PoolAsset{
			Weight: sdk.NewInt(1),
			Token:  coin,
		}
		poolAssets = append(poolAssets, poolAsset)
	}
	return poolAssets
}

func (s *KeeperTestHelper) RunBasicSwap(poolId uint64, from string, swapIn sdk.Coin, outDenom string) {
	msg := gammtypes.MsgSwapExactAmountIn{
		Sender:            from,
		Routes:            []poolmanagertypes.SwapAmountInRoute{{PoolId: poolId, TokenOutDenom: outDenom}},
		TokenIn:           swapIn,
		TokenOutMinAmount: sdk.ZeroInt(),
	}

	gammMsgServer := gammkeeper.NewMsgServerImpl(s.App.GAMMKeeper)
	_, err := gammMsgServer.SwapExactAmountIn(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err)
}

func (s *KeeperTestHelper) RunBasicExit(poolId uint64, shares sdk.Int, from string) (out sdk.Coins) {
	msg := gammtypes.MsgExitPool{
		Sender:        from,
		PoolId:        poolId,
		ShareInAmount: shares,
		TokenOutMins:  sdk.NewCoins(),
	}

	gammMsgServer := gammkeeper.NewMsgServerImpl(s.App.GAMMKeeper)
	res, err := gammMsgServer.ExitPool(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err)
	return res.TokenOut
}

// RunBasicJoin joins the pool with 10% of the total pool shares
func (s *KeeperTestHelper) RunBasicJoin(poolId uint64, from string) (shares sdk.Int, cost sdk.Coins) {
	pool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
	s.Require().NoError(err)

	totalPoolShare := pool.GetTotalShares()
	msg := gammtypes.MsgJoinPool{
		Sender:         from,
		PoolId:         poolId,
		ShareOutAmount: totalPoolShare.Quo(sdk.NewInt(10)),
		TokenInMaxs:    sdk.NewCoins(),
	}

	gammMsgServer := gammkeeper.NewMsgServerImpl(s.App.GAMMKeeper)
	res, err := gammMsgServer.JoinPool(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err)

	return res.ShareOutAmount, res.TokenIn
}
