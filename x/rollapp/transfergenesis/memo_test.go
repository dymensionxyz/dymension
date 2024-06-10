package transfergenesis

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func Test_getExample(t *testing.T) {
	// t.Skip("Not a real test. You can use this during development to quickly generate example json")

	type M struct {
		Data memo `json:"genesis_transfer"`
	}

	m := memo{
		Denom: banktypes.Metadata{
			Description: "The native staking and governance token of the rollappevm_1234-1",
			DenomUnits:  make([]*banktypes.DenomUnit, 2),
			Base:        "arax",
			Display:     "rax",
			Name:        "rax",
			Symbol:      "rax",
		},
		TotalNumTransfers: 0,
		ThisTransferIx:    0,
	}
	m.Denom.DenomUnits[0] = &banktypes.DenomUnit{
		Denom: "arax",
	}
	m.Denom.DenomUnits[1] = &banktypes.DenomUnit{
		Denom:    "rax",
		Exponent: 18,
	}

	var raw M
	raw.Data = m

	bz, err := json.MarshalIndent(raw, "", "\t")
	require.NoError(t, err)
	t.Log(string(bz))
}

func TestGetMemo(t *testing.T) {
	t.Run("empty: returns not found", func(t *testing.T) {
	})
	t.Run("does not have namespace key: returns not found", func(t *testing.T) {
	})
	t.Run("malformed: returns malformed", func(t *testing.T) {
	})
	t.Run("happy path: returns data", func(t *testing.T) {
	})
}
