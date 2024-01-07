package types

import "errors"

var (
	ErrInvalidPreStateAppHash = errors.New("invalid pre state app hash")
	ErrInvalidAppHash         = errors.New("invalid app hash")
)
