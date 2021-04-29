package godefault

import "errors"

var (
	ErrNilValue             = errors.New("Passed nil value")
	ErrUnsupportedOperation = errors.New("Unsupported operation")
)
