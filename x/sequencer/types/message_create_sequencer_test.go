package types

import (
	"strings"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateSequencer_ValidateBasic(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address()).String()
	pkAny, err := codectypes.NewAnyWithValue(pubkey)
	require.NoError(t, err)

	tests := []struct {
		name string
		msg  MsgCreateSequencer
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCreateSequencer{
				Creator:          "invalid_address",
				SequencerAddress: addr,
				Pubkey:           pkAny,
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCreateSequencer{
				Creator:          sample.AccAddress(),
				SequencerAddress: addr,
				Pubkey:           pkAny,
			},
		}, {
			name: "invalid sequencer address",
			msg: MsgCreateSequencer{
				Creator:          sample.AccAddress(),
				SequencerAddress: "invalid_address",
				Pubkey:           pkAny,
			},
			err: ErrInvalidSequencerAddress,
		}, {
			name: "empty pubkey",
			msg: MsgCreateSequencer{
				Creator:          sample.AccAddress(),
				SequencerAddress: addr,
			},
			err: sdkerrors.ErrInvalidPubKey,
		}, {
			name: "invalid pubkey",
			msg: MsgCreateSequencer{
				Creator:          sample.AccAddress(),
				SequencerAddress: addr,
			},
			err: sdkerrors.ErrInvalidPubKey,
		}, {
			name: "valid description",
			msg: MsgCreateSequencer{
				Creator:          sample.AccAddress(),
				SequencerAddress: addr,
				Pubkey:           pkAny,
				Description: Description{
					Moniker:         strings.Repeat("a", MaxMonikerLength),
					Identity:        strings.Repeat("a", MaxIdentityLength),
					Website:         strings.Repeat("a", MaxWebsiteLength),
					SecurityContact: strings.Repeat("a", MaxSecurityContactLength),
					Details:         strings.Repeat("a", MaxDetailsLength)},
			},
		}, {
			name: "invalid moniker length",
			msg: MsgCreateSequencer{
				Creator:          sample.AccAddress(),
				SequencerAddress: addr,
				Pubkey:           pkAny,
				Description: Description{
					Moniker: strings.Repeat("a", MaxMonikerLength+1)},
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "invalid identity length",
			msg: MsgCreateSequencer{
				Creator:          sample.AccAddress(),
				SequencerAddress: addr,
				Pubkey:           pkAny,
				Description: Description{
					Identity: strings.Repeat("a", MaxIdentityLength+1)},
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "invalid website length",
			msg: MsgCreateSequencer{
				Creator:          sample.AccAddress(),
				SequencerAddress: addr,
				Pubkey:           pkAny,
				Description: Description{
					Website: strings.Repeat("a", MaxWebsiteLength+1)},
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "invalid security contact length",
			msg: MsgCreateSequencer{
				Creator:          sample.AccAddress(),
				SequencerAddress: addr,
				Pubkey:           pkAny,
				Description: Description{
					SecurityContact: strings.Repeat("a", MaxSecurityContactLength+1)},
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "invalid details length",
			msg: MsgCreateSequencer{
				Creator:          sample.AccAddress(),
				SequencerAddress: addr,
				Pubkey:           pkAny,
				Description: Description{
					Details: strings.Repeat("a", MaxDetailsLength+1)},
			},
			err: sdkerrors.ErrInvalidRequest,
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
