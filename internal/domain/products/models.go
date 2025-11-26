package order

import (
	"fmt"
	"time"

	"github.com/kdv2001/onlySubscription/internal/domain/price"
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
	case "subscription":
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
	// RefundStatus аннулирован
	RefundStatus ItemStatus = "refund"
)

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
	case string(RefundStatus):
		return RefundStatus
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
	ID          ItemID
	ProductID   ID
	Status      ItemStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Description string
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
