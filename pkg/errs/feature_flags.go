package errs

import "errors"

var (
	ErrKeyIsEmpty          = errors.New("key is empty")
	ErrValueTypeIsEmpty    = errors.New("value type is empty")
	ErrDefaultValueIsEmpty = errors.New("default value is empty")
	ErrFeatureFlagNotFound = errors.New("feature flag not found")
)
