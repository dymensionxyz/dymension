package types

import (
	"encoding/json"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type GenesisTransferMemo struct {
	Denom banktypes.Metadata `json:"denom"`
	// How many transfers in total will be sent in the transfer genesis period
	TotalNumTransfers uint64 `json:"total_num_transfers"`
	// Which transfer is this? If there are 5 transfers total, they will be numbered 0,1,2,3,4.
	ThisTransferIx uint64 `json:"this_transfer_ix"`
}

func (g GenesisTransferMemo) Namespaced() GenesisTransferMemoRaw {
	return GenesisTransferMemoRaw{g}
}

// GenesisTransferMemoRaw is a namespaced wrapper
type GenesisTransferMemoRaw struct {
	Data GenesisTransferMemo `json:"genesis_transfer"`
}

// MustString returns a human-readable json string - intended for tests.
func (g GenesisTransferMemoRaw) MustString() string {
	bz, err := json.MarshalIndent(g, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(bz)
}
