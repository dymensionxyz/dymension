package utilsmemo

import (
	"strings"
	"testing"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	t.Run("no conflict", func(t *testing.T) {
		memo := `{
"foo" : 0,
"bar" : 1
}`

		res, err := Merge(memo, map[string]int{"fizz": 2, "buzz": 3})
		require.NoError(t, err)
		require.True(t, strings.Contains(res, "foo"))  // nothing is lost
		require.True(t, strings.Contains(res, "fizz")) // something is gained
	})
	t.Run("conflict", func(t *testing.T) {
		memo := `{
"foo" : 0,
"bar" : 1
}`

		_, err := Merge(memo, map[string]int{"bar": 2, "fizz": 3})
		require.True(t, errorsmod.IsOf(err, sdkerrors.ErrConflict))
	})
}
