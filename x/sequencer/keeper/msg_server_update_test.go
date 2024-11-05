package keeper_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (s *SequencerTestSuite) TestUpdateSequencer() {
	pubkey := ed25519.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	pkAny, err := codectypes.NewAnyWithValue(pubkey)
	s.Require().Nil(err)

	const rollappID = "rollapp_1234-1"

	tests := []struct {
		name         string
		update       *types.MsgUpdateSequencerInformation
		malleate     func(sequencer *types.Sequencer)
		expError     error
		expSequencer types.Sequencer
	}{
		{
			name: "Update sequencer: success",
			update: &types.MsgUpdateSequencerInformation{
				Creator: addr.String(),
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
				RollappId:    rollappID,
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
				Status: 0,
				Tokens: nil,
			},
		}, {
			name: "Update rollapp: fail - try to update wrong VM type fields",
			update: &types.MsgUpdateSequencerInformation{
				Creator: addr.String(),
				Metadata: types.SequencerMetadata{
					EvmRpcs: []string{"https://rpc.evm.rollapp.noisnemyd.xyz:443"},
				},
			},
			malleate: func(*types.Sequencer) {
				s.raK().SetRollapp(s.Ctx, rollapptypes.Rollapp{
					RollappId: rollappID,
					VmType:    rollapptypes.Rollapp_WASM,
				})
			},
			expError: types.ErrInvalidVMTypeUpdate,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			goCtx := sdk.WrapSDKContext(s.Ctx)
			rollapp := rollapptypes.Rollapp{
				RollappId:        rollappID,
				VmType:           rollapptypes.Rollapp_EVM,
				Owner:            addr.String(),
				InitialSequencer: addr.String(),
			}

			s.raK().SetRollapp(s.Ctx, rollapp)

			sequencer := types.Sequencer{
				Address:      addr.String(),
				DymintPubKey: pkAny,
				RollappId:    rollappID,
				Metadata:     types.SequencerMetadata{},
				Status:       0,
				Tokens:       nil,
			}

			s.raK().SetRollapp(s.Ctx, rollapp)
			s.k().SetSequencer(s.Ctx, sequencer)

			if tc.malleate != nil {
				tc.malleate(&sequencer)
			}

			_, err = s.msgServer.UpdateSequencerInformation(goCtx, tc.update)
			if tc.expError == nil {
				s.Require().NoError(err)
				resp, err := s.queryClient.Sequencer(goCtx, &types.QueryGetSequencerRequest{SequencerAddress: tc.update.Creator})
				s.Require().NoError(err)
				s.equalSequencers(&tc.expSequencer, &resp.Sequencer)
			} else {
				s.ErrorIs(err, tc.expError)
			}
		})
	}
}

func (s *SequencerTestSuite) TestChangeOptInStatusBasicFlow() {
	ra := s.createRollapp()
	seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)

	m := &types.MsgUpdateOptInStatus{Creator: seq.Address, OptedIn: false}
	_, err := s.msgServer.UpdateOptInStatus(s.Ctx, m)
	s.Require().NoError(err)
	seq = s.k().GetSequencer(s.Ctx, seq.Address)
	s.Require().False(seq.OptedIn)

	m = &types.MsgUpdateOptInStatus{Creator: seq.Address, OptedIn: true}
	_, err = s.msgServer.UpdateOptInStatus(s.Ctx, m)
	s.Require().NoError(err)
	seq = s.k().GetSequencer(s.Ctx, seq.Address)
	s.Require().True(seq.OptedIn)
}
