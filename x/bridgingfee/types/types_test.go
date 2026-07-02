package types

import (
	"testing"

	"cosmossdk.io/math"
	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/stretchr/testify/require"
)

func TestHLAssetFee_Validate_OutboundBounds(t *testing.T) {
	tokenId := hyputil.CreateMockHexAddress("test", 1)
	base := func() HLAssetFee {
		return HLAssetFee{
			TokenId:     tokenId,
			InboundFee:  math.LegacyMustNewDecFromStr("0.01"),
			OutboundFee: math.LegacyMustNewDecFromStr("0.02"),
		}
	}

	tests := []struct {
		name    string
		mutate  func(*HLAssetFee)
		wantErr string
	}{
		{
			name:   "nil bounds (pre-upgrade hooks) are valid",
			mutate: func(f *HLAssetFee) {},
		},
		{
			name:   "min set, max unbounded (zero)",
			mutate: func(f *HLAssetFee) { f.MinOutboundFee = math.NewInt(1000); f.MaxOutboundFee = math.ZeroInt() },
		},
		{
			name:   "min == max == 0",
			mutate: func(f *HLAssetFee) { f.MinOutboundFee = math.ZeroInt(); f.MaxOutboundFee = math.ZeroInt() },
		},
		{
			name:   "min < max",
			mutate: func(f *HLAssetFee) { f.MinOutboundFee = math.NewInt(1000); f.MaxOutboundFee = math.NewInt(2000) },
		},
		{
			name:    "negative min",
			mutate:  func(f *HLAssetFee) { f.MinOutboundFee = math.NewInt(-1) },
			wantErr: "min outbound fee cannot be negative",
		},
		{
			name:    "negative max",
			mutate:  func(f *HLAssetFee) { f.MaxOutboundFee = math.NewInt(-1) },
			wantErr: "max outbound fee cannot be negative",
		},
		{
			name:    "0 < max < min",
			mutate:  func(f *HLAssetFee) { f.MinOutboundFee = math.NewInt(2000); f.MaxOutboundFee = math.NewInt(1000) },
			wantErr: "max outbound fee must be >= min outbound fee",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := base()
			tt.mutate(&f)
			err := f.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
