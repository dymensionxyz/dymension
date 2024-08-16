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
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
				GenesisChecksum:  "new_checksum",
				Metadata:         &mockRollappMetadata,
				Bech32Prefix:     "new",
			},
			expError: nil,
			expRollapp: types.Rollapp{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
				Bech32Prefix:     "new",
				GenesisChecksum:  "new_checksum",
				VmType:           types.Rollapp_EVM,
				Metadata:         &mockRollappMetadata,
			},
		}, {
			name: "Update rollapp: fail - try to update a non-existing rollapp",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        "somerollapp_1235-1",
				InitialSequencer: initialSequencerAddress,
			},
			expError: gerrc.ErrNotFound,
		}, {
			name: "Update rollapp: fail - try to update from non-creator address",
			update: &types.MsgUpdateRollappInformation{
				Owner:            bob,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
			},
			expError: sdkerrors.ErrUnauthorized,
		}, {
			name: "Update rollapp: fail - try to update a frozen rollapp",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
			},
			frozen:   true,
			expError: types.ErrRollappFrozen,
		}, {
			name: "Update rollapp: fail - try to update InitialSequencer when sealed",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
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
			name: "Update rollapp: fail - try to update bech 32 when sealed",
			update: &types.MsgUpdateRollappInformation{
				Owner:        alice,
				RollappId:    rollappId,
				Bech32Prefix: "new",
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
				RollappId:        rollappId,
				Owner:            alice,
				InitialSequencer: "",
				ChannelId:        "",
				Frozen:           false,
				RegisteredDenoms: nil,
				Sealed:           true,
				VmType:           types.Rollapp_EVM,
				Metadata:         &mockRollappMetadata,
				Bech32Prefix:     "old",
				GenesisChecksum:  "old",
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			goCtx := sdk.WrapSDKContext(suite.Ctx)
			rollapp := types.Rollapp{
				RollappId:        rollappId,
				Owner:            alice,
				InitialSequencer: "",
				ChannelId:        "",
				Frozen:           tc.frozen,
				Sealed:           tc.sealed,
				RegisteredDenoms: nil,
				VmType:           types.Rollapp_EVM,
				Bech32Prefix:     "old",
				GenesisChecksum:  "old",
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
	const rollappId = "rollapp_1234-1"

	// 1. register rollapp
	msg := types.MsgCreateRollapp{
		RollappId:        rollappId,
		Creator:          alice,
		GenesisChecksum:  "",
		InitialSequencer: "",
		Alias:            "default",
		VmType:           types.Rollapp_EVM,
		Bech32Prefix:     "rol",
	}
	suite.FundForAliasRegistration(msg)
	_, err := suite.msgServer.CreateRollapp(suite.Ctx, &msg)
	suite.Require().NoError(err)

	// 2. try to register sequencer (not initial) - should fail because rollapp is not sealed
	err = suite.CreateSequencerByPubkey(suite.Ctx, rollappId, ed25519.GenPrivKey().PubKey())
	suite.Require().ErrorIs(err, sequencertypes.ErrNotInitialSequencer)

	// 3. update rollapp immutable fields, set InitialSequencer, Alias and GenesisChecksum
	initSeqPubKey := ed25519.GenPrivKey().PubKey()
	addrInit := sdk.AccAddress(initSeqPubKey.Address()).String()

	_, err = suite.msgServer.UpdateRollappInformation(suite.Ctx, &types.MsgUpdateRollappInformation{
		Owner:            alice,
		RollappId:        rollappId,
		InitialSequencer: addrInit,
		GenesisChecksum:  "checksum1",
	})
	suite.Require().NoError(err)

	// 4. register sequencer (initial) - should be proposer; rollapp should be sealed
	// from this point on, the rollapp is sealed and immutable fields cannot be updated
	err = suite.CreateSequencerByPubkey(suite.Ctx, rollappId, initSeqPubKey)
	suite.Require().NoError(err)
	initSeq, ok := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addrInit)
	suite.Require().True(ok)
	proposer, found := suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().True(found)
	suite.Require().Equal(initSeq, proposer)
	rollapp, ok := suite.App.RollappKeeper.GetRollapp(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Require().True(rollapp.Sealed)

	// 5. try to update rollapp immutable fields - should fail because rollapp is sealed
	_, err = suite.msgServer.UpdateRollappInformation(suite.Ctx, &types.MsgUpdateRollappInformation{
		Owner:           alice,
		RollappId:       rollappId,
		GenesisChecksum: "checksum2",
	})
	suite.Require().ErrorIs(err, types.ErrImmutableFieldUpdateAfterSealed)

	// 6. register another sequencer - should not be proposer
	newSeqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	proposer, found = suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().True(found)
	suite.Require().NotEqual(proposer, newSeqAddr)

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
		Rpcs:        []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
		EvmRpcs:     []string{"https://rpc.evm.rollapp.noisnemyd.xyz:443"},
		RestApiUrls: []string{"http://localhost:1317"},
		ExplorerUrl: "http://localhost:8000",
		GenesisUrls: []string{"http://localhost:26657"},
		ContactDetails: &sequencertypes.ContactDetails{
			Website:  "https://dymension.xyz",
			Telegram: "https://t.me/rolly",
			X:        "https://x.dymension.xyz",
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
