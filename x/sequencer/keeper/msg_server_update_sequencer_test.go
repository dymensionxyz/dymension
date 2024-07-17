package keeper_test

import (
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestUpdateSequencer() {
	pubkey := ed25519.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	pkAny, err := codectypes.NewAnyWithValue(pubkey)
	suite.Require().Nil(err)

	tests := []struct {
		name         string
		update       *types.MsgUpdateSequencerInformation
		malleate     func(sequencer types.Sequencer) types.Sequencer
		expError     error
		expSequencer types.Sequencer
	}{
		{
			name: "Update sequencer: success",
			update: &types.MsgUpdateSequencerInformation{
				Creator:   addr.String(),
				RollappId: "rollapp_1234-1",
				Metadata: types.SequencerMetadata{
					Moniker:         "Sequencer",
					Identity:        "something",
					SecurityContact: "me@yourplace.xxx",
					Details:         "Details and such",
					P2PSeed:         "seed",
					Rpcs:            []string{"rpc1", "rpc2"},
					EvmRpcs:         []string{"evm1", "evm2"},
					RestApiUrls:     []string{"rest1", "rest2"},
					GenesisUrl:      "genesis",
					ExplorerUrl:     "explorer",
					ContactDetails: &types.ContactDetails{
						Website:  "https://dymension.xyz",
						Telegram: "rolly",
						X:        "rolly",
					},
					ExtraData: []byte("extra"),
					Snapshots: []*types.SnapshotInfo{
						{
							SnapshotUrl: "snapshot",
							Height:      1234,
							Checksum:    "checksum",
						},
					},
					GasPrice: func() *sdk.Int {
						gasPrice := sdk.NewInt(100)
						return &gasPrice
					}(),
				},
			},
			expError: nil,
			expSequencer: types.Sequencer{
				Address:      addr.String(),
				DymintPubKey: pkAny,
				RollappId:    "rollapp_1234-1",
				Metadata: types.SequencerMetadata{
					Moniker:         "Sequencer",
					Identity:        "something",
					SecurityContact: "me@yourplace.xxx",
					Details:         "Details and such",
					P2PSeed:         "seed",
					Rpcs:            []string{"rpc1", "rpc2"},
					EvmRpcs:         []string{"evm1", "evm2"},
					RestApiUrls:     []string{"rest1", "rest2"},
					GenesisUrl:      "genesis",
					ExplorerUrl:     "explorer",
					ContactDetails: &types.ContactDetails{
						Website:  "https://dymension.xyz",
						Telegram: "rolly",
						X:        "rolly",
					},
					ExtraData: []byte("extra"),
					Snapshots: []*types.SnapshotInfo{
						{
							SnapshotUrl: "snapshot",
							Height:      1234,
							Checksum:    "checksum",
						},
					},
					GasPrice: func() *sdk.Int {
						gasPrice := sdk.NewInt(100)
						return &gasPrice
					}(),
				},
				Jailed:          false,
				Proposer:        false,
				Status:          0,
				Tokens:          nil,
				UnbondingHeight: 0,
				UnbondTime:      time.Time{},
			},
		}, {
			name: "Update rollapp: fail - try to update a non-existing rollapp",
			update: &types.MsgUpdateSequencerInformation{
				Creator:   addr.String(),
				RollappId: "somerollapp_1235-1",
			},
			expError: types.ErrUnknownRollappID,
		}, {
			name: "Update rollapp: fail - try to update a frozen rollapp",
			update: &types.MsgUpdateSequencerInformation{
				Creator:   addr.String(),
				RollappId: "rollapp_1235-1",
			},
			malleate: func(r types.Sequencer) types.Sequencer {
				suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapptypes.Rollapp{
					RollappId: "rollapp_1235-1",
					Frozen:    true,
				})
				return r
			},
			expError: types.ErrRollappJailed,
		}, {
			name: "Update rollapp: fail - try to update a jailed sequencer",
			update: &types.MsgUpdateSequencerInformation{
				Creator:   addr.String(),
				RollappId: "rollapp_1234-1",
			},
			malleate: func(r types.Sequencer) types.Sequencer {
				r.Jailed = true
				return r
			},
			expError: types.ErrSequencerJailed,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			goCtx := sdk.WrapSDKContext(suite.Ctx)
			rollapp := rollapptypes.Rollapp{
				RollappId:               "rollapp_1234-1",
				Creator:                 addr.String(),
				InitialSequencerAddress: addr.String(),
			}

			suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

			sequencer := types.Sequencer{
				Address:         addr.String(),
				DymintPubKey:    pkAny,
				RollappId:       "rollapp_1234-1",
				Metadata:        types.SequencerMetadata{},
				Jailed:          false,
				Proposer:        false,
				Status:          0,
				Tokens:          nil,
				UnbondingHeight: 0,
				UnbondTime:      time.Time{},
			}

			if tc.malleate != nil {
				sequencer = tc.malleate(sequencer)
			}

			suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)
			suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

			_, err = suite.msgServer.UpdateSequencerInformation(goCtx, tc.update)
			if tc.expError == nil {
				suite.Require().NoError(err)
				resp, err := suite.queryClient.Sequencer(goCtx, &types.QueryGetSequencerRequest{SequencerAddress: tc.update.Creator})
				suite.Require().NoError(err)
				suite.equalSequencer(&tc.expSequencer, &resp.Sequencer)
			} else {
				suite.ErrorIs(err, tc.expError)
			}
		})
	}
}
