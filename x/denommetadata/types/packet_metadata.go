package types

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// MemoData represents the structure of the memo with user and hub metadata
type MemoData struct {
	DenomMetadata *types.Metadata `json:"denom_metadata,omitempty"`
}

func (p MemoData) ValidateBasic() error {
	return p.DenomMetadata.Validate()
}

const memoObjectKeyDenomMetadata = "denom_metadata"

var ErrMemoDenomMetadataAlreadyExists = errorsmod.Wrapf(errortypes.ErrUnauthorized, "'denom_metadata' already exists in memo")

func ParsePacketMetadata(input string) *types.Metadata {
	bz := []byte(input)
	var memo MemoData
	_ = json.Unmarshal(bz, &memo) // we don't care about the error
	return memo.DenomMetadata
}

func MemoHasPacketMetadata(memo string) bool {
	memoMap := make(map[string]any)
	err := json.Unmarshal([]byte(memo), &memoMap)
	if err != nil {
		return false
	}

	_, ok := memoMap[memoObjectKeyDenomMetadata]
	return ok
}

func AddDenomMetadataToMemo(memo string, denomMetadata types.Metadata) (string, error) {
	memoMap := make(map[string]any)
	// doesn't matter if there is an error, the memo can be empty
	_ = json.Unmarshal([]byte(memo), &memoMap)

	if _, ok := memoMap[memoObjectKeyDenomMetadata]; ok {
		return "", ErrMemoDenomMetadataAlreadyExists
	}

	memoMap[memoObjectKeyDenomMetadata] = &denomMetadata
	bz, err := json.Marshal(memoMap)
	if err != nil {
		return memo, err
	}

	return string(bz), nil
}
