package custom_errors

import (
	"errors"
	"fmt"
	"strings"
)

// GroupError ...
type GroupError string

func (g GroupError) Error() string {
	return string(g)
}

var (
	// ErrorPermissionDenied ...
	ErrorPermissionDenied = GroupError("permission denied")
	// ErrorUnauthorized ...
	ErrorUnauthorized = GroupError("unauthorized")
	// ErrorBadRequest недопустимые аргументы
	ErrorBadRequest = GroupError("invalid arguments")
	// ErrorNotFound не найдено
	ErrorNotFound = GroupError("not found")
	// ErrorAborted ....
	ErrorAborted  = GroupError("aborted")
	ErrorInternal = GroupError("aborted")
)

// CustomError ...
type CustomError struct {
	// errType константный тип для конкретизации ошибки в группе и для последующей обработки
	errType string
	// группа ошибки
	group GroupError
	// base базовая ошибка, по задумке должна оставаться в пределах сервера
	base error
	// description описание ошибки, по задумке - нужно для передачи на клиент
	description string
	// details дополнительная информация по ошибке, по задумке - доп. детали, должны оставаться внутри сервиса
	details []string
}

// NewBadRequestError ...
func NewBadRequestError(base error) *CustomError {
	return NewCustomError(base, ErrorBadRequest)
}

// NewNotFoundError ...
func NewNotFoundError(base error) *CustomError {
	return NewCustomError(base, ErrorNotFound)
}

// NewForbiddenError ...
func NewForbiddenError(base error) *CustomError {
	return NewCustomError(base, ErrorPermissionDenied)
}

// NewUnauthorizedError ...
func NewUnauthorizedError(base error) *CustomError {
	return NewCustomError(base, ErrorUnauthorized)
}

// NewAbortedError ...
func NewAbortedError(base error) *CustomError {
	return NewCustomError(base, ErrorAborted)
}

// NewInternalError ...
func NewInternalError(base error) *CustomError {
	return NewCustomError(base, ErrorInternal)
}

// NewCustomError ...
func NewCustomError(base error, group GroupError) *CustomError {
	return &CustomError{
		group: group,
		base:  base,
	}
}

// Error реализация интерфейса
func (c *CustomError) Error() string {
	msg := fmt.Sprintf("error group = %s", c.group)
	if len(c.details) != 0 {
		msg = fmt.Sprintf("%s; details: = %s", msg, strings.Join(c.details, ","))
	}
	if c.errType != "" {
		msg = fmt.Sprintf("%s; errTypeCode: = %s", msg, c.errType)
	}

	return msg
}

// Copy ...
func (c *CustomError) Copy() *CustomError {
	if c == nil {
		return nil
	}

	copyC := *c
	return &copyC
}

// AddDetails добавляет детали к ошибке
func (c *CustomError) AddDetails(detail ...string) *CustomError {
	copyC := c.Copy()
	if copyC == nil {
		return nil
	}

	copyC.details = append(copyC.details, detail...)
	return copyC
}

// SetErrType добавляет константный тип ошибки
func (c *CustomError) SetErrType(errType string) *CustomError {
	copyC := c.Copy()
	if copyC == nil {
		return nil
	}

	copyC.errType = errType
	return copyC
}

// SetDescription добавляет описание ошибки
func (c *CustomError) SetDescription(description string) *CustomError {
	copyC := c.Copy()
	if copyC == nil {
		return nil
	}

	copyC.description = description
	return copyC
}

// GetErrType ...
func (c *CustomError) GetErrType() string {
	if c == nil {
		return ""
	}

	return c.errType
}

// GetDetails ...
func (c *CustomError) GetDetails() []string {
	if c == nil {
		return nil
	}

	return c.details
}

// GetDescription ...
func (c *CustomError) GetDescription() string {
	if c == nil {
		return ""
	}

	return c.description
}

// CustomErrorFromError ...
func CustomErrorFromError(err error) *CustomError {
	if err == nil {
		return nil
	}

	var ce *CustomError
	if !errors.As(err, &ce) {
		return nil
	}

	return ce
}

// Is проверяет гр
func (c *CustomError) Is(err error) bool {
	if c == nil {
		return false
	}

	var cErr *CustomError
	ok := errors.As(err, &cErr)
	if !ok {
		return errors.Is(err, c.group)
	}

	return cErr.group == c.group && cErr.errType == c.errType
}

// TypeCodeIs проверяет соответствие типа кода ошибки
func (c *CustomError) TypeCodeIs(typeCode string) bool {
	if c == nil {
		return false
	}

	return c.errType == typeCode
}

// Unwrap ...
func (c *CustomError) Unwrap() error {
	if c == nil {
		return nil
	}

	return c.base
}

// FormatMsg формирует сообщение об ошибке
func (c *CustomError) FormatMsg() string {
	msg := fmt.Sprintf("code = %s", c.group)
	if c.base != nil {
		msg = fmt.Sprintf("%s; base error: = %s", msg, c.base.Error())
	}
	if len(c.details) != 0 {
		msg = fmt.Sprintf("%s; details: = %s", msg, strings.Join(c.details, ","))
	}
	if c.errType != "" {
		msg = fmt.Sprintf("%s; errTypeCode: = %s", msg, c.errType)
	}
	if c.description != "" {
		msg = fmt.Sprintf("%s; description: = %s", msg, c.description)
	}

	return msg
}
