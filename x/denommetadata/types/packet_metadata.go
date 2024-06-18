package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// MemoData represents the structure of the memo with user and hub metadata
type MemoData struct {
	TransferInject *TransferInject `json:"transferinject,omitempty"`
}

type TransferInject struct {
	DenomMetadata *types.Metadata `json:"denom_metadata,omitempty"`
}

func (p TransferInject) ValidateBasic() error {
	return p.DenomMetadata.Validate()
}

func ParsePacketMetadata(input string) *TransferInject {
	bz := []byte(input)
	var memo MemoData
	_ = json.Unmarshal(bz, &memo) // we don't care about the error
	return memo.TransferInject
}
