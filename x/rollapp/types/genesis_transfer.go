package types

import (
	"encoding/json"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type GenesisTransferMemo struct {
	Denom banktypes.Metadata `json:"denom"`
}

func (g GenesisTransferMemo) Valid() error {
	return g.Denom.Validate()
}

func (g GenesisTransferMemo) Namespaced() GenesisTransferMemoNamespaced {
	return GenesisTransferMemoNamespaced{g}
}

// GenesisTransferMemoNamespaced is a namespaced wrapper
type GenesisTransferMemoNamespaced struct {
	Data GenesisTransferMemo `json:"genesis_transfer"`
}

// MustString returns a human-readable json string - intended for tests.
func (g GenesisTransferMemoNamespaced) MustString() string {
	bz, err := json.MarshalIndent(g, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(bz)
}
