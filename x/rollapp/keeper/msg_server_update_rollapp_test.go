package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *RollappTestSuite) TestUpdateRollapp() {
	const (
		rollappId               = "rollapp_1234-1"
		initialSequencerAddress = "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz"
	)

	tests := []struct {
		name       string
		update     *types.MsgUpdateRollappInformation
		sealed     bool
		frozen     bool
		expError   error
		expRollapp types.Rollapp
	}{
		{
			name: "Update rollapp: success",
			update: &types.MsgUpdateRollappInformation{
				Owner:                   alice,
				RollappId:               rollappId,
				InitialSequencerAddress: initialSequencerAddress,
				Alias:                   "rolly",
				GenesisChecksum:         "new_checksum",
				Metadata:                &mockRollappMetadata,
			},
			expError: nil,
			expRollapp: types.Rollapp{
				Owner:                   alice,
				RollappId:               rollappId,
				InitialSequencerAddress: initialSequencerAddress,
				Bech32Prefix:            "rol",
				GenesisChecksum:         "new_checksum",
				Alias:                   "rolly",
				Metadata:                &mockRollappMetadata,
			},
		}, {
			name: "Update rollapp: fail - try to update a non-existing rollapp",
			update: &types.MsgUpdateRollappInformation{
				Owner:                   alice,
				RollappId:               "somerollapp_1235-1",
				InitialSequencerAddress: initialSequencerAddress,
			},
			expError: gerrc.ErrNotFound,
		}, {
			name: "Update rollapp: fail - try to update from non-creator address",
			update: &types.MsgUpdateRollappInformation{
				Owner:                   bob,
				RollappId:               rollappId,
				InitialSequencerAddress: initialSequencerAddress,
			},
			expError: sdkerrors.ErrUnauthorized,
		}, {
			name: "Update rollapp: fail - try to update a frozen rollapp",
			update: &types.MsgUpdateRollappInformation{
				Owner:                   alice,
				RollappId:               rollappId,
				InitialSequencerAddress: initialSequencerAddress,
			},
			frozen:   true,
			expError: types.ErrRollappFrozen,
		}, {
			name: "Update rollapp: fail - try to update InitialSequencerAddress when sealed",
			update: &types.MsgUpdateRollappInformation{
				Owner:                   alice,
				RollappId:               rollappId,
				InitialSequencerAddress: initialSequencerAddress,
			},
			sealed:   true,
			expError: types.ErrImmutableFieldUpdateAfterSealed,
		}, {
			name: "Update rollapp: fail - try to update alias when sealed",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				Alias:     "rolly",
			},
			sealed:   true,
			expError: types.ErrImmutableFieldUpdateAfterSealed,
		}, {
			name: "Update rollapp: fail - try to update genesis checksum when sealed",
			update: &types.MsgUpdateRollappInformation{
				Owner:           alice,
				RollappId:       rollappId,
				GenesisChecksum: "new_checksum",
			},
			sealed:   true,
			expError: types.ErrImmutableFieldUpdateAfterSealed,
		}, {
			name: "Update rollapp: success - update metadata when sealed",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				Metadata:  &mockRollappMetadata,
			},
			sealed:   true,
			expError: nil,
			expRollapp: types.Rollapp{
				RollappId:               rollappId,
				Owner:                   alice,
				InitialSequencerAddress: "",
				GenesisChecksum:         "checksum1",
				ChannelId:               "",
				Frozen:                  false,
				Bech32Prefix:            "rol",
				Alias:                   "Rollapp2",
				RegisteredDenoms:        nil,
				Sealed:                  true,
				Metadata:                &mockRollappMetadata,
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			goCtx := sdk.WrapSDKContext(suite.Ctx)
			rollapp := types.Rollapp{
				RollappId:               rollappId,
				Owner:                   alice,
				InitialSequencerAddress: "",
				GenesisChecksum:         "checksum1",
				ChannelId:               "",
				Frozen:                  tc.frozen,
				Sealed:                  tc.sealed,
				Bech32Prefix:            "rol",
				Alias:                   "Rollapp2",
				RegisteredDenoms:        nil,
				Metadata: &types.RollappMetadata{
					Website:          "",
					Description:      "",
					LogoDataUri:      "",
					TokenLogoDataUri: "",
					Telegram:         "",
					X:                "",
				},
			}

			suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

			_, err := suite.msgServer.UpdateRollappInformation(goCtx, tc.update)
			if tc.expError == nil {
				suite.Require().NoError(err)
				resp, err := suite.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{RollappId: tc.update.RollappId})
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
		Owner:                   alice,
		GenesisChecksum:         "",
		InitialSequencerAddress: "",
		Alias:                   "default",
		Bech32Prefix:            "rol",
	})
	suite.Require().NoError(err)

	// 2. try to register sequencer (not initial) - should fail because rollapp is not sealed
	_, err = suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	suite.Require().ErrorIs(err, sequencertypes.ErrNotInitialSequencer)

	// 3. update rollapp immutable fields, set InitialSequencerAddress, Alias and GenesisChecksum
	initSeqPubKey := ed25519.GenPrivKey().PubKey()
	addrInit := sdk.AccAddress(initSeqPubKey.Address()).String()

	err = suite.App.RollappKeeper.UpdateRollapp(suite.Ctx, &types.MsgUpdateRollappInformation{
		Owner:                   alice,
		RollappId:               rollappId,
		InitialSequencerAddress: addrInit,
		GenesisChecksum:         "checksum1",
		Alias:                   "alias",
	})
	suite.Require().NoError(err)

	// 4. register sequencer (initial) - should be proposer; rollapp should be sealed
	// from this point on, the rollapp is sealed and immutable fields cannot be updated
	err = suite.CreateSequencer(suite.Ctx, rollappId, initSeqPubKey)
	suite.Require().NoError(err)
	initSeq, ok := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addrInit)
	suite.Require().True(ok)
	suite.Require().True(initSeq.Proposer)
	rollapp, ok := suite.App.RollappKeeper.GetRollapp(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Require().True(rollapp.Sealed)

	// 5. try to update rollapp immutable fields - should fail because rollapp is sealed
	err = suite.App.RollappKeeper.UpdateRollapp(suite.Ctx, &types.MsgUpdateRollappInformation{
		Owner:     alice,
		RollappId: rollappId,
		Alias:     "rolly",
	})
	suite.Require().ErrorIs(err, types.ErrImmutableFieldUpdateAfterSealed)

	// 6. register another sequencer - should not be proposer
	newSeqAddr, err := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	suite.Require().NoError(err)
	newSequencer, ok := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, newSeqAddr)
	suite.Require().True(ok)
	suite.Require().False(newSequencer.Proposer)

	// 7. create state update
	suite.App.RollappKeeper.SetLatestStateInfoIndex(suite.Ctx, types.StateInfoIndex{
		RollappId: rollappId,
		Index:     1,
	})

	// 8. update initial sequencer
	metadata := sequencertypes.SequencerMetadata{
		Moniker:     "new_moniker",
		Details:     "something",
		P2PSeeds:    []string{"seed1", "seed2"},
		Rpcs:        []string{"rpc1", "rpc2"},
		EvmRpcs:     []string{"evm1", "evm2"},
		RestApiUrls: []string{"http://localhost:1317"},
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
