package keeper_test

import (
	"cmp"
	"slices"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
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
			wantErr: gerrc.ErrPermissionDenied,
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
			wantErr: gerrc.ErrAlreadyExists,
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
				if tt.wantErr != nil {
					suite.Require().ErrorContains(err, tt.wantErr.Error())
				}
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

func (suite *RollappTestSuite) Test_msgServer_UpdateApp() {
	rollappID := urand.RollappID()

	tests := []struct {
		name     string
		msgs     []*types.MsgUpdateApp
		malleate func()
		wantErr  error
	}{
		{
			name: "success: update existing app",
			msgs: []*types.MsgUpdateApp{
				{
					Creator:     alice,
					Name:        "app1",
					RollappId:   rollappID,
					Description: "Updated description",
					Image:       "http://example.com/updated_image",
					Url:         "http://example.com/updated_app",
					Order:       2,
				},
			},
			malleate: func() {
				suite.App.RollappKeeper.SetApp(suite.Ctx, types.App{
					Name:      "app1",
					RollappId: rollappID,
					Order:     1,
				})
			},
		}, {
			name: "fail: update non-existent app",
			msgs: []*types.MsgUpdateApp{
				{
					Creator:     alice,
					Name:        "non_existent_app",
					RollappId:   rollappID,
					Description: "This app does not exist",
					Image:       "http://example.com/non_existent_image",
					Url:         "http://example.com/non_existent_app",
					Order:       1,
				},
			},
			wantErr: gerrc.ErrNotFound,
		}, {
			name: "fail: update app with different creator",
			msgs: []*types.MsgUpdateApp{
				{
					Creator:     bob,
					Name:        "app1",
					RollappId:   rollappID,
					Description: "Trying to update with a different creator",
					Image:       "http://example.com/different_creator_image",
					Url:         "http://example.com/different_creator_app",
					Order:       2,
				},
			},
			malleate: func() {
				suite.App.RollappKeeper.SetApp(suite.Ctx, types.App{
					Name:      "app1",
					RollappId: rollappID,
					Order:     1,
				})
			},
			wantErr: gerrc.ErrPermissionDenied,
		}, {
			name: "success: update multiple apps",
			msgs: []*types.MsgUpdateApp{
				{
					Creator:     alice,
					Name:        "app1",
					RollappId:   rollappID,
					Description: "Updated app1",
					Image:       "http://example.com/updated_image1",
					Url:         "http://example.com/updated_app1",
					Order:       3,
				}, {
					Creator:     alice,
					Name:        "app2",
					RollappId:   rollappID,
					Description: "Updated app2",
					Image:       "http://example.com/updated_image2",
					Url:         "http://example.com/updated_app2",
					Order:       1,
				},
			},
			malleate: func() {
				suite.App.RollappKeeper.SetApp(suite.Ctx, types.App{
					Name:      "app1",
					RollappId: rollappID,
					Order:     2,
				})
				suite.App.RollappKeeper.SetApp(suite.Ctx, types.App{
					Name:      "app2",
					RollappId: rollappID,
					Order:     1,
				})
			},
		}, {
			name: "fail: update app with different rollapp",
			msgs: []*types.MsgUpdateApp{
				{
					Creator:     alice,
					Name:        "app1",
					RollappId:   urand.RollappID(),
					Description: "Trying to update with a different rollapp",
					Image:       "http://example.com/different_rollapp_image",
					Url:         "http://example.com/different_rollapp_app",
					Order:       1,
				},
			},
			wantErr: types.ErrNotFound,
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
				_, err := suite.msgServer.UpdateApp(goCtx, msg)
				if tt.wantErr != nil {
					suite.Require().ErrorContains(err, tt.wantErr.Error())
				}
			}

			if tt.wantErr != nil {
				return
			}

			rollapp, err := suite.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{
				RollappId: rollappID,
			})
			suite.Require().NoError(err)
			suite.Require().Len(rollapp.Apps, len(tt.msgs))

			slices.SortFunc(tt.msgs, func(a, b *types.MsgUpdateApp) int {
				return cmp.Compare(a.Order, b.Order)
			})

			for i, app := range rollapp.Apps {
				suite.Require().Equal(tt.msgs[i].Order, app.Order)
				suite.Require().Equal(tt.msgs[i].Description, app.Description)
				suite.Require().Equal(tt.msgs[i].Image, app.Image)
				suite.Require().Equal(tt.msgs[i].Url, app.Url)
			}
		})
	}
}

func (suite *RollappTestSuite) Test_msgServer_RemoveApp() {
	rollappID := urand.RollappID()

	tests := []struct {
		name     string
		msgs     []*types.MsgRemoveApp
		malleate func()
		wantErr  error
	}{
		{
			name: "success: remove existing app",
			msgs: []*types.MsgRemoveApp{
				{
					Creator:   alice,
					Name:      "app1",
					RollappId: rollappID,
				},
			},
			malleate: func() {
				suite.App.RollappKeeper.SetApp(suite.Ctx, types.App{
					Name:      "app1",
					RollappId: rollappID,
				})
			},
		}, {
			name: "fail: remove non-existent app",
			msgs: []*types.MsgRemoveApp{
				{
					Creator:   alice,
					Name:      "non_existent_app",
					RollappId: rollappID,
				},
			},
			wantErr: gerrc.ErrNotFound,
		}, {
			name: "fail: remove app with different creator",
			msgs: []*types.MsgRemoveApp{
				{
					Creator:   bob,
					Name:      "app1",
					RollappId: rollappID,
				},
			},
			malleate: func() {
				suite.App.RollappKeeper.SetApp(suite.Ctx, types.App{
					Name:      "app1",
					RollappId: rollappID,
				})
			},
			wantErr: gerrc.ErrPermissionDenied,
		}, {
			name: "fail: remove app with different rollapp",
			msgs: []*types.MsgRemoveApp{
				{
					Creator:   alice,
					Name:      "app1",
					RollappId: urand.RollappID(),
				},
			},
			malleate: func() {
				suite.App.RollappKeeper.SetApp(suite.Ctx, types.App{
					Name:      "app1",
					RollappId: rollappID,
				})
			},
			wantErr: types.ErrNotFound,
		}, {
			name: "success: remove multiple apps",
			msgs: []*types.MsgRemoveApp{
				{
					Creator:   alice,
					Name:      "app1",
					RollappId: rollappID,
				}, {
					Creator:   alice,
					Name:      "app2",
					RollappId: rollappID,
				},
			},
			malleate: func() {
				suite.App.RollappKeeper.SetApp(suite.Ctx, types.App{
					Name:      "app1",
					RollappId: rollappID,
				})
				suite.App.RollappKeeper.SetApp(suite.Ctx, types.App{
					Name:      "app2",
					RollappId: rollappID,
				})
			},
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

			rollapp, err := suite.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{
				RollappId: rollappID,
			})
			suite.Require().NoError(err)
			createdAppsCount := len(rollapp.Apps)

			for _, msg := range tt.msgs {
				_, err := suite.msgServer.RemoveApp(goCtx, msg)
				if tt.wantErr != nil {
					suite.Require().ErrorContains(err, tt.wantErr.Error())
				}
			}

			if tt.wantErr != nil {
				return
			}

			rollapp, err = suite.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{
				RollappId: rollappID,
			})
			suite.Require().NoError(err)

			expectAppsCount := createdAppsCount - len(tt.msgs)
			suite.Require().Len(rollapp.Apps, expectAppsCount)

			for _, msg := range tt.msgs {
				_, found := suite.App.RollappKeeper.GetApp(suite.Ctx, msg.Name, msg.RollappId)
				suite.Require().False(found)
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
