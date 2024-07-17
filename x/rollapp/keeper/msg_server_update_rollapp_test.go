package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (suite *RollappTestSuite) TestUpdateRollapp() {
	tests := []struct {
		name       string
		update     *types.MsgUpdateRollappInformation
		malleate   func(types.Rollapp) types.Rollapp
		expPass    bool
		expRollapp types.Rollapp
	}{
		{
			name: "Update rollapp: success",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:                 alice,
					RollappId:               "rollapp_1234-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
					Alias:                   "rolly",
					GenesisChecksum:         "new_checksum",
					Metadata: &types.RollappMetadata{
						Website:      "https://dymension.xyz",
						Description:  "Sample description",
						LogoDataUri:  "data:image/png;base64,c2lzZQ==",
						TokenLogoUri: "data:image/png;base64,ZHVwZQ==",
					},
				},
			},
			expPass: true,
			expRollapp: types.Rollapp{
				Creator:                 alice,
				RollappId:               "rollapp_1234-1",
				InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
				Bech32Prefix:            "rol",
				GenesisChecksum:         "new_checksum",
				Alias:                   "rolly",
				Metadata: &types.RollappMetadata{
					Website:      "https://dymension.xyz",
					Description:  "Sample description",
					LogoDataUri:  "data:image/png;base64,c2lzZQ==",
					TokenLogoUri: "data:image/png;base64,ZHVwZQ==",
				},
			},
		}, {
			name: "Update rollapp: fail - try to update a non-existing rollapp",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:                 alice,
					RollappId:               "somerollapp_1235-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
					Alias:                   "rolly",
					GenesisChecksum:         "new_checksum",
					Metadata: &types.RollappMetadata{
						Website:      "https://dymension.xyz",
						Description:  "Sample description",
						LogoDataUri:  "data:image/png;base64,c2lzZQ==",
						TokenLogoUri: "data:image/png;base64,ZHVwZQ==",
					},
				},
			},
			expPass: false,
		}, {
			name: "Update rollapp: fail - try to update from non-creator address",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:                 bob,
					RollappId:               "rollapp_1234-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
					Alias:                   "rolly",
					GenesisChecksum:         "new_checksum",
					Metadata: &types.RollappMetadata{
						Website:      "https://dymension.xyz",
						Description:  "Sample description",
						LogoDataUri:  "data:image/png;base64,c2lzZQ==",
						TokenLogoUri: "data:image/png;base64,ZHVwZQ==",
					},
				},
			},
			expPass: false,
		}, {
			name: "Update rollapp: fail - try to update a frozen rollapp",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:                 alice,
					RollappId:               "rollapp_1234-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
					Alias:                   "rolly",
					GenesisChecksum:         "new_checksum",
					Metadata: &types.RollappMetadata{
						Website:      "https://dymension.xyz",
						Description:  "Sample description",
						LogoDataUri:  "data:image/png;base64,c2lzZQ==",
						TokenLogoUri: "data:image/png;base64,ZHVwZQ==",
					},
				},
			},
			malleate: func(r types.Rollapp) types.Rollapp {
				r.Frozen = true
				return r
			},
			expPass: false,
		}, {
			name: "Update rollapp: fail - try to update non-empty InitialSequencerAddress",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:                 alice,
					RollappId:               "rollapp_1234-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
					Alias:                   "rolly",
					GenesisChecksum:         "new_checksum",
					Metadata: &types.RollappMetadata{
						Website:      "https://dymension.xyz",
						Description:  "Sample description",
						LogoDataUri:  "data:image/png;base64,c2lzZQ==",
						TokenLogoUri: "data:image/png;base64,ZHVwZQ==",
					},
				},
			},
			malleate: func(r types.Rollapp) types.Rollapp {
				r.InitialSequencerAddress = "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz"
				return r
			},
			expPass: false,
		}, {
			name: "Update rollapp: fail - try to update using another rollapp's InitialSequencerAddress",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:                 alice,
					RollappId:               "rollapp_1234-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
					Alias:                   "rolly",
					GenesisChecksum:         "new_checksum",
					Metadata: &types.RollappMetadata{
						Website:      "https://dymension.xyz",
						Description:  "Sample description",
						LogoDataUri:  "data:image/png;base64,c2lzZQ==",
						TokenLogoUri: "data:image/png;base64,ZHVwZQ==",
					},
				},
			},
			malleate: func(r types.Rollapp) types.Rollapp {
				// create another rollapp with the same InitialSequencerAddress
				suite.App.RollappKeeper.SetRollapp(suite.Ctx, types.Rollapp{
					RollappId:               "somerollapp_1235-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
				})
				return r
			},
			expPass: false,
		}, {
			name: "Update rollapp: fail - try to update using another rollapp's alias",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:                 alice,
					RollappId:               "rollapp_1234-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
					Alias:                   "rolly",
					GenesisChecksum:         "new_checksum",
					Metadata: &types.RollappMetadata{
						Website:      "https://dymension.xyz",
						Description:  "Sample description",
						LogoDataUri:  "data:image/png;base64,c2lzZQ==",
						TokenLogoUri: "data:image/png;base64,ZHVwZQ==",
					},
				},
			},
			malleate: func(r types.Rollapp) types.Rollapp {
				// create another rollapp with the same InitialSequencerAddress
				suite.App.RollappKeeper.SetRollapp(suite.Ctx, types.Rollapp{
					RollappId: "somerollapp_1235-1",
					Alias:     "rolly",
				})
				return r
			},
			expPass: false,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			goCtx := sdk.WrapSDKContext(suite.Ctx)
			rollapp := types.Rollapp{
				RollappId:               "rollapp_1234-1",
				Creator:                 alice,
				InitialSequencerAddress: "",
				GenesisChecksum:         "checksum1",
				ChannelId:               "",
				Frozen:                  false,
				Bech32Prefix:            "rol",
				Alias:                   "Rollapp2",
				RegisteredDenoms:        nil,
				Metadata: &types.RollappMetadata{
					Website:      "",
					Description:  "",
					LogoDataUri:  "",
					TokenLogoUri: "",
				},
			}

			if tc.malleate != nil {
				rollapp = tc.malleate(rollapp)
			}

			suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

			_, err := suite.msgServer.UpdateRollappInformation(goCtx, tc.update)
			if tc.expPass {
				suite.NoError(err)
				resp, err := suite.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{RollappId: tc.update.Update.RollappId})
				suite.NoError(err)
				suite.Equal(tc.expRollapp, resp.Rollapp)
			} else {
				suite.Error(err)
			}
		})
	}
}
