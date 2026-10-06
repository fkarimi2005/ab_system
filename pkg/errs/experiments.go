package errs

import "errors"

var (
	ErrExperimentNameIsEmpty = errors.New("experiment name is empty")

	ErrExperimentNotFound      = errors.New("experiment not found")
	ErrExperimentVariantsEmpty = errors.New("experiment variants is empty")
	ErrWeightsSumMismatch      = errors.New("experiment weights sum mismatch")
	ErrControlVariantCount     = errors.New("experiment must have exactly one control variant")
	ErrCommentRequired         = errors.New("comment is required")
)
var (
	ErrExperimentVariantNotFound     = errors.New("experiment variant not found")
	ErrExperimentVariantWeightIsNil  = errors.New("experiment variant weight is nil")
	ErrExperimentVariantValueIsEmpty = errors.New("experiment variant value is empty")
	ErrExperimentVariantNameIsEmpty  = errors.New("experiment variant name is empty")
	ErrExperimentIDIsNil             = errors.New("experiment id is nil")
)
