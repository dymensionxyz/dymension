package types

import (
	"strings"
	"testing"

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
			name: "invalid address",
			msg: MsgUpdateSequencerInformation{
				Creator: "invalid_address",
			},
			err: ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgUpdateSequencerInformation{
				Creator: sample.AccAddress(),
			},
		}, {
			name: "valid metadata",
			msg: MsgUpdateSequencerInformation{
				Creator: sample.AccAddress(),
				Metadata: SequencerMetadata{
					Moniker:         strings.Repeat("a", MaxMonikerLength),
					Identity:        strings.Repeat("a", MaxIdentityLength),
					SecurityContact: strings.Repeat("a", MaxSecurityContactLength),
					Details:         strings.Repeat("a", MaxDetailsLength),
					ContactDetails: &ContactDetails{
						Website:  strings.Repeat("a", MaxContactFieldLength),
						Telegram: strings.Repeat("a", MaxContactFieldLength),
						X:        strings.Repeat("a", MaxContactFieldLength),
					},
					ExtraData: []byte(strings.Repeat("a", MaxExtraDataLength)),
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
			err: ErrInvalidRequest,
		}, {
			name: "invalid identity length",
			msg: MsgUpdateSequencerInformation{
				Creator: sample.AccAddress(),
				Metadata: SequencerMetadata{
					Identity: strings.Repeat("a", MaxIdentityLength+1),
				},
			},
			err: ErrInvalidRequest,
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
			err: ErrInvalidRequest,
		}, {
			name: "invalid security contact length",
			msg: MsgUpdateSequencerInformation{
				Creator: sample.AccAddress(),
				Metadata: SequencerMetadata{
					SecurityContact: strings.Repeat("a", MaxSecurityContactLength+1),
				},
			},
			err: ErrInvalidRequest,
		}, {
			name: "invalid details length",
			msg: MsgUpdateSequencerInformation{
				Creator: sample.AccAddress(),
				Metadata: SequencerMetadata{
					Details: strings.Repeat("a", MaxDetailsLength+1),
				},
			},
			err: ErrInvalidRequest,
		}, {
			name: "invalid extra data length",
			msg: MsgUpdateSequencerInformation{
				Creator: sample.AccAddress(),
				Metadata: SequencerMetadata{
					ExtraData: []byte(strings.Repeat("a", MaxExtraDataLength+1)),
				},
			},
			err: ErrInvalidRequest,
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
