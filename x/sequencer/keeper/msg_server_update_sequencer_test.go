package keeper_test

import (
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"

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
					Moniker:     "Sequencer",
					Details:     "Details and such",
					P2PSeeds:    []string{"seed1", "seed2"},
					Rpcs:        []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
					EvmRpcs:     []string{"https://rpc.evm.rollapp.noisnemyd.xyz:443"},
					RestApiUrls: []string{"http://localhost:1317"},
					GenesisUrls: []string{"genesis1", "genesis2"},
					ExplorerUrl: "explorer",
					ContactDetails: &types.ContactDetails{
						Website:  "https://dymension.xyz",
						Telegram: "https://t.me/rolly",
						X:        "https://x.dymension.xyz",
					},
					ExtraData: []byte("extra"),
					Snapshots: []*types.SnapshotInfo{
						{
							SnapshotUrl: "snapshot",
							Height:      1234,
							Checksum:    "checksum",
						},
					},
					GasPrice: uptr.To(sdk.NewInt(100)),
				},
			},
			expError: nil,
			expSequencer: types.Sequencer{
				Address:      addr.String(),
				DymintPubKey: pkAny,
				RollappId:    "rollapp_1234-1",
				Metadata: types.SequencerMetadata{
					Moniker:     "Sequencer",
					Details:     "Details and such",
					P2PSeeds:    []string{"seed1", "seed2"},
					Rpcs:        []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
					EvmRpcs:     []string{"https://rpc.evm.rollapp.noisnemyd.xyz:443"},
					RestApiUrls: []string{"http://localhost:1317"},
					GenesisUrls: []string{"genesis1", "genesis2"},
					ExplorerUrl: "explorer",
					ContactDetails: &types.ContactDetails{
						Website:  "https://dymension.xyz",
						Telegram: "https://t.me/rolly",
						X:        "https://x.dymension.xyz",
					},
					ExtraData: []byte("extra"),
					Snapshots: []*types.SnapshotInfo{
						{
							SnapshotUrl: "snapshot",
							Height:      1234,
							Checksum:    "checksum",
						},
					},
					GasPrice: uptr.To(sdk.NewInt(100)),
				},
				Jailed:     false,
				Status:     0,
				Tokens:     nil,
				UnbondTime: time.Time{},
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
				Metadata: types.SequencerMetadata{
					Rpcs:    []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
					EvmRpcs: []string{"https://rpc.evm.rollapp.noisnemyd.xyz:443"},
				},
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
			goCtx := sdk.WrapSDKContext(suite.Ctx)
			rollapp := rollapptypes.Rollapp{
				RollappId:        "rollapp_1234-1",
				Owner:            addr.String(),
				InitialSequencer: addr.String(),
			}

			suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

			sequencer := types.Sequencer{
				Address:      addr.String(),
				DymintPubKey: pkAny,
				RollappId:    "rollapp_1234-1",
				Metadata:     types.SequencerMetadata{},
				Jailed:       false,
				Status:       0,
				Tokens:       nil,
				UnbondTime:   time.Time{},
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
