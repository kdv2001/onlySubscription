package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	uuid2 "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainOrder "github.com/kdv2001/onlySubscription/internal/domain/order"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
	"github.com/kdv2001/onlySubscription/internal/domain/subscription"
	"github.com/kdv2001/onlySubscription/internal/domain/user"
	custom_errors "github.com/kdv2001/onlySubscription/pkg/errors"
)

var productsTable = `create table if not exists products (
    id                  uuid primary key,
    user_id             uuid NOT NULL,
    order_id            uuid NOT NULL unique,
    description         text                        NOT NULL,
    created_at          timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at          timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    state               text                        NOT NULL,
	deadline          	timestamp WITHOUT TIME ZONE NOT NULL
);`

type Implementation struct {
	conn *pgxpool.Pool
}

var tables = []string{productsTable}

func NewImplementation(ctx context.Context, conn *pgxpool.Pool) (*Implementation, error) {
	for _, t := range tables {
		_, err := conn.Exec(ctx, t)
		if err != nil {
			return nil, err
		}
	}

	return &Implementation{
		conn: conn,
	}, nil
}

func (i *Implementation) CreateSubscription(ctx context.Context, req subscription.Subscription) (domainProducts.ID, error) {
	uuid := uuid2.New()
	_, err := i.conn.Exec(ctx, `insert into subscription
    (id, user_id, order_id, description, state, deadline)
	values ($1, $2, $3, $4, $5, $6)`,
		uuid.String(),
		req.UserID.String(),
		req.OrderID.String(),
		req.Description,
		req.State.String(),
		req.Deadline)
	if err != nil {
		return domainProducts.ID{}, err
	}

	return domainProducts.NewID(uuid.String()), nil
}

type subscriptionModel struct {
	ID          sql.NullString
	UserID      sql.NullString
	OrderID     sql.NullString
	Description sql.NullString
	CreatedAt   sql.NullTime
	UpdatedAt   sql.NullTime
	State       sql.NullString
	Deadline    sql.NullTime
}

func (i *Implementation) GetSubscriptions(ctx context.Context, r subscription.RequestList) ([]subscription.Subscription, error) {
	query := `select id, user_id, order_id, description, created_at, updated_at, state, deadline from subscription`
	values := []any{}

	if r.Filters != nil {
		query += " where"
		if r.Filters.Statuses != nil {
			query += " state in ("
			for j, s := range r.Filters.Statuses {
				if j != 0 {
					query += ", "
				}
				values = append(values, s.String())
				query += `$` + fmt.Sprint(len(values))
			}
			query += ")"
		}

		if r.Filters.Deadline != nil {
			if !r.Filters.Deadline.From.IsZero() {
				values = append(values, r.Filters.Deadline.From.UTC())
				query += ` and deadline > $` +
					fmt.Sprint(len(values))
			}
			if !r.Filters.Deadline.To.IsZero() {
				values = append(values, r.Filters.Deadline.To.UTC())
				query += ` and deadline < $` +
					fmt.Sprint(len(values))
			}
		}
	}

	if r.Pagination != nil {
		if r.Pagination.Num != 0 {
			values = append(values, r.Pagination.Num)
			query += ` limit $` + fmt.Sprint(len(values))
		}
	}

	res, err := i.conn.Query(ctx, query,
		values...)
	if err != nil {
		return nil, err
	}

	itemsResult := make([]subscription.Subscription, 0, 10)
	for res.Next() {
		o := subscriptionModel{}
		err = res.Scan(
			&o.ID,
			&o.UserID,
			&o.OrderID,
			&o.Description,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.State,
			&o.Deadline,
		)
		if err != nil {
			return nil, custom_errors.NewInternalError(err)
		}

		itemsResult = append(itemsResult, subscription.Subscription{
			ID:          subscription.NewID(o.ID.String),
			UserID:      user.NewID(o.UserID.String),
			OrderID:     domainOrder.New(o.OrderID.String),
			State:       subscription.NewState(o.State.String),
			Deadline:    o.Deadline.Time,
			CreatedAt:   o.CreatedAt.Time,
			UpdatedAt:   o.UpdatedAt.Time,
			Description: o.Description.String,
		})
	}

	return itemsResult, nil
}

// ChangeStatus изменяет статус
func (i *Implementation) ChangeStatus(ctx context.Context,
	subID subscription.ID,
	changeState subscription.ChangeState,
) error {
	return i.transaction(ctx, func(tx pgx.Tx) error {
		id := sql.NullString{}
		err := tx.QueryRow(ctx, `select id from subscription where uuid_eq(id, $1)
                        	limit 1 for update;`, subID).
			Scan(&id)
		if err != nil {
			return custom_errors.NewInternalError(err)
		}

		t, err := tx.Exec(ctx, `update subscription set state = $1, updated_at = NOW() AT TIME ZONE 'UTC'
                 where uuid_eq(id, $2) and state = $3;`,
			changeState.To, subID, changeState.From)
		if err != nil {
			return custom_errors.NewInternalError(err)
		}

		if t.RowsAffected() == 0 {
			return custom_errors.NewBadRequestError(errors.New("booking is expired"))
		}

		return nil
	})
}

func (i *Implementation) transaction(ctx context.Context, fnc func(tx pgx.Tx) error) error {
	tx, err := i.conn.Begin(ctx)
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
