package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (s *RollappTestSuite) TestUpdateRollapp() {
	const (
		rollappId               = "rollapp_1234-1"
		initialSequencerAddress = "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz"
	)

	tests := []struct {
		name            string
		update          *types.MsgUpdateRollappInformation
		rollappLaunched bool
		genInfoSealed   bool
		expError        error
		expRollapp      types.Rollapp
	}{
		{
			name: "Update rollapp: success",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
				Metadata:         &mockRollappMetadata,
				GenesisInfo: &types.GenesisInfo{
					Bech32Prefix:    "new",
					GenesisChecksum: "new_checksum",
					InitialSupply:   sdk.NewInt(1000),
					NativeDenom: types.DenomMetadata{
						Display:  "DEN",
						Base:     "aden",
						Exponent: 18,
					},
					GenesisAccounts: &types.GenesisAccounts{}, // Frontend must specify empty type
				},
			},
			expError: nil,
			expRollapp: types.Rollapp{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
				MinSequencerBond: sdk.NewCoins(types.DefaultMinSequencerBondGlobalCoin),

				VmType:   types.Rollapp_EVM,
				Metadata: &mockRollappMetadata,
				GenesisInfo: types.GenesisInfo{
					GenesisChecksum: "new_checksum",
					Bech32Prefix:    "new",
					NativeDenom: types.DenomMetadata{
						Display:  "DEN",
						Base:     "aden",
						Exponent: 18,
					},
					InitialSupply:   sdk.NewInt(1000),
					GenesisAccounts: &types.GenesisAccounts{},
				},
			},
		},
		{
			name: "Update rollapp: fail - try to update a non-existing rollapp",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        "somerollapp_1235-1",
				InitialSequencer: initialSequencerAddress,
			},
			expError: gerrc.ErrNotFound,
		},
		{
			name: "Update rollapp: fail - try to update from non-creator address",
			update: &types.MsgUpdateRollappInformation{
				Owner:            bob,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
			},
			expError: sdkerrors.ErrUnauthorized,
		},
		{
			name: "Update rollapp: fail - try to update InitialSequencer when launched",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
			},
			rollappLaunched: true,
			expError:        types.ErrImmutableFieldUpdateAfterLaunched,
		},
		{
			name: "Update rollapp: fail - try to update min seq bond when launched",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				MinSequencerBond: types.DefaultMinSequencerBondGlobalCoin,
			},
			rollappLaunched: true,
			expError:        types.ErrImmutableFieldUpdateAfterLaunched,
		},
		{
			name: "invalid bond",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				MinSequencerBond: types.DefaultMinSequencerBondGlobalCoin.SubAmount(sdk.NewInt(1)),
			},
			expError: gerrc.ErrInvalidArgument,
		},
		{
			name: "Update rollapp: fail - try to update genesis checksum when sealed",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: "",
				Metadata:         nil,
				GenesisInfo: &types.GenesisInfo{
					GenesisChecksum: "new_checksum",
				},
			},
			genInfoSealed: true,
			expError:      types.ErrGenesisInfoSealed,
		},
		{
			name: "Update rollapp: fail - try to update bech32 when sealed",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				GenesisInfo: &types.GenesisInfo{
					Bech32Prefix: "new",
				},
			},
			genInfoSealed: true,
			expError:      types.ErrGenesisInfoSealed,
		},
		{
			name: "Update rollapp: fail - try to update native_denom when sealed",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				GenesisInfo: &types.GenesisInfo{
					NativeDenom: types.DenomMetadata{
						Display:  "DEN",
						Base:     "aden",
						Exponent: 18,
					},
				},
			},
			genInfoSealed: true,
			expError:      types.ErrGenesisInfoSealed,
		},
		{
			name: "Update rollapp: fail - try to update initial_supply when sealed",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				GenesisInfo: &types.GenesisInfo{
					InitialSupply: sdk.NewInt(1000),
				},
			},
			genInfoSealed: true,
			expError:      types.ErrGenesisInfoSealed,
		},
		{
			name: "Update rollapp: success - update metadata when sealed",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				Metadata:  &mockRollappMetadata,
			},
			rollappLaunched: true,
			genInfoSealed:   true,
			expError:        nil,
			expRollapp: types.Rollapp{
				RollappId:        rollappId,
				Owner:            alice,
				InitialSequencer: "",
				MinSequencerBond: sdk.NewCoins(types.DefaultMinSequencerBondGlobalCoin),
				ChannelId:        "",
				Launched:         true,
				VmType:           types.Rollapp_EVM,
				Metadata:         &mockRollappMetadata,
				GenesisInfo: types.GenesisInfo{
					Bech32Prefix:    "old",
					GenesisChecksum: "old",
					InitialSupply:   sdk.NewInt(1000),
					NativeDenom: types.DenomMetadata{
						Display:  "OLD",
						Base:     "aold",
						Exponent: 18,
					},
					GenesisAccounts: &types.GenesisAccounts{
						Accounts: []types.GenesisAccount{
							{
								Amount:  sdk.NewInt(1000),
								Address: initialSequencerAddress,
							},
						},
					},
					Sealed: true,
				},
			},
		},
		{
			name: "Update rollapp: success - unsealed, update rollapp without genesis info",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				Metadata:  &mockRollappMetadata,
			},
			rollappLaunched: false,
			genInfoSealed:   false,
			expError:        nil,
			expRollapp: types.Rollapp{
				RollappId:        rollappId,
				Owner:            alice,
				InitialSequencer: "",
				MinSequencerBond: sdk.NewCoins(types.DefaultMinSequencerBondGlobalCoin),
				ChannelId:        "",
				Launched:         false,
				VmType:           types.Rollapp_EVM,
				Metadata:         &mockRollappMetadata,
				GenesisInfo: types.GenesisInfo{
					Bech32Prefix:    "old",
					GenesisChecksum: "old",
					InitialSupply:   sdk.NewInt(1000),
					NativeDenom: types.DenomMetadata{
						Display:  "OLD",
						Base:     "aold",
						Exponent: 18,
					},
					GenesisAccounts: &types.GenesisAccounts{
						Accounts: []types.GenesisAccount{
							{
								Amount:  sdk.NewInt(1000),
								Address: initialSequencerAddress,
							},
						},
					},
					Sealed: false,
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			goCtx := sdk.WrapSDKContext(s.Ctx)
			rollapp := types.Rollapp{
				RollappId:        rollappId,
				Owner:            alice,
				InitialSequencer: "",
				MinSequencerBond: sdk.NewCoins(types.DefaultMinSequencerBondGlobalCoin),
				ChannelId:        "",
				Launched:         tc.rollappLaunched,
				VmType:           types.Rollapp_EVM,
				Metadata: &types.RollappMetadata{
					Website:     "",
					Description: "",
					LogoUrl:     "",
					Telegram:    "",
					X:           "",
				},
				GenesisInfo: types.GenesisInfo{
					GenesisChecksum: "old",
					Bech32Prefix:    "old",
					NativeDenom: types.DenomMetadata{
						Display:  "OLD",
						Base:     "aold",
						Exponent: 18,
					},
					InitialSupply: sdk.NewInt(1000),
					Sealed:        tc.genInfoSealed,
					GenesisAccounts: &types.GenesisAccounts{
						Accounts: []types.GenesisAccount{
							{
								Amount:  sdk.NewInt(1000),
								Address: initialSequencerAddress,
							},
						},
					},
				},
			}

			s.k().SetRollapp(s.Ctx, rollapp)

			_, err := s.msgServer.UpdateRollappInformation(goCtx, tc.update)
			if tc.expError == nil {
				s.Require().NoError(err)
				resp, err := s.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{RollappId: tc.update.RollappId})
				s.Require().NoError(err)
				s.Equal(tc.expRollapp, resp.Rollapp)
			} else {
				s.ErrorIs(err, tc.expError)
			}
		})
	}
}

// Update rollapp: fail - try to update genesis checksum when sealed
func (s *RollappTestSuite) TestUpdateRollappRegression() {
	const (
		rollappId               = "rollapp_1234-1"
		initialSequencerAddress = "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz"
	)

	goCtx := sdk.WrapSDKContext(s.Ctx)
	rollapp := types.Rollapp{
		RollappId:        rollappId,
		Owner:            alice,
		InitialSequencer: "",
		ChannelId:        "",
		Launched:         true,
		VmType:           types.Rollapp_EVM,
		Metadata: &types.RollappMetadata{
			Website:     "",
			Description: "",
			LogoUrl:     "",
			Telegram:    "",
			X:           "",
		},
		GenesisInfo: types.GenesisInfo{
			GenesisChecksum: "old",
			Bech32Prefix:    "old",
			NativeDenom: types.DenomMetadata{
				Display:  "OLD",
				Base:     "aold",
				Exponent: 18,
			},
			InitialSupply: sdk.NewInt(1000),
			Sealed:        true,
			GenesisAccounts: &types.GenesisAccounts{
				Accounts: []types.GenesisAccount{
					{
						Amount:  sdk.NewInt(1000),
						Address: initialSequencerAddress,
					},
				},
			},
		},
	}

	update := &types.MsgUpdateRollappInformation{
		Owner:            alice,
		RollappId:        rollappId,
		InitialSequencer: "",
		Metadata:         nil,
		GenesisInfo: &types.GenesisInfo{
			GenesisAccounts: nil,
		},
	}

	s.k().SetRollapp(s.Ctx, rollapp)

	_, err := s.msgServer.UpdateRollappInformation(goCtx, update)
	s.ErrorIs(err, types.ErrGenesisInfoSealed)

	expect := len(rollapp.GenesisInfo.Accounts())
	ra := s.k().MustGetRollapp(s.Ctx, rollappId)
	s.Require().Equal(expect, len(ra.GenesisInfo.Accounts()))
}

func (s *RollappTestSuite) TestCreateAndUpdateRollapp() {
	const rollappId = "rollapp_1234-1"

	// 1. register rollapp
	msg := types.MsgCreateRollapp{
		RollappId:        rollappId,
		Creator:          alice,
		InitialSequencer: "",
		MinSequencerBond: types.DefaultMinSequencerBondGlobalCoin,
		Alias:            "default",
		VmType:           types.Rollapp_EVM,
		GenesisInfo: &types.GenesisInfo{
			Bech32Prefix:    "rol",
			GenesisChecksum: "checksum",
			InitialSupply:   sdk.NewInt(1000),
			NativeDenom: types.DenomMetadata{
				Display:  "DEN",
				Base:     "aden",
				Exponent: 18,
			},
		},
	}
	s.FundForAliasRegistration(msg)
	_, err := s.msgServer.CreateRollapp(s.Ctx, &msg)
	s.Require().NoError(err)

	// 2. try to register sequencer (not initial) - should fail because rollapp is not launched
	err = s.CreateSequencerByPubkey(s.Ctx, rollappId, ed25519.GenPrivKey().PubKey())
	s.Require().ErrorIs(err, sequencertypes.ErrNotInitialSequencer)

	// 3. update rollapp immutable fields, set InitialSequencer, Alias and GenesisChecksum
	initSeqPubKey := ed25519.GenPrivKey().PubKey()
	addrInit := sdk.AccAddress(initSeqPubKey.Address()).String()

	_, err = s.msgServer.UpdateRollappInformation(s.Ctx, &types.MsgUpdateRollappInformation{
		Owner:            alice,
		RollappId:        rollappId,
		InitialSequencer: addrInit,
		GenesisInfo:      &types.GenesisInfo{GenesisChecksum: "checksum1"},
	})
	s.Require().NoError(err)

	// 4. register sequencer (initial) - should be proposer; rollapp should be launched
	// from this point on, the rollapp is launched and immutable fields cannot be updated
	err = s.CreateSequencerByPubkey(s.Ctx, rollappId, initSeqPubKey)
	s.Require().NoError(err)
	initSeq, err := s.App.SequencerKeeper.RealSequencer(s.Ctx, addrInit)
	s.Require().NoError(err)
	proposer := s.App.SequencerKeeper.GetProposer(s.Ctx, rollappId)
	s.Require().Equal(initSeq, proposer)
	rollapp, ok := s.k().GetRollapp(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Require().True(rollapp.Launched)

	// 5. try to update rollapp immutable fields - should fail because rollapp is launched
	_, err = s.msgServer.UpdateRollappInformation(s.Ctx, &types.MsgUpdateRollappInformation{
		Owner:            alice,
		RollappId:        rollappId,
		InitialSequencer: "new",
	})
	s.Require().ErrorIs(err, types.ErrImmutableFieldUpdateAfterLaunched)

	// 6. register another sequencer - should not be proposer
	newSeqAddr := s.CreateDefaultSequencer(s.Ctx, rollappId)
	proposer = s.App.SequencerKeeper.GetProposer(s.Ctx, rollappId)
	s.Require().NotEqual(proposer, newSeqAddr)

	// 7. create state update
	s.k().SetLatestStateInfoIndex(s.Ctx, types.StateInfoIndex{
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
	_, err = s.seqMsgServer.UpdateSequencerInformation(s.Ctx, &sequencertypes.MsgUpdateSequencerInformation{
		Creator:  addrInit,
		Metadata: metadata,
	})
	s.Require().NoError(err)
	initSeq, err = s.App.SequencerKeeper.RealSequencer(s.Ctx, addrInit)
	s.Require().NoError(err)
	s.Require().Equal(metadata, initSeq.Metadata)
}

func (s *RollappTestSuite) TestForceGenesisInfoChange() {
	govModuleAccount := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	tests := []struct {
		name     string
		msg      *types.MsgForceGenesisInfoChange
		expError error
	}{
		{
			name: "happy path - valid genesis info change",
			msg: &types.MsgForceGenesisInfoChange{
				Authority: govModuleAccount,
				NewGenesisInfo: types.GenesisInfo{
					Bech32Prefix:    "new",
					GenesisChecksum: "new_checksum",
					InitialSupply:   sdk.NewInt(2000),
					NativeDenom: types.DenomMetadata{
						Display:  "NEW",
						Base:     "anew",
						Exponent: 18,
					},
				},
			},
			expError: nil,
		},
		{
			name: "missing bech32 prefix",
			msg: &types.MsgForceGenesisInfoChange{
				Authority: govModuleAccount,
				NewGenesisInfo: types.GenesisInfo{
					GenesisChecksum: "new_checksum",
					InitialSupply:   sdk.NewInt(2000),
					NativeDenom: types.DenomMetadata{
						Display:  "NEW",
						Base:     "anew",
						Exponent: 18,
					},
				},
			},
			expError: gerrc.ErrInvalidArgument,
		},
		{
			name: "missing native denom",
			msg: &types.MsgForceGenesisInfoChange{
				Authority: govModuleAccount,
				NewGenesisInfo: types.GenesisInfo{
					Bech32Prefix:    "new",
					GenesisChecksum: "new_checksum",
					InitialSupply:   sdk.NewInt(2000),
				},
			},
			expError: gerrc.ErrInvalidArgument,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			rollappId, _ := s.CreateDefaultRollappAndProposer()
			tc.msg.RollappId = rollappId

			// Verify rollapp was created with sealed genesis info
			rollapp, found := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappId)
			s.Require().True(found)
			s.Require().True(rollapp.GenesisInfo.Sealed)

			// Execute the message
			_, err := s.App.RollappKeeper.ForceGenesisInfoChange(s.Ctx, tc.msg)

			if tc.expError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expError)
				return
			}

			s.Require().NoError(err)

			// Verify genesis info was changed correctly
			rollapp, found = s.App.RollappKeeper.GetRollapp(s.Ctx, tc.msg.RollappId)
			s.Require().True(found)
			s.Require().Equal(tc.msg.NewGenesisInfo.Bech32Prefix, rollapp.GenesisInfo.Bech32Prefix)
			s.Require().Equal(tc.msg.NewGenesisInfo.GenesisChecksum, rollapp.GenesisInfo.GenesisChecksum)
			s.Require().Equal(tc.msg.NewGenesisInfo.InitialSupply, rollapp.GenesisInfo.InitialSupply)
			s.Require().Equal(tc.msg.NewGenesisInfo.NativeDenom.Display, rollapp.GenesisInfo.NativeDenom.Display)
			s.Require().Equal(tc.msg.NewGenesisInfo.NativeDenom.Base, rollapp.GenesisInfo.NativeDenom.Base)
			s.Require().Equal(tc.msg.NewGenesisInfo.NativeDenom.Exponent, rollapp.GenesisInfo.NativeDenom.Exponent)
			s.Require().True(rollapp.GenesisInfo.Sealed)
		})
	}
}
