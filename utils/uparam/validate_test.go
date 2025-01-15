package uparam

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestValidateUint64(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		x := uint64(42)
		require.NoError(t, ValidateUint64(x))
	})
	t.Run("invalid", func(t *testing.T) {
		x := "foo"
		require.Error(t, ValidateUint64(x))
	})
}

func TestValidatePositiveUint64(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		x := uint64(42)
		require.NoError(t, ValidatePositiveUint64(x))
	})
	t.Run("invalid", func(t *testing.T) {
		x := uint64(0)
		require.Error(t, ValidatePositiveUint64(x))
	})
}

func TestValidateDec(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		x := sdk.Dec{}
		require.NoError(t, ValidateDec(x))
	})
	t.Run("invalid", func(t *testing.T) {
		x := "foo"
		require.Error(t, ValidateDec(x))
	})
}

func TestValidateNonNegativeDec(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		x := sdk.OneDec()
		require.NoError(t, ValidateNonNegativeDec(x))
	})
	t.Run("invalid - nil", func(t *testing.T) {
		x := sdk.ZeroDec().Sub(sdk.OneDec())
		require.Error(t, ValidateNonNegativeDec(x))
	})
	t.Run("invalid - negative", func(t *testing.T) {
		x := sdk.Dec{}
		require.Error(t, ValidateNonNegativeDec(x))
	})
}

func TestValidateZeroToOneDec(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		x := sdk.MustNewDecFromStr("0.5")
		require.NoError(t, ValidateZeroToOneDec(x))
	})
	t.Run("invalid - large number", func(t *testing.T) {
		x := sdk.MustNewDecFromStr("100")
		require.Error(t, ValidateZeroToOneDec(x))
	})
}
