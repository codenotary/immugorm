package immudb

import "errors"

var (
	ErrConstraintsNotImplemented = errors.New("constraints not implemented")
	ErrNotImplemented            = errors.New("not implemented")
	ErrCorruptedData             = errors.New("corrupted data")
	ErrTimeTravelNotAvailable    = errors.New("time travel is not available if verify flag is provided. This will change soon")
)
