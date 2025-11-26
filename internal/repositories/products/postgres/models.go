package postgres

import (
	"database/sql"

	"github.com/shopspring/decimal"
)

type product struct {
	ID           sql.NullString
	Name         sql.NullString
	Description  sql.NullString
	CreatedAt    sql.NullTime
	UpdatedAt    sql.NullTime
	Type         sql.NullString
	Status       sql.NullString
	RecordStatus sql.NullString
	Price        decimal.NullDecimal

	SubscriptionPeriod sql.NullString
}

type item struct {
	ID          sql.NullString
	ProductID   sql.NullString
	Status      sql.NullString
	CreatedAt   sql.NullTime
	UpdatedAt   sql.NullTime
	Description sql.NullString
}
