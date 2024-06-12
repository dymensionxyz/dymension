package types

import (
	"encoding/json"
	"fmt"

	"github.com/dymensionxyz/dymension/v3/utils/gerr"

	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// MemoData represents the structure of the memo with user and hub metadata
type MemoData struct {
	TransferInject *TransferInject `json:"transferinject"`
}

type TransferInject struct {
	DenomMetadata *types.Metadata `json:"denom_metadata"`
}

func (p TransferInject) ValidateBasic() error {
	return p.DenomMetadata.Validate()
}

const memoObjectKeyTransferInject = "transferinject"

var (
	ErrMemoTransferInjectEmpty         = fmt.Errorf("memo 'transferinject' is missing")
	ErrMemoTransferInjectAlreadyExists = errorsmod.Wrapf(errortypes.ErrUnauthorized, "'transferinject' already exists in memo")
)

func ParsePacketMetadata(input string) (*TransferInject, error) {
	bz := []byte(input)

	var memo MemoData
	if err := json.Unmarshal(bz, &memo); err != nil {
		return nil, err
	}

	if memo.TransferInject == nil {
		return nil, ErrMemoTransferInjectEmpty
	}

	if memo.TransferInject.DenomMetadata == nil {
		return nil, errorsmod.Wrap(gerr.ErrNotFound, "denom metadata")
	}

	return memo.TransferInject, nil
}

func MemoAlreadyHasPacketMetadata(memo string) bool {
	memoMap := make(map[string]any)
	err := json.Unmarshal([]byte(memo), &memoMap)
	if err != nil {
		return false
	}

	_, ok := memoMap[memoObjectKeyTransferInject]
	return ok
}

func AddDenomMetadataToMemo(memo string, denomMetadata types.Metadata) (string, error) {
	memoMap := make(map[string]any)
	// doesn't matter if there is an error, the memo can be empty
	_ = json.Unmarshal([]byte(memo), &memoMap)

	if _, ok := memoMap[memoObjectKeyTransferInject]; ok {
		return "", ErrMemoTransferInjectAlreadyExists
	}

	memoMap[memoObjectKeyTransferInject] = TransferInject{DenomMetadata: &denomMetadata}
	bz, err := json.Marshal(memoMap)
	if err != nil {
		return memo, err
	}

	return string(bz), nil
}
