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

type State string

const (
	UnknownState  State = ""
	ActiveState   State = "active"
	InactiveState State = "inactive"
)

func (s State) String() string {
	return string(s)
}

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

type Subscription struct {
	ID          ID
	UserID      user.ID
	OrderID     domainOrder.ID
	State       State
	Deadline    time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Description string
}

type Filters struct {
	Statuses []State
	Deadline *primitives.IntervalFilter[time.Time]
}

type RequestList struct {
	Pagination *primitives.Pagination
	Filters    *Filters
}

// ChangeState изменяет статус
type ChangeState struct {
	From State
	To   State
}

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

var ErrStatusIsEqual = errors.New("state is equal")

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
