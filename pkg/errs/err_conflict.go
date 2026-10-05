package errs

import "errors"

var (
	ErrConflict               = errors.New("конфликт версий, обновите данные")
	ErrExperimentNotEditable  = errors.New("эксперимент нельзя изменять в текущем статусе")
	ErrInvalidTransition      = errors.New("недопустимый переход статуса")
	ErrActiveExperimentExists = errors.New("для этого флага уже есть запущенный или приостановленный эксперимент")
	ErrSelfApproval           = errors.New("нельзя согласовывать собственный эксперимент")
	ErrNotInApproverGroup     = errors.New("вы не входите в группу согласующих владельца эксперимента")
	ErrCommentRequired        = errors.New("для этого решения нужен комментарий")
	ErrInvalidApproverGroup   = errors.New("некорректная группа согласующих")
	ErrVersionNotFound        = errors.New("версия эксперимента не найдена")
)
