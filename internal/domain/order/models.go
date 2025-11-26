package order

import (
	"fmt"
	"time"

	"github.com/kdv2001/onlySubscription/internal/domain/price"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
	"github.com/kdv2001/onlySubscription/internal/domain/user"
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
	UnknownStatus  Status = ""
	Form           Status = "form"
	ExpectPayments Status = "expect_payment"
	Processing     Status = "processing"
	Performed      Status = "performed"
	Cancelled      Status = "cancelled"
)

func (s Status) String() string {
	return string(s)
}

func StatusFromString(str string) Status {
	switch str {
	case string(ExpectPayments):
		return ExpectPayments
	case string(Processing):
		return Processing
	case string(Performed):
		return Performed
	case string(Cancelled):
		return Cancelled
	}

	return UnknownStatus
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
}

type Product struct {
	ItemID    domainProducts.ItemID
	ProductID domainProducts.ID
}
