package order

import (
	"errors"
	"fmt"
	"time"

	"github.com/kdv2001/onlySubscription/internal/domain/price"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
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

// Type тип продукта
type Type string

const (
	// UnspecifiedType тип продукта - неизвестный
	UnspecifiedType Type = ""
	// SubscriptionType тип продукта - подписка
	SubscriptionType Type = "subscription"
)

// String строковое представление типа продукта
func (t Type) String() string {
	return string(t)
}

// TypeFromString преобразует строку в тип продукта
func TypeFromString(s string) Type {
	switch s {
	case string(SubscriptionType):
		return SubscriptionType
	default:
		return UnspecifiedType
	}
}

// ItemStatus статус продукта
type ItemStatus string

const (
	// UnknownStatus неизвестный статус
	UnknownStatus ItemStatus = ""
	// SaleStatus продается
	SaleStatus ItemStatus = "sale"
	// PreReservedStatus ожидание
	PreReservedStatus ItemStatus = "preReserved"
	// ReservedStatus зарезервирован
	ReservedStatus ItemStatus = "reserved"
	// PerformedStatus оплачен
	PerformedStatus ItemStatus = "performed"
	// RealizedStatus статус реализован
	RealizedStatus ItemStatus = "realized"
)

func (s ItemStatus) CanChangeStatus(toStatus ItemStatus) bool {
	if s == toStatus {
		return true
	}

	switch s {
	case SaleStatus:
		switch toStatus {
		case PreReservedStatus:
			return true
		}
	case PreReservedStatus:
		switch toStatus {
		case ReservedStatus:
			return true
		case SaleStatus:
			return true
		}

		return false
	case ReservedStatus:
		switch toStatus {
		case PerformedStatus:
			return true
		case SaleStatus:
			return true
		}

		return false
	case PerformedStatus:
		switch toStatus {
		case RealizedStatus:
			return true
		}

		return false
	}

	return false
}

// String строковое представление
func (s ItemStatus) String() string {
	return string(s)
}

// ItemStatusFromString возвращает статус из строки
func ItemStatusFromString(status string) ItemStatus {
	switch status {
	case string(SaleStatus):
		return SaleStatus
	case string(PreReservedStatus):
		return PreReservedStatus
	case string(ReservedStatus):
		return ReservedStatus
	case string(PerformedStatus):
		return PerformedStatus
	case string(RealizedStatus):
		return RealizedStatus
	default:
		return UnknownStatus
	}
}

type Products []Product

// Product продукт
type Product struct {
	ID          ID
	Type        Type
	Name        string
	Description string
	Image       Image
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Price       price.Price

	SubscriptionPeriod time.Duration
}

// ItemID айди
type ItemID struct {
	ID string
}

// String строковое представление
func (id ItemID) String() string {
	return id.ID
}

// NewItemID создает объект ID
func NewItemID[T string | int | int64](i T) ItemID {
	return ItemID{
		ID: fmt.Sprint(i),
	}
}

// Inventory инвентарь
type Inventory []Item

type Item struct {
	ID        ItemID
	ProductID ID
	Status    ItemStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	Payload   string
}

// Image изображение
type Image struct {
	URL string
}

// ChangeItemStatus изменяет статус товара
type ChangeItemStatus struct {
	From ItemStatus
	To   ItemStatus
}

var ErrStatusIsEqual = errors.New("state is equal")

func NewChangeItemStatus(from, to ItemStatus) (ChangeItemStatus, error) {
	if from == to {
		return ChangeItemStatus{}, custom_errors.NewBadRequestError(ErrStatusIsEqual)
	}

	canChange := from.CanChangeStatus(to)
	if !canChange {
		return ChangeItemStatus{},
			custom_errors.NewBadRequestError(errors.New("can not change status"))
	}

	return ChangeItemStatus{
		From: from,
		To:   to,
	}, nil
}

type Filters struct {
	UpdatedAt *primitives.IntervalFilter[time.Time]
	Statuses  []ItemStatus
	// ItemsExist флаг наличия товаров
	ItemsExist bool
	ProductID  ID
}

type RequestList struct {
	Pagination *primitives.Pagination
	Filters    *Filters
}
