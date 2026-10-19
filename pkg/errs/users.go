package errs

import "errors"

var (
	ErrEmailIsEmpty          = errors.New("email не может быть пустым")
	ErrNameIsEmpty           = errors.New("name не может быть пустым")
	ErrEmailUniquenessFailed = errors.New("пользователь с таким email уже зарегистрирован в системе")
	ErrUserNotFound          = errors.New("пользователь не найден")
	ErrIdIsEmpty             = errors.New("id не может быть пустым")
	ErrRoleNotFound          = errors.New("role не найдено")
)
