package eibc

import (
	fmt "fmt"

	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const ModuleName = "eibc"

// Params defines the parameters for the module.
type Params struct {
	EpochIdentifier string         `protobuf:"bytes,1,opt,name=epoch_identifier,json=epochIdentifier,proto3" json:"epoch_identifier,omitempty" yaml:"epoch_identifier"`
	TimeoutFee      math.LegacyDec `protobuf:"bytes,2,opt,name=timeout_fee,json=timeoutFee,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"timeout_fee" yaml:"timeout_fee"`
	ErrackFee       math.LegacyDec `protobuf:"bytes,3,opt,name=errack_fee,json=errackFee,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"errack_fee" yaml:"errack_fee"`
}

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// KeyEpochIdentifier is the key for the epoch identifier
	KeyEpochIdentifier = []byte("EpochIdentifier")
	// KeyTimeoutFee is the key for the timeout fee
	KeyTimeoutFee = []byte("TimeoutFee")
	// KeyErrAckFee is the key for the error acknowledgement fee
	KeyErrAckFee = []byte("ErrAckFee")
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEpochIdentifier, &p.EpochIdentifier, validateEpochIdentifier),
		paramtypes.NewParamSetPair(KeyTimeoutFee, &p.TimeoutFee, validateTimeoutFee),
		paramtypes.NewParamSetPair(KeyErrAckFee, &p.ErrackFee, validateErrAckFee),
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateEpochIdentifier(i any) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(v) == 0 {
		return fmt.Errorf("epoch identifier cannot be empty")
	}
	return nil
}

func validateTimeoutFee(i any) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNil() {
		return fmt.Errorf("invalid global pool params: %+v", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("negative fee: %s", v)
	}

	if v.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("fee too high: %s", v)
	}

	return nil
}

func validateErrAckFee(i any) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNil() {
		return fmt.Errorf("invalid global pool params: %+v", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("negative fee: %s", v)
	}

	if v.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("fee too high: %s", v)
	}

	return nil
}
