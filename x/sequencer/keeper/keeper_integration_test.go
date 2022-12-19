package keeper_test

import (
	"fmt"
	"testing"

	dymensionapp "github.com/dymensionxyz/dymension/app"
	sharedtypes "github.com/dymensionxyz/dymension/shared/types"
	"github.com/dymensionxyz/dymension/testutil/sample"
	"github.com/dymensionxyz/dymension/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/x/sequencer/types"
	sequencertypes "github.com/dymensionxyz/dymension/x/sequencer/types"

	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"

	ce "github.com/tendermint/tendermint/crypto/encoding"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	transferEventCount            = 3 // As emitted by the bank
	createEventCount              = 8
	playEventCountFirst           = 8 // Extra "sender" attribute emitted by the bank
	playEventCountNext            = 7
	rejectEventCount              = 4
	rejectEventCountWithTransfer  = 5 // Extra "sender" attribute emitted by the bank
	forfeitEventCount             = 4
	forfeitEventCountWithTransfer = 5 // Extra "sender" attribute emitted by the bank
	alice                         = "cosmos1jmjfq0tplp9tmx4v9uemw72y4d2wa5nr3xn9d3"
	bob                           = "cosmos1xyxs3skf3f4jfqeuv89yyaqvjc6lffavxqhc8g"
	carol                         = "cosmos1e0w5t53nrq7p66fye6c8p0ynyhf6y24l4yuxd7"
	balAlice                      = 50000000
	balBob                        = 20000000
	balCarol                      = 10000000
	foreignToken                  = "foreignToken"
	balTokenAlice                 = 5
	balTokenBob                   = 2
	balTokenCarol                 = 1
)

var (
	sequencerModuleAddress string
)

type IntegrationTestSuite struct {
	suite.Suite

	app         *dymensionapp.App
	msgServer   types.MsgServer
	ctx         sdk.Context
	queryClient types.QueryClient
}

func TestSequencerKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupTest() {
	app := dymensionapp.Setup(false)
	ctx := app.GetBaseApp().NewContext(false, tmproto.Header{})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	app.RollappKeeper.SetParams(ctx, rollapptypes.DefaultParams())
	sequencerModuleAddress = app.AccountKeeper.GetModuleAddress(types.ModuleName).String()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.SequencerKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.msgServer = keeper.NewMsgServerImpl(app.SequencerKeeper)
	suite.ctx = ctx
	suite.queryClient = queryClient
}

func (suite *IntegrationTestSuite) TestCreateSequencer() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// max sequencers per rollapp
	maxSequencers := 10

	// sequencersExpect is the expected result of query all
	sequencersExpect := []*types.Sequencer{}

	// rollappSequencersExpect is a map from rollappId to a map of sequencer addresses list
	type rollappSequencersExpectKey struct {
		rollappId, sequencerAddress string
	}
	rollappSequencersExpect := make(map[rollappSequencersExpectKey]string)

	// for 3 rollapps, test 10 sequencers creations
	for j := 0; j < 3; j++ {
		rollapp := rollapptypes.Rollapp{
			RollappId:             fmt.Sprintf("%s%d", "rollapp", j),
			Creator:               alice,
			Version:               0,
			CodeStamp:             "",
			GenesisPath:           "",
			MaxWithholdingBlocks:  1,
			MaxSequencers:         uint64(maxSequencers),
			PermissionedAddresses: sharedtypes.Sequencers{Addresses: []string{}},
		}
		suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

		rollappId := rollapp.GetRollappId()

		for i := 0; i < 10; i++ {
			pk := tmtypes.NewMockPV().PrivKey.PubKey()
			pubkey, err := ce.PubKeyToProto(pk)
			addr := sdk.AccAddress(pk.Address())
			suite.Require().Nil(err)

			// sequencer is the sequencer to create
			sequencer := types.MsgCreateSequencer{
				Creator:          alice,
				SequencerAddress: addr.String(),
				DymintPubKey:     pubkey,
				RollappId:        rollappId,
				Description:      sequencertypes.Description{},
			}
			// sequencerExpect is the expected result of creating a sequencer
			sequencerExpect := types.Sequencer{
				SequencerAddress: sequencer.GetSequencerAddress(),
				Creator:          sequencer.GetCreator(),
				DymintPubKey:     sequencer.GetDymintPubKey(),
				RollappId:        rollappId,
				Description:      sequencer.GetDescription(),
			}
			// create sequencer
			createResponse, err := suite.msgServer.CreateSequencer(goCtx, &sequencer)
			suite.Require().Nil(err)
			suite.Require().EqualValues(types.MsgCreateSequencerResponse{}, *createResponse)

			// query the spesific sequencer
			queryResponse, err := suite.queryClient.Sequencer(goCtx, &types.QueryGetSequencerRequest{
				SequencerAddress: sequencer.GetSequencerAddress(),
			})
			suite.Require().Nil(err)
			suite.Require().EqualValues(sequencerExpect, queryResponse.SequencerInfo.Sequencer)

			// add the sequencer to the list of get all expected list
			sequencersExpect = append(sequencersExpect, &sequencerExpect)

			sequencersRes, totalRes := getAll(suite)
			suite.Require().EqualValues(len(sequencersExpect), totalRes)
			// verify that query all contains all the sequencers that were created
			vereifyAll(suite, sequencersExpect, sequencersRes)

			// add the sequencer to the list of spesific rollapp
			rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencerExpect.SequencerAddress}] =
				sequencerExpect.SequencerAddress
		}
	}

	totalFound := 0
	// check query by rollapp
	for j := 0; j < 3; j++ {
		rollappId := fmt.Sprintf("%s%d", "rollapp", j)
		queryAllResponse, err := suite.queryClient.SequencersByRollapp(goCtx,
			&types.QueryGetSequencersByRollappRequest{RollappId: rollappId})
		suite.Require().Nil(err)
		// verify that all the addresses of the rollapp are found
		for _, sequencerInfo := range queryAllResponse.SequencerInfoList {
			sequencerResAddresses := sequencerInfo.Sequencer.SequencerAddress
			suite.Require().EqualValues(rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencerResAddresses}],
				sequencerResAddresses)
		}
		totalFound += len(queryAllResponse.SequencerInfoList)
	}
	suite.Require().EqualValues(totalFound, len(rollappSequencersExpect))
}

func (suite *IntegrationTestSuite) TestCreateSequencerAlreadyExists() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	rollapp := rollapptypes.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               0,
		CodeStamp:             "",
		GenesisPath:           "",
		MaxWithholdingBlocks:  1,
		MaxSequencers:         1,
		PermissionedAddresses: sharedtypes.Sequencers{Addresses: []string{}},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	rollappId := rollapp.GetRollappId()

	pk := tmtypes.NewMockPV().PrivKey.PubKey()
	pubkey, err := ce.PubKeyToProto(pk)
	addr := sdk.AccAddress(pk.Address())
	suite.Require().Nil(err)
	sequencer := types.MsgCreateSequencer{
		Creator:          alice,
		SequencerAddress: addr.String(),
		DymintPubKey:     pubkey,
		RollappId:        rollappId,
		Description:      sequencertypes.Description{},
	}
	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencer)
	suite.Require().Nil(err)

	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencer)
	suite.EqualError(err, types.ErrSequencerExists.Error())
}

func (suite *IntegrationTestSuite) TestCreateSequencerUnknownRollappId() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	pk := tmtypes.NewMockPV().PrivKey.PubKey()
	pubkey, err := ce.PubKeyToProto(pk)
	addr := sdk.AccAddress(pk.Address())
	suite.Require().Nil(err)
	sequencer := types.MsgCreateSequencer{
		Creator:          alice,
		SequencerAddress: addr.String(),
		DymintPubKey:     pubkey,
		RollappId:        "rollappId",
		Description:      sequencertypes.Description{},
	}

	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencer)
	suite.EqualError(err, types.ErrUnknownRollappId.Error())
}

func (suite *IntegrationTestSuite) TestCreatePermissionedSequencer() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	pk := tmtypes.NewMockPV().PrivKey.PubKey()
	pubkey, err := ce.PubKeyToProto(pk)
	addr := sdk.AccAddress(pk.Address())
	suite.Require().Nil(err)

	sequencerAddress := addr.String()

	rollapp := rollapptypes.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               0,
		CodeStamp:             "",
		GenesisPath:           "",
		MaxWithholdingBlocks:  1,
		MaxSequencers:         1,
		PermissionedAddresses: sharedtypes.Sequencers{Addresses: []string{sequencerAddress}},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	rollappId := rollapp.GetRollappId()

	suite.Require().Nil(err)
	sequencer := types.MsgCreateSequencer{
		Creator:          alice,
		SequencerAddress: sequencerAddress,
		DymintPubKey:     pubkey,
		RollappId:        rollappId,
		Description:      sequencertypes.Description{},
	}

	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencer)
	suite.Require().Nil(err)

	// query the spesific sequencer
	queryResponse, err := suite.queryClient.Sequencer(goCtx, &types.QueryGetSequencerRequest{
		SequencerAddress: sequencer.GetSequencerAddress(),
	})
	suite.Require().Nil(err)

	// sequencerExpect is the expected result of creating a sequencer
	sequencerExpect := types.Sequencer{
		SequencerAddress: sequencer.GetSequencerAddress(),
		Creator:          sequencer.GetCreator(),
		DymintPubKey:     sequencer.GetDymintPubKey(),
		RollappId:        rollappId,
		Description:      sequencer.GetDescription(),
	}
	suite.Require().EqualValues(sequencerExpect, queryResponse.SequencerInfo.Sequencer)
}

func (suite *IntegrationTestSuite) TestCreateSequencerNotPermissioned() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	rollapp := rollapptypes.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               0,
		CodeStamp:             "",
		GenesisPath:           "",
		MaxWithholdingBlocks:  1,
		MaxSequencers:         1,
		PermissionedAddresses: sharedtypes.Sequencers{Addresses: []string{sample.AccAddress()}},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	rollappId := rollapp.GetRollappId()

	pk := tmtypes.NewMockPV().PrivKey.PubKey()
	pubkey, err := ce.PubKeyToProto(pk)
	addr := sdk.AccAddress(pk.Address())
	suite.Require().Nil(err)

	sequencer := types.MsgCreateSequencer{
		Creator:          alice,
		SequencerAddress: addr.String(),
		DymintPubKey:     pubkey,
		RollappId:        rollappId,
		Description:      sequencertypes.Description{},
	}

	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencer)
	suite.EqualError(err, types.ErrSequencerNotPermissioned.Error())
}

func (suite *IntegrationTestSuite) TestMaxSequencersLimit() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)
	maxSequencers := 3

	rollapp := rollapptypes.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               0,
		CodeStamp:             "",
		GenesisPath:           "",
		MaxWithholdingBlocks:  1,
		MaxSequencers:         uint64(maxSequencers),
		PermissionedAddresses: sharedtypes.Sequencers{Addresses: []string{}},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	rollappId := rollapp.GetRollappId()

	// create MaxSequencers
	for i := 0; i < maxSequencers; i++ {
		pk := tmtypes.NewMockPV().PrivKey.PubKey()
		pubkey, err := ce.PubKeyToProto(pk)
		addr := sdk.AccAddress(pk.Address())
		suite.Require().Nil(err)

		sequencer := types.MsgCreateSequencer{
			Creator:          alice,
			SequencerAddress: addr.String(),
			DymintPubKey:     pubkey,
			RollappId:        rollappId,
			Description:      sequencertypes.Description{},
		}
		_, err = suite.msgServer.CreateSequencer(goCtx, &sequencer)
		suite.Require().Nil(err)
	}

	// add more to be failed
	for i := 0; i < 2; i++ {
		pk := tmtypes.NewMockPV().PrivKey.PubKey()
		pubkey, err := ce.PubKeyToProto(pk)
		addr := sdk.AccAddress(pk.Address())
		suite.Require().Nil(err)

		sequencer := types.MsgCreateSequencer{
			Creator:          alice,
			SequencerAddress: addr.String(),
			DymintPubKey:     pubkey,
			RollappId:        rollappId,
			Description:      sequencertypes.Description{},
		}
		_, err = suite.msgServer.CreateSequencer(goCtx, &sequencer)
		suite.EqualError(err, types.ErrMaxSequencersLimit.Error())
	}
}

func (suite *IntegrationTestSuite) TestUpdateStateSecondSeqErrNotActiveSequencer() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	rollapp := rollapptypes.Rollapp{
		RollappId:            "rollapp1",
		Creator:              alice,
		Version:              0,
		CodeStamp:            "",
		GenesisPath:          "",
		MaxWithholdingBlocks: 1,
		MaxSequencers:        2,
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	rollappId := rollapp.GetRollappId()

	// create first sequencer
	pk1 := tmtypes.NewMockPV().PrivKey.PubKey()
	pubkey1, err := ce.PubKeyToProto(pk1)
	addr1 := sdk.AccAddress(pk1.Address())
	suite.Require().Nil(err)

	suite.Require().Nil(err)
	sequencer1 := types.MsgCreateSequencer{
		Creator:          alice,
		SequencerAddress: addr1.String(),
		DymintPubKey:     pubkey1,
		RollappId:        rollappId,
		Description:      sequencertypes.Description{},
	}
	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencer1)
	suite.Require().Nil(err)

	// create second sequencer
	pk2 := tmtypes.NewMockPV().PrivKey.PubKey()
	pubkey2, err := ce.PubKeyToProto(pk2)
	addr2 := sdk.AccAddress(pk2.Address())
	suite.Require().Nil(err)
	sequencer2 := types.MsgCreateSequencer{
		Creator:          alice,
		SequencerAddress: addr2.String(),
		DymintPubKey:     pubkey2,
		RollappId:        rollappId,
		Description:      sequencertypes.Description{},
	}
	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencer2)
	suite.Require().Nil(err)

	// check scheduler operating status
	scheduler, found := suite.app.SequencerKeeper.GetScheduler(suite.ctx, sequencer1.SequencerAddress)
	suite.Require().True(found)
	suite.EqualValues(scheduler.Status, types.Proposer)

	// check scheduler operating status
	scheduler, found = suite.app.SequencerKeeper.GetScheduler(suite.ctx, sequencer2.SequencerAddress)
	suite.Require().True(found)
	suite.EqualValues(scheduler.Status, types.Inactive)
}

//---------------------------------------
// vereifyAll receives a list of expected results and a map of sequencerAddress->sequencer
// the function verifies that the map contains all the sequencers that are in the list and only them
func vereifyAll(suite *IntegrationTestSuite, sequencersExpect []*types.Sequencer, sequencersRes map[string]*types.Sequencer) {
	// check number of items are equal
	suite.Require().EqualValues(len(sequencersExpect), len(sequencersRes))
	for i := 0; i < len(sequencersExpect); i++ {
		sequencerExpect := sequencersExpect[i]
		sequencerRes := sequencersRes[sequencerExpect.GetSequencerAddress()]
		suite.Require().EqualValues(sequencerExpect, sequencerRes)
	}
}

// getAll quires for all exsisting sequencers and returns a map of sequencerId->sequencer
func getAll(suite *IntegrationTestSuite) (map[string]*types.Sequencer, int) {
	goCtx := sdk.WrapSDKContext(suite.ctx)
	totalChecked := 0
	totalRes := 0
	nextKey := []byte{}
	sequencersRes := make(map[string]*types.Sequencer)
	for {
		queryAllResponse, err := suite.queryClient.SequencerAll(goCtx,
			&types.QueryAllSequencerRequest{
				Pagination: &query.PageRequest{
					Key:        nextKey,
					Offset:     0,
					Limit:      0,
					CountTotal: true,
					Reverse:    false,
				}})
		suite.Require().Nil(err)

		if totalRes == 0 {
			totalRes = int(queryAllResponse.GetPagination().GetTotal())
		}

		for i := 0; i < len(queryAllResponse.SequencerInfoList); i++ {
			sequencerRes := queryAllResponse.SequencerInfoList[i].Sequencer
			sequencersRes[sequencerRes.GetSequencerAddress()] = &sequencerRes
		}
		totalChecked += len(queryAllResponse.SequencerInfoList)
		nextKey = queryAllResponse.GetPagination().GetNextKey()

		if nextKey == nil {
			break
		}
	}

	return sequencersRes, totalRes
}
