package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

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
	bankKeeper    dymnstypes.BankKeeper
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

	// others
	dymnskeeper.ClearCaches()
}

func (s *KeeperTestSuite) priceDenom() string {
	return s.dymNsKeeper.GetParams(s.ctx).Price.PriceDenom
}

func (s *KeeperTestSuite) mintToModuleAccount(amount int64) {
	err := s.bankKeeper.MintCoins(s.ctx,
		dymnstypes.ModuleName,
		sdk.Coins{sdk.NewCoin(s.priceDenom(), sdk.NewInt(amount))},
	)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) mintToModuleAccount2(amount sdkmath.Int) {
	err := s.bankKeeper.MintCoins(s.ctx,
		dymnstypes.ModuleName,
		sdk.Coins{sdk.NewCoin(s.priceDenom(), amount)},
	)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) mintToAccount(bech32Account string, amount int64) {
	s.mintToModuleAccount(amount)
	err := s.bankKeeper.SendCoinsFromModuleToAccount(s.ctx,
		dymnstypes.ModuleName,
		sdk.MustAccAddressFromBech32(bech32Account),
		sdk.Coins{sdk.NewCoin(s.priceDenom(), sdk.NewInt(amount))},
	)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) mintToAccount2(bech32Account string, amount sdkmath.Int) {
	s.mintToModuleAccount2(amount)
	err := s.bankKeeper.SendCoinsFromModuleToAccount(s.ctx,
		dymnstypes.ModuleName,
		sdk.MustAccAddressFromBech32(bech32Account),
		sdk.Coins{sdk.NewCoin(s.priceDenom(), amount)},
	)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) balance(bech32Account string) int64 {
	return s.bankKeeper.GetBalance(s.ctx,
		sdk.MustAccAddressFromBech32(bech32Account),
		s.priceDenom(),
	).Amount.Int64()
}

func (s *KeeperTestSuite) balance2(bech32Account string) sdkmath.Int {
	return s.bankKeeper.GetBalance(s.ctx,
		sdk.MustAccAddressFromBech32(bech32Account),
		s.priceDenom(),
	).Amount
}

func (s *KeeperTestSuite) moduleBalance() int64 {
	return s.balance(dymNsModuleAccAddr.String())
}

func (s *KeeperTestSuite) moduleBalance2() sdkmath.Int {
	return s.balance2(dymNsModuleAccAddr.String())
}

func (s *KeeperTestSuite) persistRollApp(ra rollapp) {
	s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
		RollappId:    ra.rollAppId,
		Owner:        ra.owner,
		Bech32Prefix: ra.bech32,
	})

	for _, alias := range ra.aliases {
		err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, ra.rollAppId, alias)
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

func (s *KeeperTestSuite) setBuyOrderWithFunctionsAfter(buyOrder dymnstypes.BuyOrder) {
	err := s.dymNsKeeper.SetBuyOrder(s.ctx, buyOrder)
	s.Require().NoError(err)

	err = s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx,
		buyOrder.AssetId, buyOrder.AssetType, buyOrder.Id,
	)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) setDymNameWithFunctionsAfter(dymName dymnstypes.DymName) {
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName))
	s.Require().NoError(s.dymNsKeeper.AfterDymNameOwnerChanged(s.ctx, dymName.Name))
	s.Require().NoError(s.dymNsKeeper.AfterDymNameConfigChanged(s.ctx, dymName.Name))
}

//

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
	aliases   []string
}

func newRollApp(rollAppId string) *rollapp {
	return &rollapp{
		rollAppId: rollAppId,
		owner:     testAddr(0).bech32(),
		bech32:    "",
		aliases:   nil,
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
	r.aliases = append(r.aliases, alias)
	if r.alias == "" {
		r.alias = alias
	}
	return r
}

//

type reqRollApp struct {
	s         *KeeperTestSuite
	rollAppId string
}

func (s *KeeperTestSuite) requireRollApp(rollAppId string) *reqRollApp {
	return &reqRollApp{
		s:         s,
		rollAppId: rollAppId,
	}
}

func (m reqRollApp) HasAlias(aliases ...string) {
	if len(aliases) == 0 {
		panic("must provide at least one alias")
	}
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

type reqAlias struct {
	s     *KeeperTestSuite
	alias string
}

func (s *KeeperTestSuite) requireAlias(alias string) *reqAlias {
	return &reqAlias{
		s:     s,
		alias: alias,
	}
}

func (m reqAlias) NotInUse() {
	gotRollAppId, found := m.s.dymNsKeeper.GetRollAppIdByAlias(m.s.ctx, m.alias)
	m.s.Require().False(found)
	m.s.Require().Empty(gotRollAppId)
}

func (m reqAlias) LinkedToRollApp(rollAppId string) {
	gotRollAppId, found := m.s.dymNsKeeper.GetRollAppIdByAlias(m.s.ctx, m.alias)
	m.s.Require().True(found)
	m.s.Require().Equal(rollAppId, gotRollAppId)
}

//

type sellOrderBuilder struct {
	s *KeeperTestSuite
	//
	assetId   string
	assetType dymnstypes.AssetType
	expiry    int64
	minPrice  int64
	sellPrice *int64
	// bid
	bidder    string
	bidAmount int64
	params    []string
}

func (s *KeeperTestSuite) newDymNameSellOrder(dymName string) *sellOrderBuilder {
	return s.newSellOrder(dymName, dymnstypes.TypeName)
}

func (s *KeeperTestSuite) newAliasSellOrder(alias string) *sellOrderBuilder {
	return s.newSellOrder(alias, dymnstypes.TypeAlias)
}

func (s *KeeperTestSuite) newSellOrder(assetId string, assetType dymnstypes.AssetType) *sellOrderBuilder {
	return &sellOrderBuilder{
		s:         s,
		assetId:   assetId,
		assetType: assetType,
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
		AssetId:    b.assetId,
		AssetType:  b.assetType,
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

//

type buyOrderBuilder struct {
	s *KeeperTestSuite
	//
	id         string
	assetId    string
	assetType  dymnstypes.AssetType
	buyer      string
	offerPrice int64
	params     []string
}

func (s *KeeperTestSuite) newAliasBuyOrder(buyer, alias, rollAppId string) *buyOrderBuilder {
	ob := s.newBuyOrder(buyer, alias, dymnstypes.TypeAlias)
	ob.params = []string{rollAppId}
	return ob
}

func (s *KeeperTestSuite) newBuyOrder(buyer, assetId string, assetType dymnstypes.AssetType) *buyOrderBuilder {
	return &buyOrderBuilder{
		s:          s,
		assetId:    assetId,
		assetType:  assetType,
		buyer:      buyer,
		offerPrice: 1,
		params:     nil,
	}
}

func (b *buyOrderBuilder) WithID(id string) *buyOrderBuilder {
	b.id = id
	return b
}

func (b *buyOrderBuilder) WithOfferPrice(p int64) *buyOrderBuilder {
	b.offerPrice = p
	return b
}

func (b *buyOrderBuilder) BuildP() *dymnstypes.BuyOrder {
	bo := b.Build()
	return &bo
}

func (b *buyOrderBuilder) Build() dymnstypes.BuyOrder {
	bo := dymnstypes.BuyOrder{
		Id:         b.id,
		AssetId:    b.assetId,
		AssetType:  b.assetType,
		Params:     b.params,
		Buyer:      b.buyer,
		OfferPrice: dymnsutils.TestCoin(b.offerPrice),
	}

	return bo
}
