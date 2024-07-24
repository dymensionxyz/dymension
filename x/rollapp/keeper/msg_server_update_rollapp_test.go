package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
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
					Address:   initialSequencer,
					RollappId: r.RollappId,
					Status:    sequencertypes.Bonded,
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
					Address:   initialSequencer,
					RollappId: r.RollappId,
					Status:    sequencertypes.Bonded,
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
					Address:   initialSequencer,
					RollappId: r.RollappId,
					Status:    sequencertypes.Bonded,
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
					Address:   initialSequencer,
					RollappId: r.RollappId,
					Status:    sequencertypes.Bonded,
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

func (suite *RollappTestSuite) TestCreateAndUpdateRollapp() {
	suite.SetupTest()

	const rollappId = "rollapp_1234-1"

	// 1. register rollapp
	err := suite.App.RollappKeeper.RegisterRollapp(suite.Ctx, types.Rollapp{
		RollappId:               rollappId,
		Creator:                 alice,
		GenesisChecksum:         "",
		InitialSequencerAddress: "",
		Alias:                   "",
		Bech32Prefix:            "rol",
	})
	suite.Require().NoError(err)

	// 2. try to register sequencer (not initial) - should fail because GenesisChecksum is empty
	_, err = suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	suite.Require().ErrorIs(err, sequencertypes.ErrGenesisChecksumNotSet)

	// 3. update rollapp immutable fields, set InitialSequencerAddress and GenesisChecksum
	initSeqPubKey := ed25519.GenPrivKey().PubKey()
	addrInit := sdk.AccAddress(initSeqPubKey.Address()).String()

	err = suite.App.RollappKeeper.UpdateRollapp(suite.Ctx, types.UpdateRollappInformation{
		Creator:                 alice,
		RollappId:               rollappId,
		InitialSequencerAddress: addrInit,
		GenesisChecksum:         "checksum1",
	})
	suite.Require().NoError(err)

	// 4. register sequencer (not initial) - should not be proposer
	addr2, err := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	suite.Require().NoError(err)
	seq1, ok := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2)
	suite.Require().True(ok)
	suite.Require().False(seq1.Proposer)

	// 5. register sequencer (initial) - should be proposer
	err = suite.CreateSequencer(suite.Ctx, rollappId, initSeqPubKey)
	suite.Require().NoError(err)
	initSeq, ok := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addrInit)
	suite.Require().True(ok)
	suite.Require().True(initSeq.Proposer)

	// 6. try to update rollapp immutable fields - should fail because sequencer is bonded
	err = suite.App.RollappKeeper.UpdateRollapp(suite.Ctx, types.UpdateRollappInformation{
		Creator:   alice,
		RollappId: rollappId,
		Alias:     "rolly",
	})
	suite.Require().ErrorIs(err, types.ErrImmutableFieldUpdateAfterInitialSequencerBonded)

	// 7. unbond initial sequencer
	_, err = suite.seqMsgServer.Unbond(suite.Ctx, &sequencertypes.MsgUnbond{
		Creator: addrInit,
	})
	suite.Require().NoError(err)

	// 8. update rollapp immutable fields
	err = suite.App.RollappKeeper.UpdateRollapp(suite.Ctx, types.UpdateRollappInformation{
		Creator:   alice,
		RollappId: rollappId,
		Alias:     "rolly",
	})
	suite.Require().NoError(err)
	updated, ok := suite.App.RollappKeeper.GetRollapp(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Require().Equal("rolly", updated.Alias)

	// 9. bond initial sequencer
	initSeq.Status = sequencertypes.Bonded
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, initSeq)

	// 10. try to update rollapp immutable fields - should fail because sequencer is bonded
	err = suite.App.RollappKeeper.UpdateRollapp(suite.Ctx, types.UpdateRollappInformation{
		Creator:                 alice,
		RollappId:               rollappId,
		InitialSequencerAddress: sample.AccAddress(),
	})
	suite.Require().ErrorIs(err, types.ErrImmutableFieldUpdateAfterInitialSequencerBonded)

	// 11. create state update
	suite.App.RollappKeeper.SetLatestStateInfoIndex(suite.Ctx, types.StateInfoIndex{
		RollappId: rollappId,
		Index:     1,
	})

	// 12. unbond initial sequencer
	initSeq.Status = sequencertypes.Unbonded
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, initSeq)

	// 13. try to update rollapp immutable fields - should fail because state is created
	err = suite.App.RollappKeeper.UpdateRollapp(suite.Ctx, types.UpdateRollappInformation{
		Creator:                 alice,
		RollappId:               rollappId,
		InitialSequencerAddress: sample.AccAddress(),
	})
	suite.Require().ErrorIs(err, types.ErrImmutableFieldUpdateAfterState)

	// 14. update initial sequencer
	metadata := sequencertypes.SequencerMetadata{
		Moniker:     "new_moniker",
		Details:     "something",
		P2PSeeds:    []string{"seed1", "seed2"},
		Rpcs:        []string{"rpc1", "rpc2"},
		EvmRpcs:     []string{"evm1", "evm2"},
		RestApiUrl:  "http://localhost:1317",
		ExplorerUrl: "http://localhost:8000",
		GenesisUrls: []string{"http://localhost:26657"},
		ContactDetails: &sequencertypes.ContactDetails{
			Website:  "https://dymension.xyz",
			Telegram: "sequencer",
			X:        "sequencer",
		},
		ExtraData: []byte("extra"),
		Snapshots: []*sequencertypes.SnapshotInfo{
			{
				SnapshotUrl: "http://localhost:1317/snapshot",
				Height:      123,
				Checksum:    "checksum",
			},
		},
		GasPrice: uptr.To(sdk.NewInt(100)),
	}
	_, err = suite.seqMsgServer.UpdateSequencerInformation(suite.Ctx, &sequencertypes.MsgUpdateSequencerInformation{
		Creator:   addrInit,
		RollappId: rollappId,
		Metadata:  metadata,
	})
	suite.Require().NoError(err)
	initSeq, ok = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addrInit)
	suite.Require().True(ok)
	suite.Require().Equal(metadata, initSeq.Metadata)
}
