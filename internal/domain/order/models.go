package order

import (
	"errors"
	"fmt"
	"time"

	"github.com/kdv2001/onlySubscription/internal/domain/price"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
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

// New создает объект ID
func New[T string | int | int64](i T) ID {
	return ID{
		ID: fmt.Sprint(i),
	}
}

// Status статус заказа.
type Status string

const (
	// UnknownStatus неизвестный статус заказа
	UnknownStatus Status = ""
	// Form заказ сформирован
	Form Status = "form"
	// ExpectPayments заказ ожидает оплаты
	ExpectPayments Status = "expect_payment"
	// Handling заказ обрабатывается
	Handling Status = "handling"
	// Processing заказ выполняется
	Processing Status = "processing"
	// Performed исполнен
	Performed Status = "performed"
	// Cancelled отменен
	Cancelled Status = "cancelled"
)

// String строковое представление
func (s Status) String() string {
	return string(s)
}

// CanChangeStatus возвращает признак возможности перехода статуса
func (s Status) CanChangeStatus(toStatus Status) bool {
	if s == toStatus {
		return true
	}

	switch s {
	case Form:
		switch toStatus {
		case ExpectPayments:
			return true
		case Cancelled:
			return true
		}
	case ExpectPayments:
		switch toStatus {
		case Handling:
			return true
		case Cancelled:
			return true
		}

		return false
	case Handling:
		switch toStatus {
		case Processing:
			return true
		case Cancelled:
			return true
		}

		return false
	case Processing:
		switch toStatus {
		case Performed:
			return true
		}

		return false
	}

	return false
}

// StatusFromString создает статус из строки
func StatusFromString(str string) Status {
	switch str {
	case string(Handling):
		return Handling
	case string(ExpectPayments):
		return ExpectPayments
	case string(Processing):
		return Processing
	case string(Performed):
		return Performed
	case string(Cancelled):
		return Cancelled
	case string(Form):
		return Form
	}

	return UnknownStatus
}

// ChangeOrderStatus структура для перехода статуса
type ChangeOrderStatus struct {
	// From исходный статус
	From Status
	// To конечный статус
	To Status
}

// ErrStatusIsEqual статусы эквивалентны
var ErrStatusIsEqual = errors.New("state is equal")

// NewChangeOrderStatus создает структуру перехода статуса
func NewChangeOrderStatus(from, to Status) (ChangeOrderStatus, error) {
	if from == to {
		return ChangeOrderStatus{}, custom_errors.NewBadRequestError(ErrStatusIsEqual)
	}

	canChange := from.CanChangeStatus(to)
	if !canChange {
		return ChangeOrderStatus{},
			custom_errors.NewBadRequestError(errors.New("can not change status"))
	}

	return ChangeOrderStatus{
		From: from,
		To:   to,
	}, nil
}

// CreateOrder структура для создания заказа
type CreateOrder struct {
	// UserID ID пользователя
	UserID user.ID
	// ProductID ID продукта с витрины
	ProductID domainProducts.ID
}

// Order заказ
type Order struct {
	// ID заказа
	ID ID
	// TotalPrice цена заказа
	TotalPrice price.Price
	// Status состояние заказа
	Status Status
	// UserID ID пользователя
	UserID user.ID
	// CreatedAt время создания заказа
	CreatedAt time.Time
	// UpdatedAt время последнего обновления заказа
	UpdatedAt time.Time
	// Product продукт продаваемый в заказе
	Product Product
	// TTL время жизни заказа
	TTL time.Time
}

// SetProduct устанавливает продукт в заказе
func (o *Order) SetProduct(product Product) {
	o.Product = product
}

// Product продукт заказа
type Product struct {
	// ItemID ID item из инвентаря
	ItemID domainProducts.ItemID
	// ProductID ID продукта с ветрины
	ProductID domainProducts.ID
	// Title название продукта
	Title string
	// Description описание
	Description string
}

// Filters фильтры для заказа
type Filters struct {
	// Statuses фильтр по статусу
	Statuses []Status
	// TTL фильтр по времени жизни
	TTL *primitives.IntervalFilter[time.Time]
	// UserID фильтр по ID пользователя
	UserID user.ID
}

// Sort сортировки заказ
type Sort struct {
	// CreatedAt сортировка по времени создания
	CreatedAt primitives.SortType
}

// RequestList список параметров запроса
type RequestList struct {
	// Pagination пагинация
	Pagination *primitives.Pagination
	// Filters фильтры
	Filters *Filters
	// Sort сортировка
	Sort *Sort
}
