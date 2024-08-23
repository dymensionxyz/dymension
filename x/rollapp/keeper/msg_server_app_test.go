package keeper_test

import (
	"cmp"
	"slices"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/urand"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (suite *RollappTestSuite) TestAddApp() {
	suite.SetupTest()
	suite.createRollappWithApp()
}

func (suite *RollappTestSuite) createRollappWithApp() types.RollappSummary {
	suite.SetupTest()
	creator := alice
	res := suite.createRollappWithCreatorAndVerify(nil, creator, true)
	req := &types.MsgAddApp{
		Creator:     creator,
		Name:        "app1",
		RollappId:   res.RollappId,
		Description: "My first app",
		Image:       "http://example.com/image1",
		Url:         "http://example.com/app1",
		Order:       1,
	}
	_, err := suite.msgServer.AddApp(suite.Ctx, req)
	suite.Require().NoError(err)

	// query the specific rollapp
	goCtx := sdk.WrapSDKContext(suite.Ctx)
	queryResponse, err := suite.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{
		RollappId: req.GetRollappId(),
	})
	suite.Require().Nil(err)

	appExpect := types.App{
		Name:        req.GetName(),
		RollappId:   req.GetRollappId(),
		Description: req.GetDescription(),
		Image:       req.GetImage(),
		Url:         req.GetUrl(),
		Order:       req.GetOrder(),
	}

	app, ok := suite.App.RollappKeeper.GetApp(suite.Ctx, req.Name, res.RollappId)
	suite.Require().True(ok)
	suite.Require().EqualValues(&appExpect, &app)
	suite.Require().Len(queryResponse.Apps, 1)
	suite.Require().EqualValues(&appExpect, queryResponse.Apps[0])

	return res
}

func (suite *RollappTestSuite) Test_msgServer_AddApp() {
	rollappID := urand.RollappID()

	tests := []struct {
		name     string
		msgs     []*types.MsgAddApp
		malleate func()
		wantErr  error
	}{
		{
			name: "success: add 1 app",
			msgs: []*types.MsgAddApp{
				{
					Creator:     alice,
					Name:        "app1",
					RollappId:   rollappID,
					Description: "My first app",
					Image:       "http://example.com/image1",
					Url:         "http://example.com/app1",
					Order:       1,
				},
			},
		}, {
			name: "success: add 1 app no order",
			msgs: []*types.MsgAddApp{
				{
					Creator:     alice,
					Name:        "app1",
					RollappId:   rollappID,
					Description: "My first app",
					Image:       "http://example.com/image1",
					Url:         "http://example.com/app1",
				},
			},
		}, {
			name: "success: add multiple apps",
			msgs: []*types.MsgAddApp{
				{
					Creator:     alice,
					Name:        "app1",
					RollappId:   rollappID,
					Description: "My first app",
					Image:       "http://example.com/image1",
					Url:         "http://example.com/app1",
					Order:       3,
				}, {
					Creator:     alice,
					Name:        "app2",
					RollappId:   rollappID,
					Description: "My second app",
					Image:       "http://example.com/image2",
					Url:         "http://example.com/app2",
					Order:       1,
				}, {
					Creator:     alice,
					Name:        "app3",
					RollappId:   rollappID,
					Description: "My third app",
					Image:       "http://example.com/image3",
					Url:         "http://example.com/app3",
					Order:       4,
				},
			},
		}, {
			name: "fail: add app with different creator",
			msgs: []*types.MsgAddApp{
				{
					Creator:     bob,
					Name:        "app1",
					RollappId:   rollappID,
					Description: "My first app",
					Image:       "http://example.com/image1",
					Url:         "http://example.com/app1",
					Order:       1,
				},
			},
			wantErr: types.ErrUnauthorizedSigner,
		}, {
			name: "fail: add app with different rollapp",
			msgs: []*types.MsgAddApp{
				{
					Creator:     alice,
					Name:        "app1",
					RollappId:   urand.RollappID(),
					Description: "My first app",
					Image:       "http://example.com/image1",
					Url:         "http://example.com/app1",
					Order:       1,
				},
			},
			wantErr: types.ErrNotFound,
		}, {
			name: "fail: add app with same name and rollappID",
			msgs: []*types.MsgAddApp{
				{
					Creator:     alice,
					Name:        "app1",
					RollappId:   rollappID,
					Description: "My first app",
					Image:       "http://example.com/image1",
					Url:         "http://example.com/app1",
					Order:       1,
				},
			},
			malleate: func() {
				suite.App.RollappKeeper.SetApp(suite.Ctx, types.App{
					Name:      "app1",
					RollappId: rollappID,
				})
			},
			wantErr: types.ErrAppExists,
		}, {
			name: "success: add app with same name and different rollappID",
			msgs: []*types.MsgAddApp{
				{
					Creator:     alice,
					Name:        "app1",
					RollappId:   rollappID,
					Description: "My first app",
					Image:       "http://example.com/image1",
					Url:         "http://example.com/app1",
					Order:       1,
				},
			},
			malleate: func() {
				rollapp2 := urand.RollappID()
				suite.createRollappWithIDAndCreator(rollapp2, alice)
				suite.App.RollappKeeper.SetApp(suite.Ctx, types.App{
					Name:      "app1",
					RollappId: rollapp2,
				})
			},
		}, {
			name: "fail: add 1 app - not enough funds",
			msgs: []*types.MsgAddApp{
				{
					Creator:     alice,
					Name:        "app1",
					RollappId:   rollappID,
					Description: "My first app",
					Image:       "http://example.com/image1",
					Url:         "http://example.com/app1",
					Order:       1,
				},
			},
			malleate: func() {
				params := suite.App.RollappKeeper.GetParams(suite.Ctx)
				params.AppCreationCost = sdk.NewInt64Coin("arax", 1)
				suite.App.RollappKeeper.SetParams(suite.Ctx, params)
			},
			wantErr: types.ErrAppCreationCostPayment,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			suite.SetupTest()
			suite.createRollappWithIDAndCreator(rollappID, alice)

			goCtx := sdk.WrapSDKContext(suite.Ctx)

			if tt.malleate != nil {
				tt.malleate()
			}

			for _, msg := range tt.msgs {
				_, err := suite.msgServer.AddApp(goCtx, msg)
				suite.Require().ErrorIs(err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			rollapp, err := suite.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{
				RollappId: rollappID,
			})
			suite.Require().NoError(err)
			suite.Require().Len(rollapp.Apps, len(tt.msgs))

			slices.SortFunc(tt.msgs, func(a, b *types.MsgAddApp) int {
				return cmp.Compare(a.Order, b.Order)
			})

			for i, app := range rollapp.Apps {
				suite.Require().Equal(tt.msgs[i].Order, app.Order)
			}
		})
	}
}

func (suite *RollappTestSuite) createRollappWithIDAndCreator(rollappId string, creator string) {
	rollapp := types.MsgCreateRollapp{
		Creator:          creator,
		RollappId:        rollappId,
		InitialSequencer: sample.AccAddress(),
		Bech32Prefix:     "rol",
		GenesisChecksum:  "checksum",
		VmType:           types.Rollapp_EVM,
		Metadata:         &mockRollappMetadata,
	}
	suite.FundForAliasRegistration(rollapp)
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp.GetRollapp())
}
