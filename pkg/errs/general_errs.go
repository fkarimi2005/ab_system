package errs

import "errors"

var (
	ErrRecordNotFound   = errors.New("запись не найдена в базе данных")
	ErrPermissionDenied = errors.New("недостаточно прав для выполнения данной операции")
	ErrIdIsInvalid      = errors.New("неверный UUID")
	ErrDeleteFailed     = errors.New("не удалось удалить запись из базы данных")
	ErrCantParseToTime  = errors.New("cannot parse")
)
