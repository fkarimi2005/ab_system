package errs

import "errors"

var (
	ErrConflict              = errors.New("конфликт версий, обновите данные")
	ErrExperimentNotEditable = errors.New("эксперимент нельзя изменять в текущем статусе")
	ErrInvalidTransition     = errors.New("недопустимый переход статуса")
)
