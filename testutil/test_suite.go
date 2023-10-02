package testutil

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	simapp "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"golang.org/x/exp/slices"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/dymensionxyz/dymension/app"

	"github.com/dymensionxyz/dymension/x/gamm/pool-models/balancer"
	gammtypes "github.com/dymensionxyz/dymension/x/gamm/types"
	poolmanagertypes "github.com/dymensionxyz/dymension/x/poolmanager/types"
)

type KeeperTestHelper struct {
	suite.Suite

	App         *app.App
	Ctx         sdk.Context
	QueryHelper *baseapp.QueryServiceTestHelper
	TestAccs    []sdk.AccAddress
}

var (
	SecondaryDenom  = "uion"
	SecondaryAmount = sdk.NewInt(100000000)
)

// Setup sets up basic environment for suite (App, Ctx, and test accounts)
func (s *KeeperTestHelper) Setup() {
	s.App = app.Setup(s.T(), false)
	s.Ctx = s.App.BaseApp.NewContext(false, tmtypes.Header{Height: 1, ChainID: "dymension_100-1", Time: time.Now().UTC()})
	s.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: s.App.GRPCQueryRouter(),
		Ctx:             s.Ctx,
	}

	s.TestAccs = CreateRandomAccounts(3)
}

func (s *KeeperTestHelper) SetupTestForInitGenesis() {
	// Setting to True, leads to init genesis not running
	s.App = app.Setup(s.T(), true)
	s.Ctx = s.App.BaseApp.NewContext(true, tmtypes.Header{})
}

// CreateTestContext creates a test context.
func (s *KeeperTestHelper) CreateTestContext() sdk.Context {
	ctx, _ := s.CreateTestContextWithMultiStore()
	return ctx
}

// CreateTestContextWithMultiStore creates a test context and returns it together with multi store.
func (s *KeeperTestHelper) CreateTestContextWithMultiStore() (sdk.Context, sdk.CommitMultiStore) {
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()

	ms := rootmulti.NewStore(db, logger)

	return sdk.NewContext(ms, tmtypes.Header{}, false, logger), ms
}

// CreateTestContext creates a test context.
func (s *KeeperTestHelper) Commit() {
	oldHeight := s.Ctx.BlockHeight()
	oldHeader := s.Ctx.BlockHeader()
	s.App.Commit()
	newHeader := tmtypes.Header{Height: oldHeight + 1, ChainID: oldHeader.ChainID, Time: oldHeader.Time.Add(time.Second)}
	s.App.BeginBlock(abci.RequestBeginBlock{Header: newHeader})
	s.Ctx = s.App.GetBaseApp().NewContext(false, newHeader)
}

// FundAcc funds target address with specified amount.
func (s *KeeperTestHelper) FundAcc(acc sdk.AccAddress, amounts sdk.Coins) {
	err := simapp.FundAccount(s.App.BankKeeper, s.Ctx, acc, amounts)
	s.Require().NoError(err)
}

// FundModuleAcc funds target modules with specified amount.
func (s *KeeperTestHelper) FundModuleAcc(moduleName string, amounts sdk.Coins) {
	err := simapp.FundModuleAccount(s.App.BankKeeper, s.Ctx, moduleName, amounts)
	s.Require().NoError(err)
}

func (s *KeeperTestHelper) MintCoins(coins sdk.Coins) {
	err := s.App.BankKeeper.MintCoins(s.Ctx, gammtypes.ModuleName, coins)
	s.Require().NoError(err)
}

// BeginNewBlock starts a new block.
func (s *KeeperTestHelper) BeginNewBlock(executeNextEpoch bool) {
	var valAddr []byte

	validators := s.App.StakingKeeper.GetAllValidators(s.Ctx)
	if len(validators) >= 1 {
		valAddrFancy, err := validators[0].GetConsAddr()
		s.Require().NoError(err)
		valAddr = valAddrFancy.Bytes()
	}
	// else {
	// 	valAddrFancy := s.SetupValidator(stakingtypes.Bonded)
	// 	validator, _ := s.App.StakingKeeper.GetValidator(s.Ctx, valAddrFancy)
	// 	valAddr2, _ := validator.GetConsAddr()
	// 	valAddr = valAddr2.Bytes()
	// }

	s.BeginNewBlockWithProposer(executeNextEpoch, valAddr)
}

// BeginNewBlockWithProposer begins a new block with a proposer.
func (s *KeeperTestHelper) BeginNewBlockWithProposer(executeNextEpoch bool, proposer sdk.ValAddress) {
	validator, found := s.App.StakingKeeper.GetValidator(s.Ctx, proposer)
	s.Assert().True(found)

	valConsAddr, err := validator.GetConsAddr()
	s.Require().NoError(err)

	valAddr := valConsAddr.Bytes()

	newBlockTime := s.Ctx.BlockTime().Add(5 * time.Second)

	header := tmtypes.Header{Height: s.Ctx.BlockHeight() + 1, Time: newBlockTime}
	newCtx := s.Ctx.WithBlockTime(newBlockTime).WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Ctx = newCtx
	lastCommitInfo := abci.LastCommitInfo{
		Votes: []abci.VoteInfo{{
			Validator:       abci.Validator{Address: valAddr, Power: 1000},
			SignedLastBlock: true,
		}},
	}
	reqBeginBlock := abci.RequestBeginBlock{Header: header, LastCommitInfo: lastCommitInfo}

	fmt.Println("beginning block ", s.Ctx.BlockHeight())
	s.App.BeginBlocker(s.Ctx, reqBeginBlock)
	s.Ctx = s.App.NewContext(false, reqBeginBlock.Header)
}

// EndBlock ends the block, and runs commit
func (s *KeeperTestHelper) EndBlock() {
	reqEndBlock := abci.RequestEndBlock{Height: s.Ctx.BlockHeight()}
	s.App.EndBlocker(s.Ctx, reqEndBlock)
}

func (s *KeeperTestHelper) RunMsg(msg sdk.Msg) (*sdk.Result, error) {
	// cursed that we have to copy this internal logic from SDK
	router := s.App.GetBaseApp().MsgServiceRouter()
	if handler := router.Handler(msg); handler != nil {
		// ADR 031 request type routing
		return handler(s.Ctx, msg)
	}
	s.FailNow("msg %v could not be ran", msg)
	return nil, fmt.Errorf("msg %v could not be ran", msg)
}

// AllocateRewardsToValidator allocates reward tokens to a distribution module then allocates rewards to the validator address.
func (s *KeeperTestHelper) AllocateRewardsToValidator(valAddr sdk.ValAddress, rewardAmt sdk.Int) {
	validator, found := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().True(found)

	// allocate reward tokens to distribution module
	coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, rewardAmt)}
	err := simapp.FundModuleAccount(s.App.BankKeeper, s.Ctx, distrtypes.ModuleName, coins)
	s.Require().NoError(err)

	// allocate rewards to validator
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	decTokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(20000)}}
	s.App.DistrKeeper.AllocateTokensToValidator(s.Ctx, validator, decTokens)
}

// SetupGammPoolsWithBondDenomMultiplier uses given multipliers to set initial pool supply of bond denom.
func (s *KeeperTestHelper) SetupGammPoolsWithBondDenomMultiplier(multipliers []sdk.Dec) []gammtypes.CFMMPoolI {
	bondDenom := s.App.StakingKeeper.BondDenom(s.Ctx)
	// TODO: use sdk crypto instead of tendermint to generate address
	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())

	params := s.App.GAMMKeeper.GetParams(s.Ctx)

	pools := []gammtypes.CFMMPoolI{}
	for index, multiplier := range multipliers {
		token := fmt.Sprintf("token%d", index)
		udymAmount := sdk.NewDecFromBigInt(gammtypes.InitPoolSharesSupply.BigInt()).Mul(multiplier).RoundInt()

		s.FundAcc(acc1, sdk.NewCoins(
			sdk.NewCoin(bondDenom, udymAmount.Mul(sdk.NewInt(10))),
			sdk.NewInt64Coin(token, 100000),
		).Add(params.PoolCreationFee...))

		var (
			defaultFutureGovernor = ""

			// pool assets
			defaultFooAsset = balancer.PoolAsset{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(bondDenom, udymAmount),
			}
			defaultBarAsset = balancer.PoolAsset{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(token, sdk.NewInt(10000)),
			}

			poolAssets = []balancer.PoolAsset{defaultFooAsset, defaultBarAsset}
		)

		poolParams := balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}
		msg := balancer.NewMsgCreateBalancerPool(acc1, poolParams, poolAssets, defaultFutureGovernor)

		poolId, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, msg)
		s.Require().NoError(err)

		pool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
		s.Require().NoError(err)

		pools = append(pools, pool)
	}

	return pools
}

// SwapAndSetSpotPrice runs a swap to set Spot price of a pool using arbitrary values
// returns spot price after the arbitrary swap.
func (s *KeeperTestHelper) SwapAndSetSpotPrice(poolId uint64, fromAsset sdk.Coin, toAsset sdk.Coin) sdk.Dec {
	// create a dummy account
	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())

	// fund dummy account with tokens to swap
	coins := sdk.Coins{sdk.NewInt64Coin(fromAsset.Denom, 100000000000000)}
	s.FundAcc(acc1, coins)

	route := []poolmanagertypes.SwapAmountOutRoute{
		{
			PoolId:       poolId,
			TokenInDenom: fromAsset.Denom,
		},
	}
	_, err := s.App.PoolManagerKeeper.RouteExactAmountOut(
		s.Ctx,
		acc1,
		route,
		fromAsset.Amount,
		sdk.NewCoin(toAsset.Denom,
			toAsset.Amount.Quo(sdk.NewInt(4))))
	s.Require().NoError(err)

	spotPrice, err := s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, fromAsset.Denom, toAsset.Denom)
	s.Require().NoError(err)

	return spotPrice
}

// StateNotAltered validates that app state is not altered. Fails if it is.
func (s *KeeperTestHelper) StateNotAltered() {
	oldState := s.App.ExportState(s.Ctx)
	s.App.Commit()
	newState := s.App.ExportState(s.Ctx)
	s.Require().Equal(oldState, newState)
}

// CreateRandomAccounts is a function return a list of randomly generated AccAddresses
func CreateRandomAccounts(numAccts int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}

func TestMessageAuthzSerialization(t *testing.T, msg sdk.Msg) {
	someDate := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)
	const (
		mockGranter string = "cosmos1abc"
		mockGrantee string = "cosmos1xyz"
	)

	var (
		mockMsgGrant  authz.MsgGrant
		mockMsgRevoke authz.MsgRevoke
		mockMsgExec   authz.MsgExec
	)

	// Authz: Grant Msg
	typeURL := sdk.MsgTypeURL(msg)
	someDate = someDate.Add(time.Hour)
	grant, err := authz.NewGrant(someDate, authz.NewGenericAuthorization(typeURL), nil)
	require.NoError(t, err)

	msgGrant := authz.MsgGrant{Granter: mockGranter, Grantee: mockGrantee, Grant: grant}
	msgGrantBytes := json.RawMessage(sdk.MustSortJSON(authzcodec.ModuleCdc.MustMarshalJSON(&msgGrant)))
	err = authzcodec.ModuleCdc.UnmarshalJSON(msgGrantBytes, &mockMsgGrant)
	require.NoError(t, err)

	// Authz: Revoke Msg
	msgRevoke := authz.MsgRevoke{Granter: mockGranter, Grantee: mockGrantee, MsgTypeUrl: typeURL}
	msgRevokeByte := json.RawMessage(sdk.MustSortJSON(authzcodec.ModuleCdc.MustMarshalJSON(&msgRevoke)))
	err = authzcodec.ModuleCdc.UnmarshalJSON(msgRevokeByte, &mockMsgRevoke)
	require.NoError(t, err)

	// Authz: Exec Msg
	msgAny, err := cdctypes.NewAnyWithValue(msg)
	require.NoError(t, err)
	msgExec := authz.MsgExec{Grantee: mockGrantee, Msgs: []*cdctypes.Any{msgAny}}
	execMsgByte := json.RawMessage(sdk.MustSortJSON(authzcodec.ModuleCdc.MustMarshalJSON(&msgExec)))
	err = authzcodec.ModuleCdc.UnmarshalJSON(execMsgByte, &mockMsgExec)
	require.NoError(t, err)
	require.Equal(t, msgExec.Msgs[0].Value, mockMsgExec.Msgs[0].Value)
}

func GenerateTestAddrs() (string, string) {
	pk1 := ed25519.GenPrivKey().PubKey()
	validAddr := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid").String()
	return validAddr, invalidAddr
}

// AssertEventEmitted asserts that ctx's event manager has emitted the given number of events
// of the given type.
func (s *KeeperTestHelper) AssertEventEmitted(ctx sdk.Context, eventTypeExpected string, numEventsExpected int) {
	allEvents := ctx.EventManager().Events()
	// filter out other events
	actualEvents := make([]sdk.Event, 0)
	for _, event := range allEvents {
		if event.Type == eventTypeExpected {
			actualEvents = append(actualEvents, event)
		}
	}
	s.Equal(numEventsExpected, len(actualEvents))
}

func (s *KeeperTestHelper) FindEvent(events []sdk.Event, name string) sdk.Event {
	index := slices.IndexFunc(events, func(e sdk.Event) bool { return e.Type == name })
	if index == -1 {
		return sdk.Event{}
	}
	return events[index]
}

func (s *KeeperTestHelper) ExtractAttributes(event sdk.Event) map[string]string {
	attrs := make(map[string]string)
	if event.Attributes == nil {
		return attrs
	}
	for _, a := range event.Attributes {
		attrs[string(a.Key)] = string(a.Value)
	}
	return attrs
}
