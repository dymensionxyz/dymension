package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"
)

func TestRegisterCodec(t *testing.T) {
	t.Run("can register codec", func(t *testing.T) {
		require.NotPanics(t, func() {
			RegisterCodec(codec.NewLegacyAmino())
		})
	})
}

func TestRegisterInterfaces(t *testing.T) {
	t.Run("can register interfaces", func(t *testing.T) {
		require.NotPanics(t, func() {
			RegisterInterfaces(cdctypes.NewInterfaceRegistry())
		})
	})
}
