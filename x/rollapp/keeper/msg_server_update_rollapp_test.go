package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

const (
	rollappId               = "rollapp_1234-1"
	initialSequencerAddress = "dym10l6edrf9gjv02um5kp7cmy4zgd26tafz6eqajz"
)

// TestUpdateRollapp tests updates for a basic non-launched, non-sealed rollapp
// It should allow updating the rollapp with any valid change
func (s *RollappTestSuite) TestUpdateRollapp() {
	gInfo := types.GenesisInfo{
		GenesisChecksum: "old",
		Bech32Prefix:    "old",
		NativeDenom: types.DenomMetadata{
			Display:  "OLD",
			Base:     "aold",
			Exponent: 18,
		},
		InitialSupply: sdk.NewInt(1000),
		Sealed:        false,
		GenesisAccounts: &types.GenesisAccounts{
			Accounts: []types.GenesisAccount{
				{
					Amount:  sdk.NewInt(1000),
					Address: initialSequencerAddress,
				},
			},
		},
	}

	tests := []struct {
		name     string
		update   *types.MsgUpdateRollappInformation
		expError error
		mallete  func(expected *types.Rollapp)
	}{
		{
			name: "Update rollapp: success - complete update",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
				MinSequencerBond: ucoin.SimpleMul(types.DefaultMinSequencerBondGlobalCoin, 3),
				Metadata:         &mockRollappMetadata,
				GenesisInfo: &types.GenesisInfo{
					Bech32Prefix:    "new",
					GenesisChecksum: "new_checksum",
					InitialSupply:   sdk.NewInt(1000),
					NativeDenom: types.DenomMetadata{
						Display:  "NEWDEN",
						Base:     "anewden",
						Exponent: 18,
					},
					GenesisAccounts: &types.GenesisAccounts{},
				},
			},
			mallete: func(expected *types.Rollapp) {
				expected.InitialSequencer = initialSequencerAddress
				expected.MinSequencerBond = sdk.NewCoins(ucoin.SimpleMul(types.DefaultMinSequencerBondGlobalCoin, 3))
				expected.Metadata = &mockRollappMetadata
				expected.GenesisInfo = types.GenesisInfo{
					Bech32Prefix:    "new",
					GenesisChecksum: "new_checksum",
					InitialSupply:   sdk.NewInt(1000),
					NativeDenom: types.DenomMetadata{
						Display:  "NEWDEN",
						Base:     "anewden",
						Exponent: 18,
					},
					GenesisAccounts: &types.GenesisAccounts{},
				}
			},
		},
		{
			name: "Update rollapp: success - only update initial sequencer",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
			},
			mallete: func(expected *types.Rollapp) {
				expected.InitialSequencer = initialSequencerAddress
			},
		},
		{
			name: "Update rollapp: success - update only metadata",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				Metadata:  &mockRollappMetadata,
			},
			expError: nil,
			mallete: func(expected *types.Rollapp) {
				expected.Metadata = &mockRollappMetadata
			},
		},
		{
			name: "Update rollapp: success - update only genesis info",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				GenesisInfo: &types.GenesisInfo{
					Bech32Prefix:    "new",
					GenesisChecksum: "new_checksum",
					InitialSupply:   sdk.NewInt(1000),
					NativeDenom: types.DenomMetadata{
						Display:  "NEWDEN",
						Base:     "anewden",
						Exponent: 18,
					},
					GenesisAccounts: &types.GenesisAccounts{},
				},
			},
			mallete: func(expected *types.Rollapp) {
				expected.GenesisInfo = types.GenesisInfo{
					Bech32Prefix:    "new",
					GenesisChecksum: "new_checksum",
					InitialSupply:   sdk.NewInt(1000),
					NativeDenom: types.DenomMetadata{
						Display:  "NEWDEN",
						Base:     "anewden",
						Exponent: 18,
					},
					GenesisAccounts: &types.GenesisAccounts{},
				}
			},
		},
		{
			name: "Update rollapp: success - clear genesis info",
			update: &types.MsgUpdateRollappInformation{
				Owner:       alice,
				RollappId:   rollappId,
				GenesisInfo: &types.GenesisInfo{},
			},
			mallete: func(expected *types.Rollapp) {
				expected.GenesisInfo = types.GenesisInfo{InitialSupply: sdk.NewInt(0)}
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
			name: "invalid bond",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				MinSequencerBond: types.DefaultMinSequencerBondGlobalCoin.SubAmount(sdk.NewInt(1)),
			},
			expError: gerrc.ErrInvalidArgument,
		},
		{
			name: "invalid metadata",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				Metadata:  &types.RollappMetadata{X: "not-url"},
			},
			expError: gerrc.ErrInvalidArgument,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			goCtx := sdk.WrapSDKContext(s.Ctx)

			rollapp := types.NewRollapp(alice, rollappId, "*", types.DefaultMinSequencerBondGlobalCoin, types.Rollapp_EVM, &types.RollappMetadata{}, gInfo)
			s.k().SetRollapp(s.Ctx, rollapp)

			_, err := s.msgServer.UpdateRollappInformation(goCtx, tc.update)
			if tc.expError == nil {
				s.Require().NoError(err)

				tc.mallete(&rollapp)

				resp, err := s.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{RollappId: tc.update.RollappId})
				s.Require().NoError(err)
				s.Equal(rollapp, resp.Rollapp)
			} else {
				s.ErrorIs(err, tc.expError)
			}
		})
	}
}

// TestUpdateRollappSealed tests update to the rollapp when the genesis info is sealed
// It should allow updating the rollapp only with non-genesis info data
func (s *RollappTestSuite) TestUpdateRollappSealed() {
	gInfo := types.GenesisInfo{
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
	}

	tests := []struct {
		name     string
		update   *types.MsgUpdateRollappInformation
		mallete  func(expected *types.Rollapp)
		expError error
	}{
		{
			name: "Update sealed rollapp: success - metadata update",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				Metadata:  &mockRollappMetadata,
			},
			mallete: func(expected *types.Rollapp) {
				expected.Metadata = &mockRollappMetadata
			},
			expError: nil,
		},
		{
			name: "Update sealed rollapp: success - initial sequencer update",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
			},
			mallete: func(expected *types.Rollapp) {
				expected.InitialSequencer = initialSequencerAddress
			},
			expError: nil,
		},
		{
			name: "Update sealed rollapp: fail - try to update genesis checksum when sealed",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				GenesisInfo: &types.GenesisInfo{
					GenesisChecksum: "new_checksum",
				},
			},
			expError: types.ErrGenesisInfoSealed,
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
			expError: types.ErrGenesisInfoSealed,
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
			expError: types.ErrGenesisInfoSealed,
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
			expError: types.ErrGenesisInfoSealed,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			goCtx := sdk.WrapSDKContext(s.Ctx)

			rollapp := types.NewRollapp(alice, rollappId, initialSequencerAddress, types.DefaultMinSequencerBondGlobalCoin, types.Rollapp_EVM, &types.RollappMetadata{}, gInfo)
			s.k().SetRollapp(s.Ctx, rollapp)

			_, err := s.msgServer.UpdateRollappInformation(goCtx, tc.update)
			if tc.expError == nil {
				s.Require().NoError(err)

				tc.mallete(&rollapp)

				resp, err := s.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{RollappId: tc.update.RollappId})
				s.Require().NoError(err)
				s.Equal(rollapp, resp.Rollapp)
			} else {
				s.ErrorIs(err, tc.expError)
			}
		})
	}
}

// TestUpdateRollappLaunched tests update to the rollapp when the rollapp is launched
func (s *RollappTestSuite) TestUpdateRollappLaunched() {
	gInfo := types.GenesisInfo{
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
	}

	tests := []struct {
		name     string
		update   *types.MsgUpdateRollappInformation
		mallete  func(expected *types.Rollapp)
		expError error
	}{
		{
			name: "Update sealed rollapp: success - metadata update",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				Metadata:  &mockRollappMetadata,
			},
			mallete: func(expected *types.Rollapp) {
				expected.Metadata = &mockRollappMetadata
			},
			expError: nil,
		},
		{
			name: "Update sealed rollapp: fail - initial sequencer update",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
			},
			expError: types.ErrImmutableFieldUpdateAfterLaunched,
		},
		{
			name: "Update sealed rollapp: fail - try to update genesis checksum when sealed",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				GenesisInfo: &types.GenesisInfo{
					GenesisChecksum: "new_checksum",
				},
			},
			expError: types.ErrImmutableFieldUpdateAfterLaunched,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			goCtx := sdk.WrapSDKContext(s.Ctx)

			rollapp := types.NewRollapp(alice, rollappId, initialSequencerAddress, types.DefaultMinSequencerBondGlobalCoin, types.Rollapp_EVM, &types.RollappMetadata{}, gInfo)
			rollapp.Launched = true
			s.k().SetRollapp(s.Ctx, rollapp)

			_, err := s.msgServer.UpdateRollappInformation(goCtx, tc.update)
			if tc.expError == nil {
				s.Require().NoError(err)

				tc.mallete(&rollapp)

				resp, err := s.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{RollappId: tc.update.RollappId})
				s.Require().NoError(err)
				s.Equal(rollapp, resp.Rollapp)
			} else {
				s.ErrorIs(err, tc.expError)
			}
		})
	}
}

// TestUpdateRollappUpdateGenesisInfo tests use case of updating genesis info
func (s *RollappTestSuite) TestUpdateRollappUpdateGenesisInfo() {
	// we start with empty genesis info
	gInfo := types.GenesisInfo{
		InitialSupply: sdk.NewInt(0),
	}

	tests := []struct {
		name     string
		update   *types.MsgUpdateRollappInformation
		mallete  func(expected *types.Rollapp)
		expError error
	}{
		{
			name: "Update genesis info: success - set initial native token",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
				GenesisInfo: &types.GenesisInfo{
					NativeDenom: types.DenomMetadata{
						Display:  "DEN",
						Base:     "aden",
						Exponent: 18,
					},
				},
			},
			mallete: func(expected *types.Rollapp) {
				expected.GenesisInfo.NativeDenom = types.DenomMetadata{
					Display:  "DEN",
					Base:     "aden",
					Exponent: 18,
				}
			},
			expError: nil,
		},
		{
			name: "Update genesis info: success - update checksum",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
				GenesisInfo: &types.GenesisInfo{
					GenesisChecksum: "checksum",
				},
			},
			mallete: func(expected *types.Rollapp) {
				expected.GenesisInfo.GenesisChecksum = "checksum"
			},
			expError: nil,
		},
		{
			name: "Update genesis info: success - native token with zero supply",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
				GenesisInfo: &types.GenesisInfo{
					NativeDenom: types.DenomMetadata{
						Display:  "DEN",
						Base:     "aden",
						Exponent: 18,
					},
					InitialSupply: sdk.NewInt(0),
				},
			},
			mallete: func(expected *types.Rollapp) {
				expected.GenesisInfo.NativeDenom = types.DenomMetadata{
					Display:  "DEN",
					Base:     "aden",
					Exponent: 18,
				}
				expected.GenesisInfo.InitialSupply = sdk.NewInt(0)
			},
			expError: nil,
		},
		{
			name: "Update genesis info: success - genesis accounts with valid total",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				GenesisInfo: &types.GenesisInfo{
					InitialSupply: sdk.NewInt(1000),
					NativeDenom: types.DenomMetadata{
						Display:  "DEN",
						Base:     "aden",
						Exponent: 18,
					},
					GenesisAccounts: &types.GenesisAccounts{
						Accounts: []types.GenesisAccount{
							{
								Address: initialSequencerAddress,
								Amount:  sdk.NewInt(500),
							},
						},
					},
				},
			},
			mallete: func(expected *types.Rollapp) {
				expected.GenesisInfo.InitialSupply = sdk.NewInt(1000)
				expected.GenesisInfo.NativeDenom = types.DenomMetadata{
					Display:  "DEN",
					Base:     "aden",
					Exponent: 18,
				}
				expected.GenesisInfo.GenesisAccounts = &types.GenesisAccounts{
					Accounts: []types.GenesisAccount{
						{
							Address: initialSequencerAddress,
							Amount:  sdk.NewInt(500),
						},
					},
				}
			},
			expError: nil,
		},
		{
			name: "Update genesis info: fail - invalid initial supply",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
				GenesisInfo: &types.GenesisInfo{
					InitialSupply: sdk.NewInt(-1), // Invalid negative supply
					NativeDenom: types.DenomMetadata{
						Display:  "DEN",
						Base:     "aden",
						Exponent: 18,
					},
				},
			},
			expError: types.ErrInvalidInitialSupply,
		},
		{
			name: "Update genesis info: fail - can't set supply without native token",
			update: &types.MsgUpdateRollappInformation{
				Owner:            alice,
				RollappId:        rollappId,
				InitialSequencer: initialSequencerAddress,
				GenesisInfo: &types.GenesisInfo{
					InitialSupply: sdk.NewInt(1000), // Non-zero supply without native token
				},
			},
			expError: types.ErrInvalidInitialSupply,
		},

		{
			name: "Update genesis info: fail - genesis accounts exceed total supply",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				GenesisInfo: &types.GenesisInfo{
					InitialSupply: sdk.NewInt(500),
					NativeDenom: types.DenomMetadata{
						Display:  "DEN",
						Base:     "aden",
						Exponent: 18,
					},
					GenesisAccounts: &types.GenesisAccounts{
						Accounts: []types.GenesisAccount{
							{
								Address: initialSequencerAddress,
								Amount:  sdk.NewInt(600),
							},
						},
					},
				},
			},
			expError: types.ErrInvalidInitialSupply,
		},
		{
			name: "Update genesis info: fail - duplicate genesis accounts",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				GenesisInfo: &types.GenesisInfo{
					Bech32Prefix:    "test",
					GenesisChecksum: "checksum",
					InitialSupply:   sdk.NewInt(1000),
					NativeDenom: types.DenomMetadata{
						Display:  "DEN",
						Base:     "aden",
						Exponent: 18,
					},
					GenesisAccounts: &types.GenesisAccounts{
						Accounts: []types.GenesisAccount{
							{
								Address: initialSequencerAddress,
								Amount:  sdk.NewInt(300),
							},
							{
								Address: initialSequencerAddress, // Duplicate address
								Amount:  sdk.NewInt(400),
							},
						},
					},
				},
			},
			expError: gerrc.ErrInvalidArgument,
		},
		{
			name: "Update genesis info: fail - invalid genesis account address",
			update: &types.MsgUpdateRollappInformation{
				Owner:     alice,
				RollappId: rollappId,
				GenesisInfo: &types.GenesisInfo{
					Bech32Prefix:    "test",
					GenesisChecksum: "checksum",
					InitialSupply:   sdk.NewInt(1000),
					NativeDenom: types.DenomMetadata{
						Display:  "DEN",
						Base:     "aden",
						Exponent: 18,
					},
					GenesisAccounts: &types.GenesisAccounts{
						Accounts: []types.GenesisAccount{
							{
								Address: "invalid_address",
								Amount:  sdk.NewInt(300),
							},
						},
					},
				},
			},
			expError: gerrc.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			goCtx := sdk.WrapSDKContext(s.Ctx)

			rollapp := types.NewRollapp(alice, rollappId, initialSequencerAddress, types.DefaultMinSequencerBondGlobalCoin, types.Rollapp_EVM, &types.RollappMetadata{}, gInfo)
			s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)

			_, err := s.msgServer.UpdateRollappInformation(goCtx, tt.update)
			if tt.expError == nil {
				s.Require().NoError(err)

				tt.mallete(&rollapp)

				resp, err := s.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{RollappId: tt.update.RollappId})
				s.Require().NoError(err)
				s.Equal(rollapp, resp.Rollapp)
			} else {
				s.ErrorIs(err, tt.expError)
			}
		})
	}
}

func (s *RollappTestSuite) TestUpdateRollappRegression() {
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

	gInfo := types.GenesisInfo{
		GenesisChecksum: "checksum",
		Bech32Prefix:    "rol",
		NativeDenom: types.DenomMetadata{
			Display:  "DEN",
			Base:     "aden",
			Exponent: 18,
		},
		InitialSupply: sdk.NewInt(1000),
	}

	// 1. register rollapp
	msg := types.MsgCreateRollapp{
		RollappId:        rollappId,
		Creator:          alice,
		InitialSequencer: "",
		MinSequencerBond: types.DefaultMinSequencerBondGlobalCoin,
		Alias:            "default",
		VmType:           types.Rollapp_EVM,
		GenesisInfo:      &gInfo,
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

	updatedgInfo := gInfo
	updatedgInfo.GenesisChecksum = "checksum1"
	_, err = s.msgServer.UpdateRollappInformation(s.Ctx, &types.MsgUpdateRollappInformation{
		Owner:            alice,
		RollappId:        rollappId,
		InitialSequencer: addrInit,
		GenesisInfo:      &updatedgInfo,
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
		GasPrice: "100",
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
