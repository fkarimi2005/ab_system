package errs

import "errors"

var (
	ErrDuplicateEntry    = errors.New("запись с такими данными уже существует")
	ErrInvalidField      = errors.New("указано недопустимое значение поля")
	ErrUnsupportedDriver = errors.New("используемый драйвер базы данных не поддерживается")
	ErrNotImplemented    = errors.New("функционал еще не реализован")
)
