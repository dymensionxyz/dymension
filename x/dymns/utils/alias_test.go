package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidAlias(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		wantValid bool
	}{
		{
			name:      "pass - valid 1 char",
			alias:     "a",
			wantValid: true,
		},
		{
			name:      "pass - valid 2 chars",
			alias:     "aa",
			wantValid: true,
		},
		{
			name:      "pass - valid 10 chars",
			alias:     "1234567890",
			wantValid: true,
		},
		{
			name:      "pass - valid 32 chars",
			alias:     "12345678901234567890123456789012",
			wantValid: true,
		},
		{
			name:      "fail - not accept 33+ chars",
			alias:     "123456789012345678901234567890123",
			wantValid: false,
		},
		{
			name:      "fail - not accept special chars",
			alias:     "a$a",
			wantValid: false,
		},
		{
			name:      "fail - not accept underscore",
			alias:     "a_a",
			wantValid: false,
		},
		{
			name:      "fail - not accept dash",
			alias:     "a-a",
			wantValid: false,
		},
		{
			name:      "fail - not accept empty",
			alias:     "",
			wantValid: false,
		},
		{
			name:      "fail - not accept leading space",
			alias:     " a",
			wantValid: false,
		},
		{
			name:      "fail - not accept trailing space",
			alias:     "a ",
			wantValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantValid, IsValidAlias(tt.alias))
		})
	}
}
