package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *RollappTestSuite) TestUpdateRollapp() {
	tests := []struct {
		name       string
		update     *types.MsgUpdateRollappInformation
		malleate   func(types.Rollapp) types.Rollapp
		expError   error
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
					Metadata:                &mockRollappMetadata,
				},
			},
			expError: nil,
			expRollapp: types.Rollapp{
				Creator:                 alice,
				RollappId:               "rollapp_1234-1",
				InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
				Bech32Prefix:            "rol",
				GenesisChecksum:         "new_checksum",
				Alias:                   "rolly",
				Metadata:                &mockRollappMetadata,
			},
		}, {
			name: "Update rollapp: fail - try to update a non-existing rollapp",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:                 alice,
					RollappId:               "somerollapp_1235-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
				},
			},
			expError: gerrc.ErrNotFound,
		}, {
			name: "Update rollapp: fail - try to update from non-creator address",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:                 bob,
					RollappId:               "rollapp_1234-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
				},
			},
			expError: sdkerrors.ErrUnauthorized,
		}, {
			name: "Update rollapp: fail - try to update a frozen rollapp",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:                 alice,
					RollappId:               "rollapp_1234-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
				},
			},
			malleate: func(r types.Rollapp) types.Rollapp {
				r.Frozen = true
				return r
			},
			expError: types.ErrRollappFrozen,
		}, {
			name: "Update rollapp: fail - try to update using another rollapp's alias",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:                 alice,
					RollappId:               "rollapp_1234-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
					Alias:                   "rolly",
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
			expError: gerrc.ErrAlreadyExists,
		}, {
			name: "Update rollapp: fail - try to update InitialSequencerAddress with existing state",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:                 alice,
					RollappId:               "rollapp_1234-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
				},
			},
			malleate: func(r types.Rollapp) types.Rollapp {
				// create state for the rollapp
				suite.App.RollappKeeper.SetLatestStateInfoIndex(suite.Ctx, types.StateInfoIndex{
					RollappId: r.RollappId,
					Index:     1,
				})
				return r
			},
			expError: types.ErrImmutableFieldUpdateAfterState,
		}, {
			name: "Update rollapp: fail - try to update InitialSequencerAddress with bonded initial sequencer",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:                 alice,
					RollappId:               "rollapp_1234-1",
					InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
				},
			},
			malleate: func(r types.Rollapp) types.Rollapp {
				// create initial bonded sequencer
				initialSequencer := "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz"
				suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencertypes.Sequencer{
					SequencerAddress: initialSequencer,
					RollappId:        r.RollappId,
					Status:           sequencertypes.Bonded,
				})
				r.InitialSequencerAddress = initialSequencer
				return r
			},
			expError: types.ErrImmutableFieldUpdateAfterInitialSequencerBonded,
		}, {
			name: "Update rollapp: fail - try to update alias with existing state",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:   alice,
					RollappId: "rollapp_1234-1",
					Alias:     "rolly",
				},
			},
			malleate: func(r types.Rollapp) types.Rollapp {
				// create state for the rollapp
				suite.App.RollappKeeper.SetLatestStateInfoIndex(suite.Ctx, types.StateInfoIndex{
					RollappId: r.RollappId,
					Index:     1,
				})
				return r
			},
			expError: types.ErrImmutableFieldUpdateAfterState,
		}, {
			name: "Update rollapp: fail - try to update alias with bonded initial sequencer",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:   alice,
					RollappId: "rollapp_1234-1",
					Alias:     "rolly",
				},
			},
			malleate: func(r types.Rollapp) types.Rollapp {
				// create initial bonded sequencer
				initialSequencer := "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz"
				suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencertypes.Sequencer{
					SequencerAddress: initialSequencer,
					RollappId:        r.RollappId,
					Status:           sequencertypes.Bonded,
				})
				r.InitialSequencerAddress = initialSequencer
				return r
			},
			expError: types.ErrImmutableFieldUpdateAfterInitialSequencerBonded,
		}, {
			name: "Update rollapp: fail - try to update genesis checksum with existing state",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:         alice,
					RollappId:       "rollapp_1234-1",
					GenesisChecksum: "new_checksum",
				},
			},
			malleate: func(r types.Rollapp) types.Rollapp {
				// create state for the rollapp
				suite.App.RollappKeeper.SetLatestStateInfoIndex(suite.Ctx, types.StateInfoIndex{
					RollappId: "rollapp_1234-1",
					Index:     1,
				})
				return r
			},
			expError: types.ErrImmutableFieldUpdateAfterState,
		}, {
			name: "Update rollapp: fail - try to update genesis checksum with bonded initial sequencer",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:         alice,
					RollappId:       "rollapp_1234-1",
					GenesisChecksum: "new_checksum",
				},
			},
			malleate: func(r types.Rollapp) types.Rollapp {
				// create initial bonded sequencer
				initialSequencer := "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz"
				suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencertypes.Sequencer{
					SequencerAddress: initialSequencer,
					RollappId:        r.RollappId,
					Status:           sequencertypes.Bonded,
				})
				r.InitialSequencerAddress = initialSequencer
				return r
			},
			expError: types.ErrImmutableFieldUpdateAfterInitialSequencerBonded,
		}, {
			name: "Update rollapp: success - update metadata with existing state",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:   alice,
					RollappId: "rollapp_1234-1",
					Metadata:  &mockRollappMetadata,
				},
			},
			malleate: func(r types.Rollapp) types.Rollapp {
				// create state for the rollapp
				suite.App.RollappKeeper.SetLatestStateInfoIndex(suite.Ctx, types.StateInfoIndex{
					RollappId: "rollapp_1234-1",
					Index:     1,
				})
				return r
			},
			expError: nil,
			expRollapp: types.Rollapp{
				RollappId:               "rollapp_1234-1",
				Creator:                 alice,
				InitialSequencerAddress: "",
				GenesisChecksum:         "checksum1",
				ChannelId:               "",
				Frozen:                  false,
				Bech32Prefix:            "rol",
				Alias:                   "Rollapp2",
				RegisteredDenoms:        nil,
				Metadata:                &mockRollappMetadata,
			},
		}, {
			name: "Update rollapp: success - update metadata with bonded initial sequencer",
			update: &types.MsgUpdateRollappInformation{
				Update: &types.UpdateRollappInformation{
					Creator:   alice,
					RollappId: "rollapp_1234-1",
					Metadata:  &mockRollappMetadata,
				},
			},
			malleate: func(r types.Rollapp) types.Rollapp {
				// create initial bonded sequencer
				initialSequencer := "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz"
				suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencertypes.Sequencer{
					SequencerAddress: initialSequencer,
					RollappId:        r.RollappId,
					Status:           sequencertypes.Bonded,
				})
				r.InitialSequencerAddress = initialSequencer
				return r
			},
			expError: nil,
			expRollapp: types.Rollapp{
				RollappId:               "rollapp_1234-1",
				Creator:                 alice,
				InitialSequencerAddress: "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz",
				GenesisChecksum:         "checksum1",
				ChannelId:               "",
				Frozen:                  false,
				Bech32Prefix:            "rol",
				Alias:                   "Rollapp2",
				RegisteredDenoms:        nil,
				Metadata:                &mockRollappMetadata,
			},
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
					Telegram:     "",
					X:            "",
				},
			}

			if tc.malleate != nil {
				rollapp = tc.malleate(rollapp)
			}

			suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

			_, err := suite.msgServer.UpdateRollappInformation(goCtx, tc.update)
			if tc.expError == nil {
				suite.Require().NoError(err)
				resp, err := suite.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{RollappId: tc.update.Update.RollappId})
				suite.Require().NoError(err)
				suite.Equal(tc.expRollapp, resp.Rollapp)
			} else {
				suite.ErrorIs(err, tc.expError)
			}
		})
	}
}
