package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidChainIdFormat(t *testing.T) {
	tests := []struct {
		chainId string
		invalid bool
	}{
		{
			chainId: "cosmoshub-4",
		},
		{
			chainId: "dymension_1100-1",
		},
		{
			chainId: "ethermint_9000-1",
		},
		{
			chainId: "axelar-dojo-1",
		},
		{
			chainId: "abc",
		},
		{
			chainId: "ab",
			invalid: true,
		},
		{
			chainId: "ab99",
			invalid: true,
		},
		{
			chainId: "abc-1",
		},
		{
			chainId: "abc-",
			invalid: true,
		},
		{
			chainId: "-abc",
			invalid: true,
		},
		{
			chainId: "abc_",
			invalid: true,
		},
		{
			chainId: "_abc",
			invalid: true,
		},
		{
			chainId: "500_abc",
			invalid: true,
		},
		{
			chainId: "4-cosmoshub",
			invalid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.chainId, func(t *testing.T) {
			valid := IsValidChainIdFormat(tt.chainId)
			if tt.invalid {
				require.Falsef(t, valid, "expected invalid chain id: %s", tt.chainId)
			} else {
				require.Truef(t, valid, "expected valid chain id: %s", tt.chainId)
			}
		})
	}
}

func TestIsValidEIP155ChainId(t *testing.T) {
	tests := []struct {
		name          string
		eip155ChainId string
		want          bool
	}{
		{
			name:          "pass - valid, single digit",
			eip155ChainId: "1",
			want:          true,
		},
		{
			name:          "pass - valid, single digit",
			eip155ChainId: "9",
			want:          true,
		},
		{
			name:          "fail - negative number",
			eip155ChainId: "-9",
			want:          false,
		},
		{
			name:          "fail - zero is not allowed as it can be output from failed-to-parse cases",
			eip155ChainId: "0",
			want:          false,
		},
		{
			name:          "fail - leading zero is not allowed",
			eip155ChainId: "09",
			want:          false,
		},
		{
			name:          "fail - leading zero is not allowed",
			eip155ChainId: "000",
			want:          false,
		},
		{
			name:          "pass - valid, multiple digits",
			eip155ChainId: "99",
			want:          true,
		},
		{
			name:          "fail - alphabet",
			eip155ChainId: "a",
			want:          false,
		},
		{
			name:          "fail - alphabet",
			eip155ChainId: "crypto",
			want:          false,
		},
		{
			name:          "fail - alphanumeric",
			eip155ChainId: "crypto9",
			want:          false,
		},
		{
			name:          "fail - alphanumeric",
			eip155ChainId: "9crypto",
			want:          false,
		},
		{
			name:          "fail - Cosmos chain id",
			eip155ChainId: "cosmoshub-4",
			want:          false,
		},
		{
			name:          "fail - Cosmos chain id",
			eip155ChainId: "dymension_100-1",
			want:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsValidEIP155ChainId(tt.eip155ChainId))
		})
	}
}

func TestMustGetEIP155ChainIdFromRollAppId(t *testing.T) {
	tests := []struct {
		name           string
		rollAppChainId string
		want           string
		wantPanic      bool
	}{
		{
			name:           "pass",
			rollAppChainId: "rollapp_1-1",
			want:           "1",
		},
		{
			name:           "pass",
			rollAppChainId: "rollapp_9999-1",
			want:           "9999",
		},
		{
			name:           "fail - panic when invalid format",
			rollAppChainId: "rollapp_-1",
			wantPanic:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				require.Panics(t, func() {
					_ = MustGetEIP155ChainIdFromRollAppId(tt.rollAppChainId)
				})
				return
			}
			require.Equal(t, tt.want, MustGetEIP155ChainIdFromRollAppId(tt.rollAppChainId))
		})
	}
}
