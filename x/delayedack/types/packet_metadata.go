package types

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/gogoproto/proto"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

type Memo struct {
	EIBC *EIBCMemo `json:"eibc"`
}

// TODO: would be better to rework this whole thing into a pb message
type EIBCMemo struct {
	// mandatory
	Fee string `json:"fee"`
	// can be nil
	OnCompletionHook []byte `json:"dym_on_completion,omitempty"`
}

func DefaultEIBCMemo() EIBCMemo {
	return EIBCMemo{Fee: "0"}
}

func MakeEIBCMemo(fee string, onComplete []byte) EIBCMemo {
	return EIBCMemo{Fee: fee, OnCompletionHook: onComplete}
}

func (p Memo) ValidateBasic() error {
	return p.EIBC.ValidateBasic()
}

// Returns a memo that can be passed to the rollapp for an ibc transfer to the hub
func CreateMemo(eibcFee string, onComplete []byte) string {

	eibcM := MakeEIBCMemo(eibcFee, onComplete)
	m := Memo{
		EIBC: &eibcM,
	}

	eibcJson, _ := json.Marshal(m)

	return string(eibcJson)
}

// TODO: avoid duplicate calls
func (e EIBCMemo) GetCompletionHook() (*commontypes.CompletionHookCall, error) {
	if len(e.OnCompletionHook) == 0 {
		return nil, nil
	}
	var hook commontypes.CompletionHookCall
	err := proto.Unmarshal(e.OnCompletionHook, &hook)
	if err != nil {
		return nil, fmt.Errorf("unmarshal on completion hook: %w", err)
	}
	if err := hook.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("validate on completion hook: %w", err)
	}
	return &hook, nil
}

func (e EIBCMemo) ValidateBasic() error {
	_, err := e.FeeInt()
	if err != nil {
		return fmt.Errorf("fee: %w", err)
	}
	if _, err := e.GetCompletionHook(); err != nil {
		return fmt.Errorf("get on completion hook: %w", err)
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
