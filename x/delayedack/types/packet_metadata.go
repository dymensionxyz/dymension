package types

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PacketMetadata struct {
	EIBC *EIBCMetadata `json:"eibc"`
}

type EIBCMetadata struct {
	Fee string `json:"fee"`
}

func (p PacketMetadata) ValidateBasic() error {
	return p.EIBC.ValidateBasic()
}

func (e EIBCMetadata) ValidateBasic() error {
	_, err := e.FeeInt()
	if err != nil {
		return fmt.Errorf("fee: %w", err)
	}
	return nil
}

func (e EIBCMetadata) FeeInt() (math.Int, error) {
	i, ok := sdk.NewIntFromString(e.Fee)
	if !ok || i.IsNegative() {
		return math.Int{}, ErrBadEIBCFee
	}
	return i, nil
}

const (
	memoObjectKeyEIBC = "eibc"
	memoObjectKeyPFM  = "forward"
)

var (
	ErrMemoUnmarshal         = fmt.Errorf("unmarshal memo")
	ErrEIBCMetadataUnmarshal = fmt.Errorf("unmarshal eibc metadata")
	ErrMemoHashPFMandEIBC    = fmt.Errorf("EIBC packet with PFM is currently not supported")
	ErrMemoEibcEmpty         = fmt.Errorf("memo eIBC field is missing")
)

func ParsePacketMetadata(input string) (*PacketMetadata, error) {
	bz := []byte(input)

	memo := make(map[string]any)
	err := json.Unmarshal(bz, &memo)
	if err != nil {
		return nil, ErrMemoUnmarshal
	}
	if memo[memoObjectKeyPFM] != nil {
		// Currently not supporting eibc with PFM: https://github.com/dymensionxyz/dymension/issues/599
		return nil, ErrMemoHashPFMandEIBC
	}
	if memo[memoObjectKeyEIBC] == nil {
		return nil, ErrMemoEibcEmpty
	}
	var metadata PacketMetadata
	err = json.Unmarshal(bz, &metadata)
	if err != nil {
		return nil, ErrEIBCMetadataUnmarshal
	}
	return &metadata, nil
}
