package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/kdv2001/onlySubscription/internal/domain/app_errors"
	"github.com/kdv2001/onlySubscription/internal/domain/order"
	"github.com/kdv2001/onlySubscription/internal/domain/price"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
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
    product_id   uuid                        not null,
	ttl  		 timestamp WITHOUT TIME ZONE NOT NULL 						 not null
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
    product_id,
    ttl) values($1,$2,$3,$4,$5,$6,$7)`,
		uid.String(),
		o.UserID.String(),
		o.Status.String(),
		o.TotalPrice.Value,
		o.TotalPrice.Currency,
		o.Product.ItemID.String(),
		o.TTL,
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
	TTL        sql.NullTime
}

func (i *Implementation) GetOrder(ctx context.Context, oID order.ID) (order.Order, error) {
	o := orderModel{}
	err := i.c.QueryRow(ctx, `select * from orders where uuid_eq(id, $1)`,
		oID.String()).Scan(
		&o.ID,
		&o.UserID,
		&o.Status,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.TotalPrice,
		&o.Currency,
		&o.ItemID,
		&o.TTL)
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

func (i *Implementation) UpdateOrderStatus(ctx context.Context, oID order.ID, changeState order.ChangeOrderStatus) error {
	return i.transaction(ctx, func(tx pgx.Tx) error {
		scannedID := sql.NullString{}
		err := tx.QueryRow(ctx, `select id from orders where uuid_eq(id, $1) for update;`, oID).
			Scan(&scannedID)
		if err != nil {
			return custom_errors.NewInternalError(err)
		}

		t, err := tx.Exec(ctx, `update orders set order_status = $1,
                    updated_at = now() AT TIME ZONE 'UTC'
                 where uuid_eq(id, $2) and order_status = $3;`,
			changeState.To,
			oID,
			changeState.From)
		if err != nil {
			return custom_errors.NewInternalError(err)
		}

		if t.RowsAffected() == 0 {
			return custom_errors.NewBadRequestError(app_errors.ErrNothingChanged)
		}

		return nil
	})
}

func (i *Implementation) GetOrders(ctx context.Context, r order.RequestList) ([]order.Order, error) {
	// TODO переделать на умный builder
	query := `select * from orders`
	values := []any{}

	if r.Filters != nil {
		query += " where"
		if r.Filters.Statuses != nil {
			query += " order_status in ("
			for j, s := range r.Filters.Statuses {
				if j != 0 {
					query += ", "
				}
				values = append(values, s.String())
				query += `$` + fmt.Sprint(len(values))
			}
			query += ")"
		}

		if r.Filters.TTL != nil {
			if !r.Filters.TTL.From.IsZero() {
				if len(values) != 0 {
					query += ` and`
				}
				values = append(values, r.Filters.TTL.From.UTC())
				query += ` ttl > $` +
					fmt.Sprint(len(values))
			}
			if !r.Filters.TTL.To.IsZero() {
				if len(values) != 0 {
					query += ` and`
				}
				values = append(values, r.Filters.TTL.To.UTC())
				query += ` ttl < $` +
					fmt.Sprint(len(values))
			}
		}

		if !r.Filters.UserID.IsEmpty() {
			if len(values) != 0 {
				query += ` and`
			}
			values = append(values, r.Filters.UserID.String())
			query += ` user_id = $` +
				fmt.Sprint(len(values))
		}
	}

	if r.Sort != nil {
		if r.Sort.CreatedAt == primitives.Descending {
			query += ` order by created_at desc`
		}
	}

	if r.Pagination != nil {
		if r.Pagination.Num != 0 {
			values = append(values, r.Pagination.Num)
			query += ` limit $` + fmt.Sprint(len(values))
		}
	}

	res, err := i.c.Query(ctx, query,
		values...)
	if err != nil {
		return nil, err
	}

	itemsResult := make([]order.Order, 0)
	for res.Next() {
		o := orderModel{}
		err = res.Scan(
			&o.ID,
			&o.UserID,
			&o.Status,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.TotalPrice,
			&o.Currency,
			&o.ItemID,
			&o.TTL,
		)
		if err != nil {
			return nil, custom_errors.NewInternalError(err)
		}

		itemsResult = append(itemsResult, order.Order{
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
		})
	}

	return itemsResult, nil
}

func (i *Implementation) transaction(ctx context.Context, fnc func(tx pgx.Tx) error) error {
	tx, err := i.c.Begin(ctx)
	if err != nil {
		return custom_errors.NewInternalError(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	err = fnc(tx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return custom_errors.NewInternalError(err)
	}

	return nil
}
