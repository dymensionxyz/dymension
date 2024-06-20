package transfergenesis

import (
	"testing"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	errorsmod "cosmossdk.io/errors"

	_ "embed"

	"github.com/stretchr/testify/require"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func Test_getExample(t *testing.T) {
	t.Skip("Not a real test. You can use this during development to quickly generate example json")

	m := rollapptypes.GenesisTransferMemo{
		Denom: banktypes.Metadata{
			Description: "The native staking and governance token of the rollappevm_1234-1",
			DenomUnits:  make([]*banktypes.DenomUnit, 2),
			Base:        "arax",
			Display:     "rax",
			Name:        "rax",
			Symbol:      "rax",
		},
	}
	m.Denom.DenomUnits[0] = &banktypes.DenomUnit{
		Denom: "arax",
	}
	m.Denom.DenomUnits[1] = &banktypes.DenomUnit{
		Denom:    "rax",
		Exponent: 18,
	}

	t.Log(m.Namespaced().MustString())
}

var (
	//go:embed testdata/memo_malformed.json
	memoMalformedBz string
	//go:embed testdata/memo_happy_path.json
	memoHappyPath string
	//go:embed testdata/memo_happy_path_with_other_namespace.json
	memoHappyPathWithOtherNamespace string
	//go:embed testdata/memo_namespace_empty.json
	memoNamespaceEmpty string
)

func TestGetMemo(t *testing.T) {
	t.Run("empty str: returns not found", func(t *testing.T) {
		_, err := getMemo("")
		require.True(t, errorsmod.IsOf(err, gerrc.ErrNotFound))
	})
	t.Run("empty json: returns not found", func(t *testing.T) {
		_, err := getMemo("{}")
		require.True(t, errorsmod.IsOf(err, gerrc.ErrNotFound))
	})
	t.Run("does not have namespace key: returns not found", func(t *testing.T) {
		_, err := getMemo(memoNamespaceEmpty)
		require.True(t, errorsmod.IsOf(err, gerrc.ErrNotFound))
	})
	t.Run("malformed: returns malformed", func(t *testing.T) {
		_, err := getMemo(memoMalformedBz)
		require.True(t, errorsmod.IsOf(err, gerrc.ErrInvalidArgument))
	})
	t.Run("happy path: returns data", func(t *testing.T) {
		m, err := getMemo(memoHappyPath)
		require.NoError(t, err)
		require.Equal(t, "arax", m.Denom.GetBase())
	})
	t.Run("happy path, with other namespace: returns data", func(t *testing.T) {
		m, err := getMemo(memoHappyPathWithOtherNamespace)
		require.NoError(t, err)
		require.Equal(t, "arax", m.Denom.GetBase())
	})
}
