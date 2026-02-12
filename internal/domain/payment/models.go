package payment

import (
	"errors"
	"fmt"
	"time"

	"github.com/kdv2001/onlySubscription/internal/domain/order"
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

// PaymentMethod метод оплаты
type PaymentMethod string

// String строковое представление
func (p PaymentMethod) String() string {
	return string(p)
}

const (
	// UnknownPaymentMethod неизвестный метод оплаты
	UnknownPaymentMethod PaymentMethod = ""
	// TelegramPaymentMethod telegram метод оплаты
	TelegramPaymentMethod PaymentMethod = "telegram"
)

// PaymentMethodFromString создает метод оплаты из строки
func PaymentMethodFromString(str string) PaymentMethod {
	switch str {
	case string(TelegramPaymentMethod):
		return TelegramPaymentMethod
	}

	return UnknownPaymentMethod
}

// State состояние платежа
type State string

// String строковое представление
func (s State) String() string {
	return string(s)
}

// StateFromString создает состояние из строки
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

// IsEmpty возвращает признак пустого ID
func (id ProviderID) IsEmpty() bool {
	return id.id == ""
}

// NewProviderID создает объект ID
func NewProviderID[T string | int | int64](i T) ProviderID {
	return ProviderID{
		id: fmt.Sprint(i),
	}
}

// Invoice счет
type Invoice struct {
	// ID айди счета
	ID ID
	// OrderID ID заказа
	OrderID order.ID
	// State состояние заказа
	State State
	// CreatedAt время создания заказа
	CreatedAt time.Time
	// UpdatedAt последнее время обновления заказа
	UpdatedAt time.Time
	// Price стоимость счета
	Price price.Price
	// PaymentMethod метод оплаты
	PaymentMethod PaymentMethod
	// ProviderID ID айди провайдера
	ProviderID ProviderID
}

// ChangeState изменяет статус заказа
type ChangeState struct {
	// From начальное состояние счета
	From State
	// To конечное состояние счета
	To State
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
	// ProviderID ID провайдера
	ProviderID ProviderID
	// ChangeState изменение статуса счета
	ChangeState ChangeState
}

// CreateInvoice создание счета
type CreateInvoice struct {
	// OrderID айди заказа
	OrderID order.ID
	// UserID айди пользователя
	UserID user.ID
	// Price стоимость
	Price price.Price
	// PaymentMethod метод оплаты
	PaymentMethod PaymentMethod
	// TelegramData данные поставщика
	TelegramData TelegramData
}

// Position продукт
type Position struct {
	// Title наименование позиции
	Title string
	// Description описание
	Description string
	// Price цена
	Price price.Price
}

// ReleaseInvoice предоставление счета
type ReleaseInvoice struct {
	// ID счета
	ID ID
	// Price цена
	Price price.Price
	// PaymentMethod метод оплаты
	PaymentMethod PaymentMethod
	// TelegramData данные поставщика
	TelegramData TelegramData
	// Position позиция
	Position Position
}

// ChatID ID чата
type ChatID string

// NewChatID создает ID чата
func NewChatID[T comparable](s T) ChatID {
	return ChatID(fmt.Sprint(s))
}

// TelegramData идентификаторы поставщика
type TelegramData struct {
	// ChatID ID чата
	ChatID ChatID
}

// Filters фильтры
type Filters struct {
	Statuses  []State
	UpdatedAt *primitives.IntervalFilter[time.Time]
}

// Sort сортировка
type Sort struct {
	UpdateAt primitives.SortType
}

// RequestList список параметров запрос
type RequestList struct {
	Pagination *primitives.Pagination
	Filters    *Filters
	Sort       *Sort
}

// ProviderTransaction трансакция поставщика
type ProviderTransaction struct {
	// ProviderID ID поставщика
	ProviderID ProviderID
	// InternalID внутренний ID поставщика
	InternalID ID
	// Date дата
	Date time.Time
}
