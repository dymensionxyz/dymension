// Package errors contains error helpers that preserve Dymension's protocol metadata.
package errors

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
)

// Join combines a registered Cosmos error with its underlying cause while
// retaining both Go's multi-error chain and the registered ABCI classification.
// It deliberately renders %+v as the stable single-line message and does not
// attach a new stack trace; callers can still inspect the cause through Unwrap.
func Join(registered, cause error) error {
	return join(registered, cause, "")
}

// Joinf is Join with context formatted ahead of the cause. The cause is read
// through Error rather than fmt's Formatter path so Cosmos errors do not expose
// their source location in the resulting user-facing message.
func Joinf(registered, cause error, format string, args ...any) error {
	return join(registered, cause, fmt.Sprintf(format, args...))
}

func join(registered, cause error, context string) error {
	if cause == nil {
		return nil
	}

	message := cause.Error()
	if context != "" {
		message = context + ": " + message
	}
	message += ": " + registered.Error()

	codespace, code, _ := errorsmod.ABCIInfo(registered, false)
	return &joinedError{
		message:    message,
		registered: registered,
		cause:      cause,
		codespace:  codespace,
		code:       code,
	}
}

type joinedError struct {
	message    string
	registered error
	cause      error
	codespace  string
	code       uint32
}

func (e *joinedError) Error() string {
	return e.message
}

func (e *joinedError) Unwrap() []error {
	return []error{e.registered, e.cause}
}

func (e *joinedError) Codespace() string {
	return e.codespace
}

func (e *joinedError) ABCICode() uint32 {
	return e.code
}
