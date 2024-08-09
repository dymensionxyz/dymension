package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	require.Equal(
		t,
		time.Now().Unix(), time.Now().UTC().Unix(),
		"if mis-match, 100% sure will causes AppHash",
	)
}
