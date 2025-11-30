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

// CanChangeStatus возвращает признак возможности перехода в статус
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

// Products список продуктов
type Products []Product

// Product продукт
type Product struct {
	// ID продукта
	ID ID
	// Type тип продукта
	Type Type
	// Name название
	Name string
	// Description описание
	Description string
	// Image изображение
	Image Image
	// CreatedAt время создания продукта
	CreatedAt time.Time
	// UpdatedAt последнее время обновления
	UpdatedAt time.Time
	// Price цена продукта
	Price price.Price
	// SubscriptionPeriod период подписки
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

// Item единица инвентаря
type Item struct {
	// ID единицы инвентаря
	ID ItemID
	// ProductID ID продукта
	ProductID ID
	// Status единицы инвентаря
	Status ItemStatus
	// CreatedAt время создания
	CreatedAt time.Time
	// UpdatedAt время последнего обновления
	UpdatedAt time.Time
	// Payload полезная нагрузка
	Payload string
}

// Image изображение
type Image struct {
	// URL изображения
	URL string
}

// ChangeItemStatus изменение статуса товара
type ChangeItemStatus struct {
	// From исходное состояние единицы инвентаря
	From ItemStatus
	// To конечное состояние единицы инвентаря
	To ItemStatus
}

// ErrStatusIsEqual ошибка эквивалентности статусов
var ErrStatusIsEqual = errors.New("state is equal")

// NewChangeItemStatus создает структуру перехода статуса
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

// Filters фильтры
type Filters struct {
	// UpdatedAt фильтр по дате создания
	UpdatedAt *primitives.IntervalFilter[time.Time]
	// Statuses фильтр по статусу
	Statuses []ItemStatus
	// ItemsExist флаг наличия товаров
	ItemsExist bool
	// ProductID фильтр по ID продукта
	ProductID ID
}

// RequestList список параметров запроса
type RequestList struct {
	// Pagination пагинация
	Pagination *primitives.Pagination
	// Filters фильтры
	Filters *Filters
}
