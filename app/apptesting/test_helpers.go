package apptesting

import (
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/math"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	usim "github.com/cosmos/cosmos-sdk/testutil/sims"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	cometbfttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/stretchr/testify/require"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	app "github.com/dymensionxyz/dymension/v3/app"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

var TestChainID = "dymension_100-1"

var DefaultConsensusParams = func() *cometbftproto.ConsensusParams {
	ret := usim.DefaultConsensusParams
	ret.Block.MaxGas = -1
	return ret
}()

// Passed into `Simapp` constructor.
type SetupOptions struct {
	Logger             log.Logger
	DB                 *dbm.MemDB
	InvCheckPeriod     uint
	HomePath           string
	SkipUpgradeHeights map[int64]bool
	EncConfig          params.EncodingConfig
	AppOpts            types.AppOptions
}

// Having this enabled led to some problems because some tests use intrusive methods to modify the state, whichs breaks invariants
var InvariantCheckInterval = uint(0) // disabled

func SetupTestingApp() (*app.App, app.GenesisState) {
	newApp := app.New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, usim.EmptyAppOptions{}, bam.SetChainID(TestChainID))
	encCdc := newApp.AppCodec()
	defaultGenesisState := newApp.DefaultGenesis()

	incentivesGenesisStateJson := defaultGenesisState[incentivestypes.ModuleName]
	var incentivesGenesisState incentivestypes.GenesisState
	encCdc.MustUnmarshalJSON(incentivesGenesisStateJson, &incentivesGenesisState)
	incentivesGenesisState.LockableDurations = append(incentivesGenesisState.LockableDurations, time.Second*60)
	defaultGenesisState[incentivestypes.ModuleName] = encCdc.MustMarshalJSON(&incentivesGenesisState)

	// force disable EnableCreate of x/evm
	evmGenesisStateJson := defaultGenesisState[evmtypes.ModuleName]
	var evmGenesisState evmtypes.GenesisState
	encCdc.MustUnmarshalJSON(evmGenesisStateJson, &evmGenesisState)
	evmGenesisState.Params.EnableCreate = false
	defaultGenesisState[evmtypes.ModuleName] = encCdc.MustMarshalJSON(&evmGenesisState)

	return newApp, defaultGenesisState
}

// Setup initializes a new SimApp with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit in the default token of the simapp from first genesis
// account. A Nop logger is set in SimApp.
func Setup(t *testing.T) *app.App {
	t.Helper()

	app, genesisState := SetupTestingApp()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := cometbfttypes.NewValidator(pubKey, 1)
	valSet := cometbfttypes.NewValidatorSet([]*cometbfttypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balances := []banktypes.Balance{
		{
			Address: acc.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000000000000000000))),
		},
	}

	genesisState = genesisStateWithValSet(t, app, genesisState, valSet, []authtypes.GenesisAccount{acc}, balances...)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	_, err = app.InitChain(
		&abci.RequestInitChain{
			ChainId:         TestChainID,
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)
	require.NoError(t, err)

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Hash:               app.LastCommitID().Hash,
		NextValidatorsHash: valSet.Hash(),
	})
	require.NoError(t, err)

	return app
}

func genesisStateWithValSet(t *testing.T,
	app *app.App, genesisState app.GenesisState,
	valSet *cometbfttypes.ValidatorSet, genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) app.GenesisState {
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		require.NoError(t, err)
		pkAny, err := codectypes.NewAnyWithValue(pk)
		require.NoError(t, err)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   math.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
			MinSelfDelegation: math.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress().String(), sdk.ValAddress(val.Address).String(), math.LegacyOneDec()))

	}
	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	return genesisState
}

type GenerateAccountStrategy func(int) []sdk.AccAddress

// CreateRandomAccounts is a strategy used by addTestAddrs() in order to generated addresses in random order.
func CreateRandomAccounts(accNum int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, accNum)
	for i := 0; i < accNum; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}

// AddTestAddrs constructs and returns accNum amount of accounts with an
// initial balance of accAmt in random order
func AddTestAddrs(app *app.App, ctx sdk.Context, accNum int, accAmt math.Int) []sdk.AccAddress {
	return addTestAddrs(app, ctx, accNum, accAmt, CreateRandomAccounts)
}

func addTestAddrs(app *app.App, ctx sdk.Context, accNum int, accAmt math.Int, strategy GenerateAccountStrategy) []sdk.AccAddress {
	testAddrs := strategy(accNum)

	denom, _ := app.StakingKeeper.BondDenom(ctx)
	initCoins := sdk.NewCoins(sdk.NewCoin(denom, accAmt))

	for _, addr := range testAddrs {
		FundAccount(app, ctx, addr, initCoins)
	}

	return testAddrs
}

func FundAccount(app *app.App, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	if err != nil {
		panic(err)
	}

	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins)
	if err != nil {
		panic(err)
	}
}

func FundForAliasRegistration(app *app.App, ctx sdk.Context, alias, creator string) {
	if alias == "" {
		return
	}

	dymNsParams := dymnstypes.DefaultPriceParams()
	aliasRegistrationCost := sdk.NewCoins(sdk.NewCoin(
		params.BaseDenom, dymNsParams.GetAliasPrice(alias),
	))
	FundAccount(
		app, ctx, sdk.MustAccAddressFromBech32(creator), aliasRegistrationCost,
	)
}
