package utest

import (
	errorsmod "cosmossdk.io/errors"
)

type Truer interface {
	True(value bool, msgAndArgs ...interface{})
}

func IsErr(t Truer, actual, expected error) {
	t.True(errorsmod.IsOf(actual, expected), `error is not an instance of expected: expected: %T, %s: got: %T, %s`, expected, expected, actual, actual)
}
