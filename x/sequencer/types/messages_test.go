package types

import (
	"strings"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateSequencer_ValidateBasic(t *testing.T) {
	pubkey := ed25519.GenPrivKey().PubKey()
	pkAny, err := codectypes.NewAnyWithValue(pubkey)
	require.NoError(t, err)

	bond := sdk.NewCoin("stake", sdk.NewInt(100))

	tests := []struct {
		name string
		msg  MsgCreateSequencer
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCreateSequencer{
				Creator:      "invalid_address",
				DymintPubKey: pkAny,
				Bond:         bond,
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         bond,
			},
		}, {
			name: "valid description",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         bond,
				Description: Description{
					Moniker:         strings.Repeat("a", MaxMonikerLength),
					Identity:        strings.Repeat("a", MaxIdentityLength),
					Website:         strings.Repeat("a", MaxWebsiteLength),
					SecurityContact: strings.Repeat("a", MaxSecurityContactLength),
					Details:         strings.Repeat("a", MaxDetailsLength),
				},
			},
		}, {
			name: "invalid moniker length",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         bond,
				Description: Description{
					Moniker: strings.Repeat("a", MaxMonikerLength+1),
				},
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "invalid identity length",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         bond,
				Description: Description{
					Identity: strings.Repeat("a", MaxIdentityLength+1),
				},
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "invalid website length",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         bond,
				Description: Description{
					Website: strings.Repeat("a", MaxWebsiteLength+1),
				},
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "invalid security contact length",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         bond,
				Description: Description{
					SecurityContact: strings.Repeat("a", MaxSecurityContactLength+1),
				},
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "invalid details length",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         bond,
				Description: Description{
					Details: strings.Repeat("a", MaxDetailsLength+1),
				},
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "invalid bond",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         sdk.Coin{Denom: "k", Amount: sdk.NewInt(0)},
			},
			err: sdkerrors.ErrInvalidCoins,
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
