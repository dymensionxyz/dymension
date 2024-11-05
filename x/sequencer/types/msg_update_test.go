package types

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

func TestMsgUpdateSequencerInformation_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateSequencerInformation
		err  error
	}{
		{
			name: "valid metadata",
			msg: MsgUpdateSequencerInformation{
				Creator: sample.AccAddress(),
				Metadata: SequencerMetadata{
					Moniker:     strings.Repeat("a", MaxMonikerLength),
					Details:     strings.Repeat("a", MaxDetailsLength),
					P2PSeeds:    []string{"seed1", "seed2"},
					Rpcs:        []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
					EvmRpcs:     []string{"https://rpc.evm.rollapp.noisnemyd.xyz:443"},
					RestApiUrls: []string{"http://localhost:1317"},
					GenesisUrls: []string{"genesis1", "genesis2"},
					ExplorerUrl: "explorer",
					ContactDetails: &ContactDetails{
						Website:  "https://website.com",
						Telegram: "https://t.me/rolly",
						X:        "https://x.dymension.xyz",
					},
					ExtraData: []byte(strings.Repeat("a", MaxExtraDataLength)),
					Snapshots: []*SnapshotInfo{
						{
							SnapshotUrl: "snapshot",
							Height:      1234,
							Checksum:    "checksum",
						},
					},
					GasPrice: uptr.To(sdk.NewInt(100)),
				},
			},
		}, {
			name: "invalid moniker length",
			msg: MsgUpdateSequencerInformation{
				Creator: sample.AccAddress(),
				Metadata: SequencerMetadata{
					Moniker: strings.Repeat("a", MaxMonikerLength+1),
				},
			},
			err: ErrInvalidMetadata,
		}, {
			name: "invalid website length",
			msg: MsgUpdateSequencerInformation{
				Creator: sample.AccAddress(),
				Metadata: SequencerMetadata{
					ContactDetails: &ContactDetails{
						Website: strings.Repeat("a", MaxContactFieldLength+1),
					},
				},
			},
			err: ErrInvalidMetadata,
		}, {
			name: "invalid details length",
			msg: MsgUpdateSequencerInformation{
				Creator: sample.AccAddress(),
				Metadata: SequencerMetadata{
					Details: strings.Repeat("a", MaxDetailsLength+1),
				},
			},
			err: ErrInvalidMetadata,
		}, {
			name: "invalid extra data length",
			msg: MsgUpdateSequencerInformation{
				Creator: sample.AccAddress(),
				Metadata: SequencerMetadata{
					ExtraData: []byte(strings.Repeat("a", MaxExtraDataLength+1)),
				},
			},
			err: ErrInvalidMetadata,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
