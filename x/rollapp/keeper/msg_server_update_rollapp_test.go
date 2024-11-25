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
				},
			},
			expError: nil,
			expRollapp: types.Rollapp{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
				VmType:           types.Rollapp_EVM,
				Metadata:         &mockRollappMetadata,
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
			name: "Update rollapp: fail - try to update InitialSequencer when launched",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
			},
			rollappLaunched: true,
			expError:        types.ErrImmutableFieldUpdateAfterLaunched,
		}, {
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
		}, {
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
		}, {
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
		}, {
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
		}, {
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
		}, {
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

func (s *RollappTestSuite) TestCreateAndUpdateRollapp() {
	const rollappId = "rollapp_1234-1"

	// 1. register rollapp
	msg := types.MsgCreateRollapp{
		RollappId:        rollappId,
		Creator:          alice,
		InitialSequencer: "",
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
