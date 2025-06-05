package streamer

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	ModuleName = "streamer"
)

// Params defines the parameters for the streamer module
type Params struct {
	MaxIterationsPerBlock uint64 `json:"max_iterations_per_block" yaml:"max_iterations_per_block"`
}

var _ paramtypes.ParamSet = (*Params)(nil)

var KeyMaxIterationsPerBlock = []byte("MaxIterationsPerBlock")

const (
	DefaultMaxIterationsPerBlock uint64 = 100
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(maxIterationsPerBlock uint64) Params {
	return Params{
		MaxIterationsPerBlock: maxIterationsPerBlock,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultMaxIterationsPerBlock)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxIterationsPerBlock, &p.MaxIterationsPerBlock, validateMaxIterationsPerBlock),
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateMaxIterationsPerBlock(p.MaxIterationsPerBlock); err != nil {
		return err
	}
	return nil
}

func validateMaxIterationsPerBlock(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("max iterations per block cannot be zero")
	}
	return nil
}
