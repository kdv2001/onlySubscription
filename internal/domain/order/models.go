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

type Status string

const (
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

func (s Status) String() string {
	return string(s)
}

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

type ChangeOrderStatus struct {
	From Status
	To   Status
}

var ErrStatusIsEqual = errors.New("state is equal")

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

type CreateOrder struct {
	UserID    user.ID
	ProductID domainProducts.ID
}

type Order struct {
	ID         ID
	TotalPrice price.Price
	Status     Status
	UserID     user.ID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Product    Product
	TTL        time.Time
}

func (o *Order) SetProduct(product Product) {
	o.Product = product
}

type Product struct {
	ItemID      domainProducts.ItemID
	ProductID   domainProducts.ID
	Title       string
	Description string
}

type Filters struct {
	Statuses []Status
	TTL      *primitives.IntervalFilter[time.Time]
	UserID   user.ID
}

type Sort struct {
	CreatedAt primitives.SortType
}

type RequestList struct {
	Pagination *primitives.Pagination
	Filters    *Filters
	Sort       *Sort
}
