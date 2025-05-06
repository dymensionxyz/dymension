package eibc

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var ModuleName = "eibc"
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
		paramtypes.NewParamSetPair(KeyEpochIdentifier, &p.EpochIdentifier, nil),
		paramtypes.NewParamSetPair(KeyTimeoutFee, &p.TimeoutFee, nil),
		paramtypes.NewParamSetPair(KeyErrAckFee, &p.ErrackFee, nil),
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
