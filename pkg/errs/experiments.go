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
	ErrVersionNotFound         = errors.New("версия эксперимента не найдена")
	ErrInvalidTargeting        = errors.New("некорректное правило таргетинга")
	ErrEventTypeNotFound       = errors.New("тип события не найден")
	ErrInvalidEventType        = errors.New("некорректный тип события")
	ErrExposureTypeExists      = errors.New("тип показа (exposure) уже существует")
	ErrMetricNotFound          = errors.New("метрика не найдена")
	ErrInvalidMetric           = errors.New("некорректная метрика")
	ErrInvalidExperimentConfig = errors.New("некорректная конфигурация метрик или guardrails эксперимента")
	ErrNoTargetMetric          = errors.New("у эксперимента должна быть хотя бы одна целевая метрика")
	ErrDecisionNotFound        = errors.New("решение не найдено")
	ErrInvalidApproverGroup    = errors.New("некорректная группа согласующих")
	ErrNotInApproverGroup      = errors.New("вы не входите в группу согласующих владельца эксперимента")
)
