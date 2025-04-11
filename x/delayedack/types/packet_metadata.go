package types

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/gogoproto/proto"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type Memo struct {
	EIBC *EIBCMemo `json:"eibc"`
}

type EIBCMemo struct {
	// mandatory
	Fee string `json:"fee"`
	// can be nil
	OnFulfillHook []byte `json:"on_fulfill,omitempty"` // TODO: would be better to rework this whole thing into a pb message
}

func MakeEIBCMemo() EIBCMemo {
	return EIBCMemo{Fee: "0"}
}

func (p Memo) ValidateBasic() error {
	return p.EIBC.ValidateBasic()
}

// TODO: avoid duplicate calls
func (e EIBCMemo) GetFulfillHook() (*eibctypes.OnFulfillHook, error) {
	if len(e.OnFulfillHook) == 0 {
		return nil, nil
	}
	var hook eibctypes.OnFulfillHook
	err := proto.Unmarshal(e.OnFulfillHook, &hook)
	if err != nil {
		return nil, fmt.Errorf("unmarshal on fulfill hook: %w", err)
	}
	if err := hook.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("validate on fulfill hook: %w", err)
	}
	return &hook, nil
}

func (e EIBCMemo) ValidateBasic() error {
	_, err := e.FeeInt()
	if err != nil {
		return fmt.Errorf("fee: %w", err)
	}
	if _, err := e.GetFulfillHook(); err != nil {
		return fmt.Errorf("get on fulfill hook: %w", err)
	}
	return nil
}

func (e EIBCMemo) FeeInt() (math.Int, error) {
	i, ok := math.NewIntFromString(e.Fee)
	if !ok || i.IsNegative() {
		return math.Int{}, ErrBadEIBCFee
	}
	return i, nil
}

const (
	memoObjectKeyEIBC = "eibc"
	memoObjectKeyPFM  = "forward" // not to be confused with dymension/x/forward
)

var (
	ErrMemoUnmarshal     = gerrc.ErrInvalidArgument.Wrap("unmarshal memo")
	ErrMemoClashPFM      = gerrc.ErrUnimplemented.Wrap("EIBC packet with PFM is currently not supported")
	ErrEIBCMemoUnmarshal = gerrc.ErrInvalidArgument.Wrap("unmarshal eibc metadata")
	ErrEIBCMemoEmpty     = gerrc.ErrNotFound.Wrap("memo eIBC field is missing")
)

func ParseMemo(input string) (*Memo, error) {
	bz := []byte(input)

	memo := make(map[string]any)
	err := json.Unmarshal(bz, &memo)
	if err != nil {
		return nil, ErrMemoUnmarshal
	}
	if memo[memoObjectKeyPFM] != nil {
		// Currently not supporting eibc with PFM: https://github.com/dymensionxyz/dymension/issues/599
		return nil, ErrMemoClashPFM
	}
	if memo[memoObjectKeyEIBC] == nil {
		return nil, ErrEIBCMemoEmpty
	}
	var metadata Memo
	err = json.Unmarshal(bz, &metadata)
	if err != nil {
		return nil, ErrEIBCMemoUnmarshal
	}
	return &metadata, nil
}
