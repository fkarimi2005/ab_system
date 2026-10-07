package errs

import "errors"

var (
	ErrExperimentNameIsEmpty = errors.New("experiment name is empty")

	ErrExperimentNotFound      = errors.New("experiment not found")
	ErrExperimentVariantsEmpty = errors.New("experiment variants is empty")
	ErrWeightsSumMismatch      = errors.New("experiment weights sum mismatch")
	ErrControlVariantCount     = errors.New("experiment must have exactly one control variant")
	ErrCommentRequired         = errors.New("comment is required")
	ErrActiveExperimentExists  = errors.New("на этом флаге уже есть запущенный или приостановленный эксперимент")
	ErrVersionNotFound         = errors.New("версия эксперимента не найдено")
)
