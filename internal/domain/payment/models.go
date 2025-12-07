package order

import (
	"errors"
	"fmt"
	"time"

	order "github.com/kdv2001/onlySubscription/internal/domain/order"
	"github.com/kdv2001/onlySubscription/internal/domain/price"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	"github.com/kdv2001/onlySubscription/internal/domain/user"
	custom_errors "github.com/kdv2001/onlySubscription/pkg/errors"
)

// ID айди
type ID struct {
	id string
}

// String строковое представление
func (id ID) String() string {
	return id.id
}

// New создает объект ID
func New[T string | int | int64](i T) ID {
	return ID{
		id: fmt.Sprint(i),
	}
}

type PaymentMethod string

func (p PaymentMethod) String() string {
	return string(p)
}

const (
	UnknownPaymentMethod  PaymentMethod = ""
	TelegramPaymentMethod PaymentMethod = "telegram"
)

func PaymentMethodFromString(str string) PaymentMethod {
	switch str {
	case string(TelegramPaymentMethod):
		return TelegramPaymentMethod
	}

	return UnknownPaymentMethod
}

type State string

func (s State) String() string {
	return string(s)
}

func StateFromString(str string) State {
	switch str {
	case string(ExpectPaymentState):
		return ExpectPaymentState
	case string(ProcessingState):
		return ProcessingState
	case string(PerformedState):
		return PerformedState
	case string(HandlingState):
		return HandlingState
	case string(CanceledState):
		return CanceledState
	}

	return UnknownState
}

const (
	// UnknownState не известный статус
	UnknownState State = ""
	// ExpectPaymentState ожидает оплаты, счет создан
	ExpectPaymentState State = "expect_payment"
	// HandlingState обслуживание, начато обслуживание платежа
	HandlingState State = "handling"
	// ProcessingState обработка платежа, платеж подтвержден у провайдера,
	// токены провайдера на нашей стороне сохранены
	ProcessingState State = "processing"
	// PerformedState платеж успешно исполнен
	PerformedState State = "performed"
	// CanceledState платеж отменен
	CanceledState State = "canceled"
	// RefundedState средства по счету возвращены
	RefundedState State = "refunded"
)

// CanChangeStatus проверяет можно ли изменить статус платежа
// реализует машину состояний для статусов платежей
func (s State) CanChangeStatus(toStatus State) bool {
	if s == toStatus {
		return true
	}

	switch s {
	case ExpectPaymentState:
		switch toStatus {
		case HandlingState:
			return true
		case CanceledState:
			return true
		}
	case HandlingState:
		switch toStatus {
		case ProcessingState:
			return true
		case CanceledState:
			return true
		}

		return false
	case ProcessingState:
		switch toStatus {
		case PerformedState:
			return true
		case CanceledState:
			return true
		}

		return false
	case PerformedState:
		switch toStatus {
		case RefundedState:
			return true
		}

		return false
	}

	return false
}

// ProviderID айди провайдера
type ProviderID struct {
	id string
}

// String строковое представление
func (id ProviderID) String() string {
	return id.id
}

func (id ProviderID) IsEmpty() bool {
	return id.id == ""
}

// NewProviderID создает объект ID
func NewProviderID[T string | int | int64](i T) ProviderID {
	return ProviderID{
		id: fmt.Sprint(i),
	}
}

// TransactionalProviderID айди транзакции провайдера
type TransactionalProviderID struct {
	id string
}

// String строковое представление
func (id TransactionalProviderID) String() string {
	return id.id
}

// NewTransactionalProviderID создает объект ID
func NewTransactionalProviderID[T string | int | int64](i T) TransactionalProviderID {
	return TransactionalProviderID{
		id: fmt.Sprint(i),
	}
}

// Invoice счет
type Invoice struct {
	ID            ID
	OrderID       order.ID
	State         State
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Price         price.Price
	PaymentMethod PaymentMethod
	ProviderID    ProviderID
}

// ChangeState изменяет статус заказа
type ChangeState struct {
	From State
	To   State
}

// ErrStatusIsEqual ошибка при попытке изменить статус на тот же самый
var ErrStatusIsEqual = errors.New("state is equal")

// NewChangeState создает объект для изменения статуса заказа
func NewChangeState(from, to State) (ChangeState, error) {
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

// ChangeInvoice изменяет статус товара
type ChangeInvoice struct {
	ProviderID  ProviderID
	ChangeState ChangeState
}

type ChatID string

func NewChatID[T comparable](s T) ChatID {
	return ChatID(fmt.Sprint(s))
}

type CreateInvoice struct {
	OrderID       order.ID
	UserID        user.ID
	Price         price.Price
	PaymentMethod PaymentMethod
	TelegramData  TelegramData
	Product       Product
}

type Product struct {
	Title       string
	Description string
	Price       price.Price
}

type ReleaseInvoice struct {
	ID            ID
	Price         price.Price
	PaymentMethod PaymentMethod
	TelegramData  TelegramData
	Product       Product
}

type TelegramData struct {
	ChatID ChatID
}

type Filters struct {
	Statuses  []State
	UpdatedAt *primitives.IntervalFilter[time.Time]
}

type Sort struct {
	UpdateAt primitives.SortType
}

type RequestList struct {
	Pagination *primitives.Pagination
	Filters    *Filters
	Sort       *Sort
}

type ProviderTransaction struct {
	ProviderID ProviderID
	InternalID ID
	Date       time.Time
}
