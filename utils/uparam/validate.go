package uparam

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func Uint64(x any) (uint64, error) {
	v, ok := x.(uint64)
	if !ok {
		return 0, errorsmod.WithType(gerrc.ErrInvalidArgument, x)
	}
	return v, nil
}

func ValidateUint64(x any) error {
	_, err := Uint64(x)
	return err
}

func PositiveUint64(x any) (uint64, error) {
	y, err := Uint64(x)
	if err != nil {
		return 0, err
	}
	if y == 0 {
		return 0, gerrc.ErrOutOfRange.Wrap("expect positive")
	}
	return y, nil
}

func ValidatePositiveUint64(x any) error {
	_, err := PositiveUint64(x)
	return err
}

func Dec(x any) (sdk.Dec, error) {
	v, ok := x.(sdk.Dec)
	if !ok {
		return sdk.Dec{}, errorsmod.WithType(gerrc.ErrInvalidArgument, x)
	}
	return v, nil
}

func ValidateDec(x any) error {
	_, err := Dec(x)
	return err
}

func NonNegativeDec(x any) (sdk.Dec, error) {
	y, err := Dec(x)
	if err != nil {
		return sdk.Dec{}, err
	}
	if y.IsNil() {
		return sdk.Dec{}, gerrc.ErrInvalidArgument.Wrap("expect not nil")
	}
	if y.IsNegative() {
		return sdk.Dec{}, gerrc.ErrOutOfRange.Wrap("expect not negative")
	}
	return y, nil
}

func ValidateNonNegativeDec(x any) error {
	_, err := NonNegativeDec(x)
	return err
}

// ZeroToOneDec allows any value in [0,1]
func ZeroToOneDec(x any) (sdk.Dec, error) {
	y, err := NonNegativeDec(x)
	if err != nil {
		return sdk.Dec{}, err
	}
	if y.GT(sdk.OneDec()) {
		return sdk.Dec{}, gerrc.ErrOutOfRange.Wrap("expect less than or equal to one")
	}
	return y, nil
}

func ValidateZeroToOneDec(x any) error {
	_, err := ZeroToOneDec(x)
	return err
}

func Coin(x any) (sdk.Coin, error) {
	v, ok := x.(sdk.Coin)
	if !ok {
		return sdk.Coin{}, errorsmod.WithType(gerrc.ErrInvalidArgument, x)
	}
	return v, nil
}

func ValidateCoin(x any) error {
	_, err := Coin(x)
	return err
}
