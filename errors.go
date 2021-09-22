package immudb

import "errors"

var (
	ErrConstraintsNotImplemented = errors.New("constraints not implemented")
	ErrNotImplemented            = errors.New("not implemented")
	ErrDeleteNotImplemented      = errors.New("delete is not possible on immudb")
	ErrCorruptedData             = errors.New("corrupted data")
)
