package denom_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/utils/denom"
)

func TestSourcePortChanFromTracePath(t *testing.T) {
	testCases := []struct {
		name     string
		trace    string
		expValid bool
		expPort  string
		expChan  string
	}{
		{"invalid: empty trace", "", false, "", ""},
		{"invalid: only port", "transfer", false, "", ""},
		{"invalid: only port with '/'", "transfer/", false, "", ""},
		{"invalid: only channel with '/'", "/channel-1", false, "", ""},
		{"invalid: only '/'", "/", false, "", ""},
		{"invalid: double '/'", "transfer//channel-1", false, "", ""},
		{"valid trace", "transfer/channel-1", true, "transfer", "channel-1"},
		{"valid trace with multiple port/channel pairs", "transfer/channel-1/transfer/channel-2", true, "transfer", "channel-2"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			port, channel, valid := denom.SourcePortChanFromTracePath(tc.trace)

			require.Equal(t, tc.expValid, valid)
			if tc.expValid {
				require.Equal(t, tc.expPort, port)
				require.Equal(t, tc.expChan, channel)
			}
		})
	}
}
