package keeper_test

import (
	"slices"
	"sort"
	"testing"
	"time"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/dymensionxyz/dymension/v3/app/params"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite

	anchorCtx sdk.Context
	ctx       sdk.Context

	chainId string
	now     time.Time

	cdc codec.BinaryCodec

	dymNsKeeper   dymnskeeper.Keeper
	rollAppKeeper rollappkeeper.Keeper
	bankKeeper    dymnstypes.BankKeeper

	rollappStoreKey storetypes.StoreKey
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupSuite() {
}

func (s *KeeperTestSuite) SetupTest() {
	var ctx sdk.Context

	var cdc codec.BinaryCodec

	var dk dymnskeeper.Keeper
	var bk dymnstypes.BankKeeper
	var rk *rollappkeeper.Keeper

	var rollappStoreKey storetypes.StoreKey

	{
		// initialization
		dymNsStoreKey := sdk.NewKVStoreKey(dymnstypes.StoreKey)
		dymNsMemStoreKey := storetypes.NewMemoryStoreKey(dymnstypes.MemStoreKey)

		authStoreKey := sdk.NewKVStoreKey(authtypes.StoreKey)

		bankStoreKey := sdk.NewKVStoreKey(banktypes.StoreKey)

		rollappStoreKey = sdk.NewKVStoreKey(rollapptypes.StoreKey)
		rollappMemStoreKey := storetypes.NewMemoryStoreKey(rollapptypes.MemStoreKey)

		db := tmdb.NewMemDB()
		stateStore := store.NewCommitMultiStore(db)
		stateStore.MountStoreWithDB(dymNsStoreKey, storetypes.StoreTypeIAVL, db)
		stateStore.MountStoreWithDB(dymNsMemStoreKey, storetypes.StoreTypeMemory, nil)
		stateStore.MountStoreWithDB(authStoreKey, storetypes.StoreTypeIAVL, db)
		stateStore.MountStoreWithDB(bankStoreKey, storetypes.StoreTypeIAVL, db)
		stateStore.MountStoreWithDB(rollappStoreKey, storetypes.StoreTypeIAVL, db)
		stateStore.MountStoreWithDB(rollappMemStoreKey, storetypes.StoreTypeMemory, nil)
		s.Require().NoError(stateStore.LoadLatestVersion())

		registry := codectypes.NewInterfaceRegistry()
		cdc = codec.NewProtoCodec(registry)

		dymNSParamsSubspace := typesparams.NewSubspace(cdc,
			dymnstypes.Amino,
			dymNsStoreKey,
			dymNsMemStoreKey,
			"DymNSParams",
		)

		rollappParamsSubspace := typesparams.NewSubspace(cdc,
			rollapptypes.Amino,
			rollappStoreKey,
			rollappMemStoreKey,
			"RollappParams",
		)

		authKeeper := authkeeper.NewAccountKeeper(
			cdc,
			authStoreKey,
			authtypes.ProtoBaseAccount,
			map[string][]string{
				banktypes.ModuleName:  {authtypes.Minter, authtypes.Burner},
				dymnstypes.ModuleName: {authtypes.Minter, authtypes.Burner},
			},
			params.AccountAddressPrefix,
			authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		)
		authtypes.RegisterInterfaces(registry)

		bk = bankkeeper.NewBaseKeeper(
			cdc,
			bankStoreKey,
			authKeeper,
			map[string]bool{},
			authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		)
		banktypes.RegisterInterfaces(registry)

		rk = rollappkeeper.NewKeeper(
			cdc,
			rollappStoreKey,
			rollappParamsSubspace,
			nil, nil, nil,
		)

		dk = dymnskeeper.NewKeeper(cdc,
			dymNsStoreKey,
			dymNSParamsSubspace,
			bk,
			rk,
		)

		ctx = sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

		s.Require().NoError(dk.SetParams(ctx, dymnstypes.DefaultParams()))
	}

	const chainId = "dymension_1100-1"

	// set
	s.chainId = chainId
	s.now = time.Now().UTC()
	s.anchorCtx = sdk.Context{}
	s.ctx = ctx.WithBlockTime(s.now).WithChainID(chainId)
	s.cdc = cdc
	s.dymNsKeeper = dk
	s.rollAppKeeper = *rk
	s.bankKeeper = bk
	s.rollappStoreKey = rollappStoreKey

	// custom
	s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
		moduleParams.Chains.AliasesOfChainIds = nil
		// force enable trading
		moduleParams.Misc.EnableTradingName = true
		moduleParams.Misc.EnableTradingAlias = true

		return moduleParams
	})

	// others
	dymnskeeper.ClearCaches()
}

// MakeAnchorContext copies the current context to the anchor context and convert current context into a branch context.
// This is useful when you want to set up a context and reuse multiple times.
// This is less expensive than call SetupTest.
func (s *KeeperTestSuite) MakeAnchorContext() {
	s.anchorCtx = s.ctx
	s.UseAnchorContext()
}

// UseAnchorContext clear any change to the current context and use the anchor context.
func (s *KeeperTestSuite) UseAnchorContext() {
	if s.anchorCtx.ChainID() == "" {
		panic("anchor context not set")
	}
	s.ctx, _ = s.anchorCtx.CacheContext()
	if gasMeter := s.ctx.GasMeter(); gasMeter != nil {
		gasMeter.RefundGas(gasMeter.GasConsumed(), "reset gas meter")
	}
	if blockGasMeter := s.ctx.BlockGasMeter(); blockGasMeter != nil {
		blockGasMeter.RefundGas(blockGasMeter.GasConsumed(), "reset block gas meter")
	}
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

func (s *KeeperTestSuite) persistRollApp(ras ...rollapp) {
	for _, ra := range ras {
		s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
			RollappId:    ra.rollAppId,
			Owner:        ra.owner,
			Bech32Prefix: ra.bech32,
		})

		if ra.alias != "" {
			if len(ra.aliases) == 0 {
				panic("must provide aliases if alias is set")
			} else if !slices.Contains(ra.aliases, ra.alias) {
				panic("alias must be in aliases")
			}
		}

		for _, alias := range ra.aliases {
			err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, ra.rollAppId, alias)
			s.Require().NoError(err)
		}
	}
}

// pureSetRollApp persists a rollapp without any side effects and all checking was skipped.
// Used to persist some invalid rollapp for testing.
func (s *KeeperTestSuite) pureSetRollApp(ra rollapptypes.Rollapp) {
	_store := prefix.NewStore(s.ctx.KVStore(s.rollappStoreKey), rollapptypes.KeyPrefix(rollapptypes.RollappKeyPrefix))
	b := s.cdc.MustMarshal(&ra)
	_store.Set(rollapptypes.RollappKey(
		ra.RollappId,
	), b)

	s.Require().True(s.dymNsKeeper.IsRollAppId(s.ctx, ra.RollappId))
}

func (s *KeeperTestSuite) moduleParams() dymnstypes.Params {
	return s.dymNsKeeper.GetParams(s.ctx)
}

func (s *KeeperTestSuite) updateModuleParams(f func(dymnstypes.Params) dymnstypes.Params) {
	moduleParams := s.moduleParams()
	moduleParams = f(moduleParams)
	err := s.dymNsKeeper.SetParams(s.ctx, moduleParams)
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

func (s *KeeperTestSuite) requireDymNameList(dymNames []dymnstypes.DymName, wantNames []string) {
	var gotNames []string
	for _, dymName := range dymNames {
		gotNames = append(gotNames, dymName.Name)
	}

	sort.Strings(gotNames)
	sort.Strings(wantNames)

	if len(wantNames) == 0 {
		wantNames = nil
	}

	s.Require().Equal(wantNames, gotNames)
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

type dymNameBuilder struct {
	name       string
	owner      string
	controller string
	expireAt   int64
	configs    []dymnstypes.DymNameConfig
}

func newDN(name, owner string) *dymNameBuilder {
	return &dymNameBuilder{
		name:       name,
		owner:      owner,
		controller: owner,
		expireAt:   time.Now().Unix() + 10,
		configs:    nil,
	}
}

func (m *dymNameBuilder) exp(now time.Time, offset int64) *dymNameBuilder {
	m.expireAt = now.Unix() + offset
	return m
}

func (m *dymNameBuilder) cfgN(chainId, subName, resolveTo string) *dymNameBuilder {
	m.configs = append(m.configs, dymnstypes.DymNameConfig{
		Type:    dymnstypes.DymNameConfigType_DCT_NAME,
		ChainId: chainId,
		Path:    subName,
		Value:   resolveTo,
	})
	return m
}

func (m *dymNameBuilder) build() dymnstypes.DymName {
	return dymnstypes.DymName{
		Name:       m.name,
		Owner:      m.owner,
		Controller: m.controller,
		ExpireAt:   m.expireAt,
		Configs:    m.configs,
	}
}

func (m *dymNameBuilder) buildSlice() []dymnstypes.DymName {
	return []dymnstypes.DymName{m.build()}
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

func (m reqRollApp) HasOnlyAlias(alias string) {
	list := m.s.dymNsKeeper.GetAliasesOfRollAppId(m.s.ctx, m.rollAppId)
	m.s.Require().Len(list, 1)
	m.s.Require().Equal(alias, list[0])
}

func (m reqRollApp) HasNoAlias() {
	alias, found := m.s.dymNsKeeper.GetAliasByRollAppId(m.s.ctx, m.rollAppId)
	m.s.Require().Falsef(found, "got: %v", m.s.dymNsKeeper.GetAliasesOfRollAppId(m.s.ctx, m.rollAppId))
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
	m.s.Require().False(found, "got: %s", gotRollAppId)
	m.s.Require().Empty(gotRollAppId)
}

func (m reqAlias) LinkedToRollApp(rollAppId string) {
	gotRollAppId, found := m.s.dymNsKeeper.GetRollAppIdByAlias(m.s.ctx, m.alias)
	m.s.Require().True(found)
	m.s.Require().Equal(rollAppId, gotRollAppId)
}

func (m reqAlias) mustHaveActiveSO() reqAlias {
	so := m.s.dymNsKeeper.GetSellOrder(m.s.ctx, m.alias, dymnstypes.TypeAlias)
	m.s.Require().NotNil(so)
	return m
}

func (m reqAlias) noActiveSO() reqAlias {
	so := m.s.dymNsKeeper.GetSellOrder(m.s.ctx, m.alias, dymnstypes.TypeAlias)
	m.s.Require().Nil(so)
	return m
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

//

type reqConfiguredAddr struct {
	s       *KeeperTestSuite
	cfgAddr string
}

func (s *KeeperTestSuite) requireConfiguredAddress(cfgAddr string) *reqConfiguredAddr {
	return &reqConfiguredAddr{
		s:       s,
		cfgAddr: cfgAddr,
	}
}

func (m reqConfiguredAddr) mappedDymNames(names ...string) {
	if len(names) == 0 {
		panic("must provide at least one name")
	}

	dymNames, err := m.s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(m.s.ctx, m.cfgAddr)
	m.s.Require().NoError(err)
	m.s.requireDymNameList(dymNames, names)
}

func (m reqConfiguredAddr) notMappedToAnyDymName() {
	dymNames, err := m.s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(m.s.ctx, m.cfgAddr)
	m.s.Require().NoError(err)
	m.s.Require().Empty(dymNames)
}

//

type reqFallbackAddr struct {
	s      *KeeperTestSuite
	fbAddr dymnstypes.FallbackAddress
}

func (s *KeeperTestSuite) requireFallbackAddress(fbAddr dymnstypes.FallbackAddress) *reqFallbackAddr {
	return &reqFallbackAddr{
		s:      s,
		fbAddr: fbAddr,
	}
}

func (m reqFallbackAddr) mappedDymNames(names ...string) {
	if len(names) == 0 {
		panic("must provide at least one name")
	}

	dymNames, err := m.s.dymNsKeeper.GetDymNamesContainsFallbackAddress(m.s.ctx, m.fbAddr)
	m.s.Require().NoError(err)
	m.s.requireDymNameList(dymNames, names)
}

func (m reqFallbackAddr) notMappedToAnyDymName() {
	dymNames, err := m.s.dymNsKeeper.GetDymNamesContainsFallbackAddress(m.s.ctx, m.fbAddr)
	m.s.Require().NoError(err, "got: %v", dymNames)
	m.s.Require().Empty(dymNames)
}

//

type reqDymNameS struct {
	s       *KeeperTestSuite
	dymName string
}

func (s *KeeperTestSuite) requireDymName(dymName string) *reqDymNameS {
	return &reqDymNameS{
		s:       s,
		dymName: dymName,
	}
}

func (m reqDymNameS) mustHaveActiveSO() reqDymNameS {
	so := m.s.dymNsKeeper.GetSellOrder(m.s.ctx, m.dymName, dymnstypes.TypeName)
	m.s.Require().NotNil(so)
	return m
}

func (m reqDymNameS) noActiveSO() reqDymNameS {
	so := m.s.dymNsKeeper.GetSellOrder(m.s.ctx, m.dymName, dymnstypes.TypeName)
	m.s.Require().Nil(so)
	return m
}

func (m reqDymNameS) mustHaveHistoricalSoCount(count int) reqDymNameS {
	hso := m.s.dymNsKeeper.GetHistoricalSellOrders(m.s.ctx, m.dymName, dymnstypes.TypeName)
	m.s.Require().Len(hso, count)
	return m
}

func (m reqDymNameS) noHistoricalSO() reqDymNameS {
	hso := m.s.dymNsKeeper.GetHistoricalSellOrders(m.s.ctx, m.dymName, dymnstypes.TypeName)
	m.s.Require().Empty(hso)
	return m
}
