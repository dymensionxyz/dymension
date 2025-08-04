package types_test

import (
	"testing"

	"cosmossdk.io/math"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
	"github.com/stretchr/testify/require"
)

func TestMakeHLForwardToHLMetadata(t *testing.T) {
	tokenId, _ := hyperutil.DecodeHexAddress("0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0")
	destinationDomain := uint32(42161) // Arbitrum domain
	recipient, _ := hyperutil.DecodeHexAddress("0x1234567890abcdef1234567890abcdef12345678")
	amount := math.NewInt(1000)
	maxFee := sdk.NewCoin("adym", math.NewInt(10))
	gasLimit := math.NewInt(100000)
	customHookId, _ := hyperutil.DecodeHexAddress("0xabcdef0123456789abcdef0123456789abcdef01")
	customHookMetadata := "test-metadata"

	hook := types.NewHookForwardToHL(
		tokenId,
		destinationDomain,
		recipient,
		amount,
		maxFee,
		gasLimit,
		&customHookId,
		customHookMetadata,
	)

	metadataBz, err := types.MakeHLForwardToHLMetadata(hook)
	require.NoError(t, err)
	require.NotEmpty(t, metadataBz)

	// Test unpacking
	metadata, err := types.UnpackHLMetadata(metadataBz)
	require.NoError(t, err)
	require.NotNil(t, metadata)
	require.NotEmpty(t, metadata.HookForwardToHl)
	require.Empty(t, metadata.HookForwardToIbc)

	// Test unpacking the forward data
	forwardData, err := types.UnpackForwardToHL(metadata.HookForwardToHl)
	require.NoError(t, err)
	require.NotNil(t, forwardData)
	require.Equal(t, tokenId, forwardData.HyperlaneTransfer.TokenId)
	require.Equal(t, destinationDomain, forwardData.HyperlaneTransfer.DestinationDomain)
	require.Equal(t, recipient, forwardData.HyperlaneTransfer.Recipient)
	require.Equal(t, amount, forwardData.HyperlaneTransfer.Amount)
	require.Equal(t, maxFee, forwardData.HyperlaneTransfer.MaxFee)
	require.Equal(t, gasLimit, forwardData.HyperlaneTransfer.GasLimit)
	require.Equal(t, &customHookId, forwardData.HyperlaneTransfer.CustomHookId)
	require.Equal(t, customHookMetadata, forwardData.HyperlaneTransfer.CustomHookMetadata)
}

func TestUnpackForwardToHL(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() []byte
		wantErr   bool
	}{
		{
			name: "valid data",
			setupFunc: func() []byte {
				tokenId, _ := hyperutil.DecodeHexAddress("0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0")
				hook := types.NewHookForwardToHL(
					tokenId,
					uint32(1),
					tokenId,
					math.NewInt(100),
					sdk.NewCoin("adym", math.NewInt(10)),
					math.ZeroInt(),
					nil,
					"",
				)
				bz, _ := hook.Marshal()
				return bz
			},
			wantErr: false,
		},
		{
			name: "invalid data",
			setupFunc: func() []byte {
				return []byte("invalid proto data")
			},
			wantErr: true,
		},
		{
			name: "empty hyperlane transfer",
			setupFunc: func() []byte {
				hook := &types.HookForwardToHL{}
				bz, _ := hook.Marshal()
				return bz
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := tt.setupFunc()
			result, err := types.UnpackForwardToHL(bz)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}
