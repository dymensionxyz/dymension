package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateOrderParams(t *testing.T) {
	tests := []struct {
		name            string
		params          []string
		orderType       OrderType
		wantErr         bool
		wantErrContains string
	}{
		{
			name:      "pass - no params for Dym-Name order",
			params:    nil,
			orderType: NameOrder,
			wantErr:   false,
		},
		{
			name:      "pass - no params for Dym-Name order",
			params:    []string{},
			orderType: NameOrder,
			wantErr:   false,
		},
		{
			name:            "fail - reject if has params for Dym-Name order",
			params:          []string{"one"},
			orderType:       NameOrder,
			wantErr:         true,
			wantErrContains: "not accept order params for order type:",
		},
		{
			name:            "fail - reject if has params for Dym-Name order",
			params:          []string{"one", "two", "three"},
			orderType:       NameOrder,
			wantErr:         true,
			wantErrContains: "not accept order params for order type:",
		},
		{
			name:      "pass - one params for Alias order",
			params:    []string{"rollapp_1-1"},
			orderType: AliasOrder,
			wantErr:   false,
		},
		{
			name:            "fail - reject empty params for Alias order",
			params:          nil,
			orderType:       AliasOrder,
			wantErr:         true,
			wantErrContains: "expect 1 order param of RollApp ID for order type:",
		},
		{
			name:            "fail - reject empty params for Alias order",
			params:          []string{},
			orderType:       AliasOrder,
			wantErr:         true,
			wantErrContains: "expect 1 order param of RollApp ID for order type:",
		},
		{
			name:            "fail - reject bad chain-id as params for Alias order",
			params:          []string{"@bad"},
			orderType:       AliasOrder,
			wantErr:         true,
			wantErrContains: "invalid RollApp ID format:",
		},
		{
			name:            "fail - reject bad chain-id as params for Alias order",
			params:          []string{""},
			orderType:       AliasOrder,
			wantErr:         true,
			wantErrContains: "invalid RollApp ID format:",
		},
		{
			name:            "fail - reject unknown order type",
			orderType:       OrderType_OT_UNKNOWN,
			wantErr:         true,
			wantErrContains: "unknown order type:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrderParams(tt.params, tt.orderType)
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
