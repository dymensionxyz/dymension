package types

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type PacketMetadata struct {
	DenomMetadata types.Metadata `json:"denom_metadata"`
}

func (p PacketMetadata) ValidateBasic() error {
	return p.DenomMetadata.Validate()
}

const memoObjectKeyDM = "denom_metadata"

var (
	ErrMemoUnmarshal          = fmt.Errorf("unmarshal memo")
	ErrDenomMetadataUnmarshal = fmt.Errorf("unmarshal denom metadata")
	ErrMemoDMEmpty            = fmt.Errorf("memo denom metadata field is missing")
)

func ParsePacketMetadata(input string) (*PacketMetadata, error) {
	bz := []byte(input)

	memo := make(map[string]any)

	err := json.Unmarshal(bz, &memo)
	if err != nil {
		return nil, ErrMemoUnmarshal
	}
	if memo[memoObjectKeyDM] == nil {
		return nil, ErrMemoDMEmpty
	}

	var metadata PacketMetadata
	err = json.Unmarshal(bz, &metadata)
	if err != nil {
		return nil, ErrDenomMetadataUnmarshal
	}

	return &metadata, nil
}

func AddDenomMetadataToMemo(memo string, metadata types.Metadata) (string, error) {
	memoMap := make(map[string]any)
	_ = json.Unmarshal([]byte(memo), &memoMap)
	// doesn't matter if there is an error, the memo can be empty

	if _, ok := memoMap[memoObjectKeyDM]; ok {
		return "", fmt.Errorf("denom metadata already exists in memo. probably an attack")
	}

	memoMap[memoObjectKeyDM] = metadata
	bz, err := json.Marshal(memoMap)
	if err != nil {
		return memo, err
	}

	return string(bz), nil
}
