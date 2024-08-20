package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateOrderParams(t *testing.T) {
	tests := []struct {
		name            string
		params          []string
		assetType       AssetType
		wantErr         bool
		wantErrContains string
	}{
		{
			name:      "pass - no params for Dym-Name order",
			params:    nil,
			assetType: TypeName,
			wantErr:   false,
		},
		{
			name:      "pass - no params for Dym-Name order",
			params:    []string{},
			assetType: TypeName,
			wantErr:   false,
		},
		{
			name:            "fail - reject if has params for Dym-Name order",
			params:          []string{"one"},
			assetType:       TypeName,
			wantErr:         true,
			wantErrContains: "not accept order params for asset type:",
		},
		{
			name:            "fail - reject if has params for Dym-Name order",
			params:          []string{"one", "two", "three"},
			assetType:       TypeName,
			wantErr:         true,
			wantErrContains: "not accept order params for asset type:",
		},
		{
			name:      "pass - one params for Alias order",
			params:    []string{"rollapp_1-1"},
			assetType: TypeAlias,
			wantErr:   false,
		},
		{
			name:            "fail - reject empty params for Alias order",
			params:          nil,
			assetType:       TypeAlias,
			wantErr:         true,
			wantErrContains: "expect 1 order param of RollApp ID for asset type:",
		},
		{
			name:            "fail - reject empty params for Alias order",
			params:          []string{},
			assetType:       TypeAlias,
			wantErr:         true,
			wantErrContains: "expect 1 order param of RollApp ID for asset type:",
		},
		{
			name:            "fail - reject bad chain-id as params for Alias order",
			params:          []string{"@bad"},
			assetType:       TypeAlias,
			wantErr:         true,
			wantErrContains: "invalid RollApp ID format:",
		},
		{
			name:            "fail - reject bad chain-id as params for Alias order",
			params:          []string{""},
			assetType:       TypeAlias,
			wantErr:         true,
			wantErrContains: "invalid RollApp ID format:",
		},
		{
			name:            "fail - reject unknown asset type",
			assetType:       AssetType_AT_UNKNOWN,
			wantErr:         true,
			wantErrContains: "unknown asset type:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrderParams(tt.params, tt.assetType)
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}
}
