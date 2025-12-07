package postgress

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
	domainPayment "github.com/kdv2001/onlySubscription/internal/domain/payment"
	"github.com/kdv2001/onlySubscription/internal/domain/price"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	custom_errors "github.com/kdv2001/onlySubscription/pkg/errors"
)

type Implementation struct {
	conn *pgxpool.Pool
}

var paymentTable = `create table if not exists invoices
(
    id             uuid primary key            not null ,
    state          varchar                     NOT NULL,
    payment_method varchar                     NOT NULL,
    created_at     timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at     timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    order_id       uuid                        NOT NULL,
    amount         decimal                     NOT NULL,
    currency       varchar                     NOT NULL,
    provider_id    varchar
)`

var tables = []string{
	paymentTable,
}

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

type invoiceModel struct {
	ID            sql.NullString
	State         sql.NullString
	PaymentMethod sql.NullString
	CreatedAt     sql.NullTime
	UpdatedAt     sql.NullTime
	OrderID       sql.NullString
	Amount        decimal.NullDecimal
	Currency      sql.NullString
	ProviderID    sql.NullString
}

func (i *Implementation) GetInvoice(ctx context.Context, iID domainPayment.ID) (domainPayment.Invoice, error) {
	o := invoiceModel{}
	err := i.conn.QueryRow(ctx, `select * from invoices where uuid_eq(id, $1)`,
		iID.String()).Scan(
		&o.ID,
		&o.State,
		&o.PaymentMethod,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.OrderID,
		&o.Amount,
		&o.Currency,
		&o.ProviderID,
		&sql.NullString{})
	if err != nil {
		return domainPayment.Invoice{}, custom_errors.NewInternalError(err)
	}

	return domainPayment.Invoice{
		ID:        domainPayment.New(o.ID.String),
		OrderID:   order.New(o.OrderID.String),
		State:     domainPayment.StateFromString(o.State.String),
		CreatedAt: o.CreatedAt.Time,
		UpdatedAt: o.UpdatedAt.Time,
		Price: price.Price{
			Currency: price.CurrencyFromString(o.Currency.String),
			Value:    o.Amount.Decimal,
		},
		PaymentMethod: domainPayment.PaymentMethodFromString(o.PaymentMethod.String),
		ProviderID:    domainPayment.NewProviderID(o.ProviderID.String),
	}, nil
}

func (i *Implementation) CreateInvoice(
	ctx context.Context,
	invoice domainPayment.Invoice,
) (domainPayment.ID, error) {
	uid := uuid.New()
	_, err := i.conn.Exec(ctx, `insert into invoices(
                            id,
                            state,
                            order_id,
                            amount,
                            currency,
                     		payment_method) values($1,$2,$3,$4,$5,$6)`,
		uid.String(),
		invoice.State,
		invoice.OrderID.String(),
		invoice.Price.Value,
		invoice.Price.Currency.String(),
		invoice.PaymentMethod.String(),
	)
	if err != nil {
		return domainPayment.ID{}, custom_errors.NewInternalError(err)
	}

	return domainPayment.New(uid.String()), nil
}

func (i *Implementation) UpdateInvoice(ctx context.Context,
	id domainPayment.ID,
	changeState domainPayment.ChangeInvoice,
) error {
	return i.transaction(ctx, func(tx pgx.Tx) error {
		scannedID := sql.NullString{}
		err := tx.QueryRow(ctx, `select id from invoices where uuid_eq(id, $1) for update;`, id).
			Scan(&scannedID)
		if err != nil {
			return custom_errors.NewInternalError(err)
		}

		t, err := tx.Exec(ctx, `update invoices set state = $1,
                    provider_id = $2,
                    updated_at = NOW() AT TIME ZONE 'UTC'
                 where uuid_eq(id, $3) and state = $4;`,
			changeState.ChangeState.To,
			changeState.ProviderID,
			id,
			changeState.ChangeState.From)
		if err != nil {
			return custom_errors.NewInternalError(err)
		}

		if t.RowsAffected() == 0 {
			return custom_errors.NewBadRequestError(app_errors.ErrNothingChanged)
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

func (i *Implementation) GetProcessingInvoices(
	ctx context.Context,
	r domainPayment.RequestList,
) ([]domainPayment.Invoice, error) {
	query := `select * from invoices`
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

		if r.Filters.UpdatedAt != nil {
			if !r.Filters.UpdatedAt.From.IsZero() {
				if len(values) != 0 {
					query += ` and`
				}
				values = append(values, r.Filters.UpdatedAt.From.UTC())
				query += ` updated_at > $` +
					fmt.Sprint(len(values))
			}
			if !r.Filters.UpdatedAt.To.IsZero() {
				if len(values) != 0 {
					query += ` and`
				}
				values = append(values, r.Filters.UpdatedAt.To.UTC())
				query += ` updated_at < $` +
					fmt.Sprint(len(values))
			}
		}
	}

	if r.Sort != nil {
		if r.Sort.UpdateAt == primitives.Descending {
			query += ` order by updated_at desc`
		}
	}

	if r.Pagination != nil {
		if r.Pagination.Num != 0 {
			values = append(values, r.Pagination.Num)
			query += ` limit $` + fmt.Sprint(len(values))
		}
	}

	res, err := i.conn.Query(ctx, query, values...)
	if err != nil {
		return nil, err
	}

	itemsResult := make([]domainPayment.Invoice, 0)
	for res.Next() {
		o := invoiceModel{}
		err = res.Scan(
			&o.ID,
			&o.State,
			&o.PaymentMethod,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.OrderID,
			&o.Amount,
			&o.Currency,
			&o.ProviderID,
			&sql.NullString{},
		)
		if err != nil {
			return nil, custom_errors.NewInternalError(err)
		}

		itemsResult = append(itemsResult, domainPayment.Invoice{
			ID:        domainPayment.New(o.ID.String),
			OrderID:   order.New(o.OrderID.String),
			State:     domainPayment.StateFromString(o.State.String),
			CreatedAt: o.CreatedAt.Time,
			UpdatedAt: o.UpdatedAt.Time,
			Price: price.Price{
				Currency: price.CurrencyFromString(o.Currency.String),
				Value:    o.Amount.Decimal,
			},
			PaymentMethod: domainPayment.PaymentMethodFromString(o.PaymentMethod.String),
			ProviderID:    domainPayment.NewProviderID(o.ProviderID.String),
		})
	}

	return itemsResult, nil
}
