package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/kdv2001/onlySubscription/internal/domain/order"
	"github.com/kdv2001/onlySubscription/internal/domain/price"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
	"github.com/kdv2001/onlySubscription/internal/domain/user"
	custom_errors "github.com/kdv2001/onlySubscription/pkg/errors"
)

var orderTable = `create table if not exists orders
(
    id           uuid primary key,
    user_id      uuid                        NOT NULL,
    order_status varchar                     NOT NULL,
    created_at   timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at   timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    total_price  decimal                     not null,
    currency     varchar                     not null,
    product_id    uuid                        not null
)`

type Implementation struct {
	c *pgxpool.Pool
}

var tables = []string{
	orderTable,
}

// NewImplementation создает объект репо
func NewImplementation(ctx context.Context, c *pgxpool.Pool) (*Implementation, error) {
	for _, t := range tables {
		_, err := c.Exec(ctx, t)
		if err != nil {
			return nil, nil
		}
	}

	return &Implementation{
		c: c,
	}, nil
}

func (i *Implementation) CreateOrder(ctx context.Context, o order.Order) (order.ID, error) {
	uid := uuid.New()
	_, err := i.c.Exec(ctx, `insert into orders(
    id,          
 	user_id ,
    order_status,
    total_price ,
    currency    ,
    product_id) values($1,$2,$3,$4,$5,$6)`,
		uid.String(),
		o.UserID.String(),
		o.Status.String(),
		o.TotalPrice.Value,
		o.TotalPrice.Currency,
		o.Product.ItemID.String(),
	)
	if err != nil {
		return order.ID{}, custom_errors.NewInternalError(err)
	}

	return order.New(uid.String()), nil
}

type orderModel struct {
	ID         sql.NullString
	UserID     sql.NullString
	CreatedAt  sql.NullTime
	UpdatedAt  sql.NullTime
	TotalPrice decimal.NullDecimal
	Currency   sql.NullString
	Status     sql.NullString
	ItemID     sql.NullString
}

func (i *Implementation) GetOrder(ctx context.Context, oID order.ID, userID user.ID) (order.Order, error) {
	o := orderModel{}
	err := i.c.QueryRow(ctx, `select * from orders where uuid_eq(id, $1) and uuid_eq(user_id, $2)`,
		oID.String(),
		userID.String()).Scan(
		&o.ID,
		&o.UserID,
		&o.Status,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.TotalPrice,
		&o.Currency,
		&o.ItemID)
	if err != nil {
		return order.Order{}, custom_errors.NewInternalError(err)
	}

	return order.Order{
		ID: order.New(o.ID.String),
		TotalPrice: price.Price{
			Currency: price.CurrencyFromString(o.Currency.String),
			Value:    o.TotalPrice.Decimal,
		},
		Status:    order.StatusFromString(o.Status.String),
		UserID:    user.NewID(o.UserID.String),
		CreatedAt: o.CreatedAt.Time,
		UpdatedAt: o.UpdatedAt.Time,
		Product: order.Product{
			ItemID: domainProducts.NewItemID(o.ItemID.String),
		},
	}, nil
}

func (i *Implementation) UpdateOrderStatus(ctx context.Context, oID order.ID, userID user.ID,
	status order.Status) error {
	_, err := i.c.Exec(ctx, `update orders set order_status = $1,
    updated_at = now() where uuid_eq(id, $2) and uuid_eq(user_id, $3)`,
		status.String(),
		oID.String(),
		userID.String(),
	)
	if err != nil {
		return custom_errors.NewInternalError(err)
	}

	return nil
}
