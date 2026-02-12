package subscription

import (
	"errors"
	"fmt"
	"time"

	domainOrder "github.com/kdv2001/onlySubscription/internal/domain/order"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	"github.com/kdv2001/onlySubscription/internal/domain/user"
	custom_errors "github.com/kdv2001/onlySubscription/pkg/errors"
)

// ID айди
type ID struct {
	ID string
}

// String строковое представление
func (id ID) String() string {
	return id.ID
}

// NewID создает объект ID
func NewID[T string | int | int64](i T) ID {
	return ID{
		ID: fmt.Sprint(i),
	}
}

// State состояние подписки
type State string

const (
	// UnknownState неизвестное состояние
	UnknownState State = ""
	// ActiveState статус активен
	ActiveState State = "active"
	// InactiveState статус не активен
	InactiveState State = "inactive"
)

// String строковое представление
func (s State) String() string {
	return string(s)
}

// NewState состояние из строки
func NewState(s string) State {
	switch s {
	case string(UnknownState):
		return UnknownState
	case string(ActiveState):
		return ActiveState
	case string(InactiveState):
		return InactiveState
	}

	return UnknownState
}

// Subscription подписка
type Subscription struct {
	// ID подписки
	ID ID
	// UserID пользователя
	UserID user.ID
	// OrderID ID заказа
	OrderID domainOrder.ID
	// State состояние подписки
	State State
	// Deadline время окончания подписки
	Deadline time.Time
	// CreatedAt время создания
	CreatedAt time.Time
	// UpdatedAt время последнего обновления
	UpdatedAt time.Time
	// Description описание
	Description string
}

// Filters фильтры
type Filters struct {
	// Statuses фильтр по статусу
	Statuses []State
	// Deadline фильтр по окончанию подписки
	Deadline *primitives.IntervalFilter[time.Time]
}

// RequestList список параметров запроса
type RequestList struct {
	// Pagination пагинация
	Pagination *primitives.Pagination
	// Filters фильтры
	Filters *Filters
}

// ChangeState изменяет статус
type ChangeState struct {
	// From исходное состояние
	From State
	// To конечное состояние
	To State
}

// CanChangeStatus возвращает признак возможности перехода по статусу
func (s State) CanChangeStatus(toStatus State) bool {
	if s == toStatus {
		return true
	}

	switch s {
	case ActiveState:
		switch toStatus {
		case InactiveState:
			return true
		}
	}

	return false
}

// ErrStatusIsEqual ошибка эквивалентсности статусов
var ErrStatusIsEqual = errors.New("state is equal")

// NewChangeItemStatus создает структуру перехода статуса
func NewChangeItemStatus(from, to State) (ChangeState, error) {
	if from == to {
		return ChangeState{}, custom_errors.NewBadRequestError(ErrStatusIsEqual)
	}

	canChange := from.CanChangeStatus(to)
	if !canChange {
		return ChangeState{},
			custom_errors.NewBadRequestError(errors.New("can not change status"))
	}

	return ChangeState{
		From: from,
		To:   to,
	}, nil
}
