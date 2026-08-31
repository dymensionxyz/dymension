// Package errors contains error helpers that preserve Dymension's protocol metadata.
package errors

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
)

// Join combines a registered Cosmos error with its underlying cause while
// retaining both Go's multi-error chain and the registered ABCI classification.
func Join(registered, cause error) error {
	joined := errors.Join(registered, cause)
	codespace, code, _ := errorsmod.ABCIInfo(registered, false)
	return joinedError{
		error:     joined,
		causes:    []error{registered, cause},
		codespace: codespace,
		code:      code,
	}
}

type joinedError struct {
	error
	causes    []error
	codespace string
	code      uint32
}

func (e joinedError) Unwrap() []error {
	return e.causes
}

func (e joinedError) Codespace() string {
	return e.codespace
}

func (e joinedError) ABCICode() uint32 {
	return e.code
}
