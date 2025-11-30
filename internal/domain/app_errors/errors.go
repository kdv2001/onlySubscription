package app_errors

import "errors"

var (
	// ErrNothingChanged ошибка нечего изменять
	ErrNothingChanged = errors.New("nothing has changed")
)
