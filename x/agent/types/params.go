package types

import (
	"fmt"
)

const DefaultMaxActionBytes = 100_000

func DefaultParams() Params {
	return Params{
		MaxActionBytes: DefaultMaxActionBytes,
	}
}

func (p Params) Validate() error {
	if p.MaxActionBytes == 0 {
		return fmt.Errorf("max action bytes must be positive")
	}
	return nil
}
