package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite
	ctx           sdk.Context
	chainId       string
	now           time.Time
	dymNsKeeper   dymnskeeper.Keeper
	rollAppKeeper rollappkeeper.Keeper
	bankKeeper    dymnskeeper.BankKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupSuite() {
}

func (s *KeeperTestSuite) SetupTest() {
	const chainId = "dymension_1100-1"

	now := time.Now().UTC()
	dk, bk, rk, ctx := testkeeper.DymNSKeeper(s.T())
	ctx = ctx.WithBlockTime(now).WithChainID(chainId)

	// enable trading
	moduleParams := dk.GetParams(ctx)
	moduleParams.Misc.EnableTradingName = true
	moduleParams.Misc.EnableTradingAlias = true
	err := dk.SetParams(ctx, moduleParams)
	s.Require().NoError(err)

	// set
	s.ctx = ctx
	s.chainId = chainId
	s.now = now
	s.dymNsKeeper = dk
	s.rollAppKeeper = rk
	s.bankKeeper = bk
}

func (s *KeeperTestSuite) mintToModuleAccount(amount int64) {
	err := s.bankKeeper.MintCoins(s.ctx, dymnstypes.ModuleName, dymnsutils.TestCoins(amount))
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) mintToAccount(bech32Account string, amount int64) {
	err := s.bankKeeper.MintCoins(s.ctx, dymnstypes.ModuleName, dymnsutils.TestCoins(amount))
	s.Require().NoError(err)
	err = s.bankKeeper.SendCoinsFromModuleToAccount(s.ctx,
		dymnstypes.ModuleName,
		sdk.MustAccAddressFromBech32(bech32Account),
		dymnsutils.TestCoins(amount),
	)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) balance(bech32Account string) int64 {
	return s.bankKeeper.GetBalance(s.ctx, sdk.MustAccAddressFromBech32(bech32Account), s.dymNsKeeper.GetParams(s.ctx).Price.PriceDenom).Amount.Int64()
}

func (s *KeeperTestSuite) moduleBalance() int64 {
	return s.balance(dymNsModuleAccAddr.String())
}

func (s *KeeperTestSuite) persistRollApp(ra rollapp) {
	s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
		RollappId:    ra.rollAppId,
		Owner:        ra.owner,
		Bech32Prefix: ra.bech32,
	})

	if ra.alias != "" {
		err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, ra.rollAppId, ra.alias)
		s.Require().NoError(err)
	}
}

func (s *KeeperTestSuite) moduleParams() dymnstypes.Params {
	return s.dymNsKeeper.GetParams(s.ctx)
}

func (s *KeeperTestSuite) updateModuleParams(f func(dymnstypes.Params) dymnstypes.Params) {
	params := s.moduleParams()
	params = f(params)
	err := s.dymNsKeeper.SetParams(s.ctx, params)
	s.Require().NoError(err)
}

//

func (s *KeeperTestSuite) requireRollApp(rollAppId string) *reqRollApp {
	return &reqRollApp{
		s:         s,
		rollAppId: rollAppId,
	}
}

func (s *KeeperTestSuite) requireErrorContains(err error, errMsgContains string) {
	s.Require().NotEmpty(errMsgContains, "mis-configured test")
	s.Require().Error(err)
	s.Require().Contains(err.Error(), errMsgContains)
}

func (s *KeeperTestSuite) requireErrorFContains(f func() error, contains string) {
	s.requireErrorContains(f(), contains)
}

//

type rollapp struct {
	rollAppId string
	owner     string
	bech32    string
	alias     string
}

func newRollApp(rollAppId string) *rollapp {
	return &rollapp{
		rollAppId: rollAppId,
		owner:     testAddr(0).bech32(),
		bech32:    "",
		alias:     "",
	}
}

func (r *rollapp) WithOwner(owner string) *rollapp {
	r.owner = owner
	return r
}

func (r *rollapp) WithBech32(bech32 string) *rollapp {
	r.bech32 = bech32
	return r
}

func (r *rollapp) WithAlias(alias string) *rollapp {
	r.alias = alias
	return r
}

//

type reqRollApp struct {
	s         *KeeperTestSuite
	rollAppId string
}

func (m reqRollApp) HasAlias(aliases ...string) {
	for _, alias := range aliases {
		rollAppId, found := m.s.dymNsKeeper.GetRollAppIdByAlias(m.s.ctx, alias)
		m.s.Require().True(found)
		m.s.Require().Equal(m.rollAppId, rollAppId)
	}
}

func (m reqRollApp) HasNoAlias() {
	alias, found := m.s.dymNsKeeper.GetAliasByRollAppId(m.s.ctx, m.rollAppId)
	m.s.Require().False(found)
	m.s.Require().Empty(alias)
}

//

type sellOrderBuilder struct {
	s *KeeperTestSuite
	//
	goodsId   string
	orderType dymnstypes.OrderType
	expiry    int64
	minPrice  int64
	sellPrice *int64
	// bid
	bidder    string
	bidAmount int64
	params    []string
}

func (s *KeeperTestSuite) newDymNameSellOrder(dymName string) *sellOrderBuilder {
	return s.newSellOrder(dymName, dymnstypes.NameOrder)
}

func (s *KeeperTestSuite) newAliasSellOrder(alias string) *sellOrderBuilder {
	return s.newSellOrder(alias, dymnstypes.AliasOrder)
}

func (s *KeeperTestSuite) newSellOrder(goodsId string, orderType dymnstypes.OrderType) *sellOrderBuilder {
	return &sellOrderBuilder{
		s:         s,
		goodsId:   goodsId,
		orderType: orderType,
		expiry:    s.now.Add(time.Second).Unix(),
		minPrice:  0,
	}
}

func (b *sellOrderBuilder) WithMinPrice(minPrice int64) *sellOrderBuilder {
	b.minPrice = minPrice
	return b
}

func (b *sellOrderBuilder) Expired() *sellOrderBuilder {
	return b.WithExpiry(b.s.now.Add(-time.Second).Unix())
}

func (b *sellOrderBuilder) WithExpiry(epoch int64) *sellOrderBuilder {
	b.expiry = epoch
	return b
}

func (b *sellOrderBuilder) WithSellPrice(sellPrice int64) *sellOrderBuilder {
	b.sellPrice = &sellPrice
	return b
}

func (b *sellOrderBuilder) WithDymNameBid(bidder string, bidAmount int64) *sellOrderBuilder {
	b.bidder = bidder
	b.bidAmount = bidAmount
	return b
}

func (b *sellOrderBuilder) WithAliasBid(bidder string, bidAmount int64, rollAppId string) *sellOrderBuilder {
	b.bidder = bidder
	b.bidAmount = bidAmount
	b.params = []string{rollAppId}
	return b
}

func (b *sellOrderBuilder) BuildP() *dymnstypes.SellOrder {
	so := b.Build()
	return &so
}

func (b *sellOrderBuilder) Build() dymnstypes.SellOrder {
	so := dymnstypes.SellOrder{
		GoodsId:    b.goodsId,
		Type:       b.orderType,
		ExpireAt:   b.expiry,
		MinPrice:   dymnsutils.TestCoin(b.minPrice),
		SellPrice:  nil,
		HighestBid: nil,
	}

	if b.sellPrice != nil {
		so.SellPrice = dymnsutils.TestCoinP(*b.sellPrice)
	}

	if b.bidder != "" {
		so.HighestBid = &dymnstypes.SellOrderBid{
			Bidder: b.bidder,
			Price:  dymnsutils.TestCoin(b.bidAmount),
			Params: b.params,
		}
	}

	return so
}
