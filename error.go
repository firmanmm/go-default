package godefault

import "errors"

var (
	ErrNilValue             = errors.New("Passed nil value")
	ErrUnsupportedValue     = errors.New("Unsupported Value")
	ErrUnsupportedOperation = errors.New("Unsupported operation")
)
