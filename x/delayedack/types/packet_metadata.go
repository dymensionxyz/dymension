package types

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/gogoproto/proto"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

type PacketMetadata struct {
	EIBC *EIBCMetadata `json:"eibc"`
}

type EIBCMetadata struct {
	Fee         string `json:"fee"`
	FulfillHook []byte `json:"fulfill_hook,omitempty"` // TODO: would be better to rework this whole thing into a pb message
}

func (p PacketMetadata) ValidateBasic() error {
	return p.EIBC.ValidateBasic()
}

// TODO: avoid duplicate calls
func (e EIBCMetadata) GetFulfillHook() (*eibctypes.FulfillHook, error) {
	if len(e.FulfillHook) == 0 {
		return nil, nil
	}
	var hook eibctypes.FulfillHook
	// unmarshal with protobuf
	err := proto.Unmarshal(e.FulfillHook, &hook)
	if err != nil {
		return nil, fmt.Errorf("unmarshal fulfill hook: %w", err)
	}
	if err := hook.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("validate fulfill hook: %w", err)
	}
	return &hook, nil
}

func (e EIBCMetadata) ValidateBasic() error {
	_, err := e.FeeInt()
	if err != nil {
		return fmt.Errorf("fee: %w", err)
	}
	if _, err := e.GetFulfillHook(); err != nil {
		return fmt.Errorf("fulfill hook: %w", err)
	}
	return nil
}

func (e EIBCMetadata) FeeInt() (math.Int, error) {
	i, ok := math.NewIntFromString(e.Fee)
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
