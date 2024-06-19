package types

import (
	"encoding/json"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type GenesisTransferMemo struct {
	Denom banktypes.Metadata `json:"denom"`
	// How many transfers in total will be sent in the transfer genesis period
	TotalNumTransfers uint64 `json:"total_num_transfers"`
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
