package types

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

func TestMsgCreateSequencer_ValidateBasic(t *testing.T) {
	pubkey := ed25519.GenPrivKey().PubKey()
	pkAny, err := codectypes.NewAnyWithValue(pubkey)
	require.NoError(t, err)

	invalidkey := "{\"@type\":\"/cosmos.crypto.ed25519.PubKey\",\"key\":\"OcEwSZhPfddSUr84dkfj6Sfsh6PDSkcBdySUFxPb0Fs=\"}"
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	codec := codec.NewProtoCodec(interfaceRegistry)
	var invalidpk cryptotypes.PubKey
	err = codec.UnmarshalInterfaceJSON([]byte(invalidkey), &invalidpk)
	require.NoError(t, err)
	pkInvalid, err := codectypes.NewAnyWithValue(invalidpk)
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
			err: ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         bond,
				Metadata: SequencerMetadata{
					Rpcs:        []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443", "https://rpc.wpd.wasm.rollapp.noisnemyd.xyz:443"},
					EvmRpcs:     []string{"https://rpc.wpd.evm.evm.noisnemyd.xyz:443"},
					RestApiUrls: []string{"https://api.wpd.evm.rollapp.noisnemyd.xyz:443"},
				},
			},
		}, {
			name: "valid metadata",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         bond,
				Metadata: SequencerMetadata{
					Moniker:     strings.Repeat("a", MaxMonikerLength),
					Details:     strings.Repeat("a", MaxDetailsLength),
					P2PSeeds:    []string{"seed1", "seed2"},
					Rpcs:        []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443", "https://rpc.wpd.wasm.rollapp.noisnemyd.xyz:443"},
					EvmRpcs:     []string{"https://rpc.wpd.evm.evm.noisnemyd.xyz:443"},
					RestApiUrls: []string{"https://api.wpd.evm.rollapp.noisnemyd.xyz:443"},
					ExplorerUrl: "explorer_url",
					GenesisUrls: []string{"genesis1", "genesis2"},
					ContactDetails: &ContactDetails{
						Website:  "https://website.com",
						Telegram: "https://t.me/telegram",
						X:        "https://x.com",
					},
					ExtraData: []byte(strings.Repeat("a", MaxExtraDataLength)),
					Snapshots: []*SnapshotInfo{
						{
							SnapshotUrl: "snapshot_url",
							Height:      123,
							Checksum:    "checksum",
						},
					},
					GasPrice: uptr.To(sdk.NewInt(100)),
				},
			},
		}, {
			name: "invalid moniker length",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         bond,
				Metadata: SequencerMetadata{
					Moniker:     strings.Repeat("a", MaxMonikerLength+1),
					Rpcs:        []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443", "https://rpc.wpd.wasm.rollapp.noisnemyd.xyz:443"},
					EvmRpcs:     []string{"https://rpc.wpd.evm.evm.noisnemyd.xyz:443"},
					RestApiUrls: []string{"https://api.wpd.evm.rollapp.noisnemyd.xyz:443"},
				},
			},
			err: ErrInvalidMetadata,
		}, {
			name: "invalid website length",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         bond,
				Metadata: SequencerMetadata{
					ContactDetails: &ContactDetails{
						Website: strings.Repeat("a", MaxContactFieldLength+1),
					},
					Rpcs:        []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443", "https://rpc.wpd.wasm.rollapp.noisnemyd.xyz:443"},
					EvmRpcs:     []string{"https://rpc.wpd.evm.evm.noisnemyd.xyz:443"},
					RestApiUrls: []string{"https://api.wpd.evm.rollapp.noisnemyd.xyz:443"},
				},
			},
			err: ErrInvalidMetadata,
		}, {
			name: "invalid details length",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         bond,
				Metadata: SequencerMetadata{
					Details:     strings.Repeat("a", MaxDetailsLength+1),
					Rpcs:        []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443", "https://rpc.wpd.wasm.rollapp.noisnemyd.xyz:443"},
					EvmRpcs:     []string{"https://rpc.wpd.evm.evm.noisnemyd.xyz:443"},
					RestApiUrls: []string{"https://api.wpd.evm.rollapp.noisnemyd.xyz:443"},
				},
			},
			err: ErrInvalidMetadata,
		}, {
			name: "invalid extra data length",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         bond,
				Metadata: SequencerMetadata{
					ExtraData:   []byte(strings.Repeat("a", MaxExtraDataLength+1)),
					Rpcs:        []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443", "https://rpc.wpd.wasm.rollapp.noisnemyd.xyz:443"},
					EvmRpcs:     []string{"https://rpc.wpd.evm.evm.noisnemyd.xyz:443"},
					RestApiUrls: []string{"https://api.wpd.evm.rollapp.noisnemyd.xyz:443"},
				},
			},
			err: ErrInvalidMetadata,
		}, {
			name: "invalid bond",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkAny,
				Bond:         sdk.Coin{Denom: "k", Amount: sdk.NewInt(0)},
				Metadata: SequencerMetadata{
					Rpcs:        []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443", "https://rpc.wpd.wasm.rollapp.noisnemyd.xyz:443"},
					EvmRpcs:     []string{"https://rpc.wpd.evm.evm.noisnemyd.xyz:443"},
					RestApiUrls: []string{"https://api.wpd.evm.rollapp.noisnemyd.xyz:443"},
				},
			},
			err: ErrInvalidCoins,
		}, {
			name: "invalid public key",
			msg: MsgCreateSequencer{
				Creator:      sample.AccAddress(),
				DymintPubKey: pkInvalid,
				Bond:         bond,
			},
			err: ErrInvalidPubKey,
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

func TestNewMsgIncreaseBond_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgIncreaseBond
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgIncreaseBond{
				Creator:   "invalid_address",
				AddAmount: sdk.NewInt64Coin("stake", 100),
			},
			err: ErrInvalidAddress,
		},
		{
			name: "invalid bond amount",
			msg: MsgIncreaseBond{
				Creator:   sample.AccAddress(),
				AddAmount: sdk.NewInt64Coin("stake", 0),
			},
			err: ErrInvalidCoins,
		},
		{
			name: "valid",
			msg: MsgIncreaseBond{
				Creator:   sample.AccAddress(),
				AddAmount: sdk.NewInt64Coin("stake", 100),
			},
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

func TestNewMsgDecreaseBond_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDecreaseBond
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgDecreaseBond{
				Creator:        "invalid_address",
				DecreaseAmount: sdk.NewInt64Coin("stake", 100),
			},
			err: ErrInvalidAddress,
		},
		{
			name: "invalid bond amount",
			msg: MsgDecreaseBond{
				Creator:        sample.AccAddress(),
				DecreaseAmount: sdk.NewInt64Coin("stake", 0),
			},
			err: ErrInvalidCoins,
		},
		{
			name: "valid",
			msg: MsgDecreaseBond{
				Creator:        sample.AccAddress(),
				DecreaseAmount: sdk.NewInt64Coin("stake", 100),
			},
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
