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
